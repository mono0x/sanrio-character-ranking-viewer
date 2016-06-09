package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/ChimeraCoder/anaconda"
	"github.com/robfig/cron"
)

const (
	SearchWord = "サンリオキャラクター大賞"
)

var (
	spaceReplacer  = strings.NewReplacer("　", " ")
	normalizeTable = map[string]string{
		"ちんじゅうみん＆ ゴーちゃん。":            "ちんじゅうみん＆ゴーちゃん。",
		"歯ぐるマンスタイル":                  "歯ぐるまんすたいる",
		"中年ひろいん Ｏｊｉｓａｎ’ｓ":            "中年ひろいんＯｊｉｓａｎ’ｓ",
		"徒然なる操り夢幻庵（つれづれなるあやつりむげんあん）": "徒然なる操り霧幻庵（つれづれなるあやつりむげんあん）",
	}
)

var pattern *regexp.Regexp

type Crawler struct{}

func (c *Crawler) Help() string {
	return `sanrio-character-ranking-viewer crawler`
}

func (c *Crawler) Run(args []string) int {
	context, err := newContext()
	if err != nil {
		log.Fatal(err)
	}
	defer context.Close()

	ranking, err := GetCurrentRanking(context.dbMap, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	characters, err := GetEntryCharacters(context.dbMap, ranking.Id)
	parts := make([]string, len(characters)+len(normalizeTable))
	for _, character := range characters {
		parts = append(parts, regexp.QuoteMeta(character.Name))
	}
	for name, _ := range normalizeTable {
		parts = append(parts, regexp.QuoteMeta(name))
	}
	pattern = regexp.MustCompile(
		`(` + strings.Join(parts, `|`) + `)を、「(?:.+?)」なでました！みんなもなでてみよう！`)

	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))

	api := anaconda.NewTwitterApi(os.Getenv("TWITTER_OAUTH_TOKEN"), os.Getenv("TWITTER_OAUTH_TOKEN_SECRET"))

	stream := api.PublicStreamFilter(url.Values{
		"track": []string{SearchWord},
	})

	waitGroup := &sync.WaitGroup{}

	statusChan := make(chan anaconda.Tweet)

	processorQuitChan := make(chan struct{})
	streamingQuitChan := make(chan struct{})
	quitChans := []chan struct{}{
		processorQuitChan,
		streamingQuitChan,
	}

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()

		for {
			select {
			case status := <-statusChan:
				if err := processStatus(context.dbMap, status, ranking); err != nil {
					log.Print(err)
				}
				break
			case <-processorQuitChan:
				return
			}
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()

		for {
			select {
			case item := <-stream.C:
				switch status := item.(type) {
				case anaconda.Tweet:
					statusChan <- status
					break
				}
			case <-streamingQuitChan:
				return
			}
		}
	}()

	cron := cron.New()
	cron.AddFunc("0 */5 * * * *", func() {
		waitGroup.Add(1)
		defer waitGroup.Done()

		v := url.Values{}
		v.Set("count", "200")
		searchResult, err := api.GetSearch(SearchWord, v)
		if err != nil {
			log.Print(err)
			return
		}
		for _, status := range searchResult.Statuses {
			statusChan <- status
		}
	})
	cron.Start()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()

		for {
			s := <-signalChan
			log.Print(s)
			for _, quitChan := range quitChans {
				quitChan <- struct{}{}
			}
			break
		}
	}()

	waitGroup.Wait()
	return 0
}

func (c *Crawler) Synopsis() string {
	return `Start crawler`
}

func processStatus(dbMap *gorp.DbMap, receivedStatus anaconda.Tweet, ranking *Ranking) error {
	if receivedStatus.RetweetedStatus != nil {
		return nil
	}

	var status Status
	obj, err := dbMap.Get(Status{}, receivedStatus.Id)
	if err != nil {
		createdAt, err := receivedStatus.CreatedAtTime()
		if err != nil {
			return err
		}

		statusJson, err := json.Marshal(&receivedStatus)
		if err != nil {
			return err
		}
		status = Status{
			Id:        receivedStatus.Id,
			CreatedAt: createdAt,
			Source:    string(statusJson),
		}
		if err := dbMap.Insert(&status); err != nil {
			return err
		}
	} else {
		status = *obj.(*Status)
	}

	submatches := pattern.FindStringSubmatch(receivedStatus.Text)
	if submatches == nil {
		return nil
	}

	name := spaceReplacer.Replace(submatches[1])
	if normalizedName, ok := normalizeTable[name]; ok {
		name = normalizedName
	}
	log.Print(status.Id, name)

	entry, err := FindEntryByName(dbMap, ranking.Id, name)
	if err != nil {
		return err
	}

	vote := Vote{
		RankingId:   entry.RankingId,
		CharacterId: entry.CharacterId,
		StatusId:    status.Id,
	}
	if err := dbMap.Insert(&vote); err != nil {
		return err
	}
	return nil
}
