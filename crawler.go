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
	SearchWord = "VoteSanrio"
)

var (
	pattern       = regexp.MustCompile(`サンリオキャラクター大賞で(.+?)に投票したよ`)
	spaceReplacer = strings.NewReplacer("　", " ")
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

func processStatus(dbMap *gorp.DbMap, status anaconda.Tweet) error {
	createdAt, err := status.CreatedAtTime()
	if err != nil {
		return err
	}

	count, err := dbMap.SelectInt(`SELECT COUNT(*) FROM status WHERE id = $1`, status.Id)
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}

	statusJson, err := json.Marshal(&status)
	if err != nil {
		return err
	}
	statusObject := Status{
		Id:        status.Id,
		CreatedAt: createdAt,
		Source:    string(statusJson),
	}
	if err := dbMap.Insert(&statusObject); err != nil {
		return err
	}

	submatches := pattern.FindStringSubmatch(status.Text)
	if submatches == nil {
		return nil
	}

	name := spaceReplacer.Replace(submatches[1])

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
