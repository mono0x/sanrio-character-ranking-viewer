package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mitchellh/cli"
)

//go:generate ego -package main templates/
//go:generate go-bindata -ignore .gitkeep assets/...

type appHandler struct {
	*appContext
	handler func(*appContext, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.handler(ah.appContext, w, r)
}

func handleIndex(c *appContext, w http.ResponseWriter, r *http.Request) {
	var ranking Ranking
	if err := c.dbMap.SelectOne(&ranking,
		`SELECT * FROM ranking WHERE started_on <= :today ORDER BY ended_on DESC, started_on DESC LIMIT 1`,
		map[string]interface{}{
			"today": time.Now(),
		}); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var rankingItems []RankingItem
	if _, err := c.dbMap.Select(&rankingItems, `
		SELECT
			character.name AS Name,
			x.count AS Count
		FROM (
			SELECT
				character_id, COUNT(*) AS count
			FROM vote
			WHERE vote.ranking_id = :ranking_id
			GROUP BY vote.character_id
		) x
		JOIN character ON character.id = x.character_id
		ORDER BY x.count DESC, character.name ASC
	`, map[string]interface{}{
		"ranking_id": ranking.Id,
	}); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = renderLayout(w, "Sanrio Character Ranking Viewer", func(w io.Writer) error {
		return renderIndex(w, ranking, rankingItems)
	})
}

func main() {
	c := cli.NewCLI("sanrio-character-ranking-viewer", "0.0.1")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"": func() (cli.Command, error) {
			return &Server{}, nil
		},
		"server": func() (cli.Command, error) {
			return &Server{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}
