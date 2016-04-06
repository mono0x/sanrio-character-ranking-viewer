package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

type Crawler struct{}

func (c *Crawler) Help() string {
	return `sanrio-character-ranking-viewer crawler`
}

var (
	pattern = regexp.MustCompile(`サンリオキャラクター大賞で(.+?)に投票したよ`)
)

func (c *Crawler) Run(args []string) int {
	context, err := newContext()
	if err != nil {
		log.Fatal(err)
	}
	defer context.Close()

	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))

	api := anaconda.NewTwitterApi(os.Getenv("TWITTER_OAUTH_TOKEN"), os.Getenv("TWITTER_OAUTH_TOKEN_SECRET"))
	api.SetLogger(anaconda.BasicLogger)

	stream := api.PublicStreamFilter(url.Values{
		"track": []string{"VoteSanrio"},
	})

	replacer := strings.NewReplacer("　", " ")

	for {
		select {
		case item := <-stream.C:
			switch status := item.(type) {
			case anaconda.Tweet:
				createdAt, err := status.CreatedAtTime()
				if err != nil {
					log.Print(err)
					break
				}
				statusJson, err := json.Marshal(&status)
				if err != nil {
					log.Print(err)
					break
				}
				statusObject := Status{
					CreatedAt: createdAt,
					Source:    string(statusJson),
				}
				if err := context.dbMap.Insert(&statusObject); err != nil {
					log.Print(err)
					break
				}

				submatches := pattern.FindStringSubmatch(status.Text)
				if submatches == nil {
					break
				}

				name := replacer.Replace(submatches[1])

				entry, err := FindEntryByName(context.dbMap, 2, name)
				if err != nil {
					log.Print(err)
					break
				}

				vote := Vote{
					RankingId:   entry.RankingId,
					CharacterId: entry.CharacterId,
					StatusId:    statusObject.Id,
				}
				if err := context.dbMap.Insert(&vote); err != nil {
					log.Print(err)
					break
				}
			}
		}
	}

	return 0
}

func (c *Crawler) Synopsis() string {
	return `Start crawler`
}
