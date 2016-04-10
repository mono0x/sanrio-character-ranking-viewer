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

	"gopkg.in/gorp.v1"

	"github.com/ChimeraCoder/anaconda"
	"github.com/robfig/cron"
)

const (
	SearchWord = "サンリオキャラクター大賞"
)

var (
	pattern        = regexp.MustCompile(`(?::\s+)?(.+?)を、「(?:.+?)」なでました！みんなもなでてみよう！`)
	spaceReplacer  = strings.NewReplacer("　", " ")
	normalizeTable = map[string]string{
		"ちんじゅうみん＆ ゴーちゃん。": "ちんじゅうみん＆ゴーちゃん。",
		"歯ぐるマンスタイル":       "歯ぐるまんすたいる",
	}
)

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
				if err := processStatus(context.dbMap, status); err != nil {
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

func processStatus(dbMap *gorp.DbMap, receivedStatus anaconda.Tweet) error {
	if receivedStatus.Retweeted {
		return nil
	}

	var status Status
	if err := dbMap.SelectOne(&status,
		`SELECT * FROM status WHERE id = $1`, receivedStatus.Id); err != nil {

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

	entry, err := FindEntryByName(dbMap, 2, name)
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
