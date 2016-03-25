package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goji.io"
	"goji.io/pat"

	"github.com/braintree/manners"
	"github.com/lestrrat/go-server-starter/listener"
	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

//go:generate ego -package main templates/
//go:generate go-bindata -ignore .gitkeep assets/...

type Character struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Status struct {
	Id        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	Source    string    `db:"source"`
}

type Ranking struct {
	Id        int       `db:"id"`
	Name      string    `db:"name"`
	StartedOn time.Time `db:"started_on"`
	EndedOn   time.Time `db:"ended_on"`
}

type Vote struct {
	RankingId   int   `db:"ranking_id"`
	CharacterId int   `db:"character_id"`
	StatusId    int64 `db:"status_id"`
}

type RankingItem struct {
	Name  string
	Count int
}

type appContext struct {
	dbMap *gorp.DbMap
}

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
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Println("ok")

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
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = renderLayout(w, "Sanrio Character Ranking Viewer", func(w io.Writer) error {
		return renderIndex(w, ranking, rankingItems)
	})
}

func main() {
	context := &appContext{}

	db, err := sql.Open("postgres", "user=app dbname=sanrio-character-ranking sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbMap.AddTableWithName(Character{}, "character").SetKeys(true, "Id")
	dbMap.AddTableWithName(Status{}, "status").SetKeys(true, "Id")
	dbMap.AddTableWithName(Ranking{}, "ranking").SetKeys(true, "Id")
	dbMap.AddTableWithName(Vote{}, "vote")

	context.dbMap = dbMap

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGTERM)
	go func() {
		for {
			s := <-signalChan
			if s == syscall.SIGTERM {
				manners.Close()
			}
		}
	}()

	listeners, err := listener.ListenAll()
	if err != nil {
		log.Fatal(err)
	}

	var l net.Listener
	if len(listeners) > 0 {
		l = listeners[0]
	} else {
		l, err = net.Listen("tcp", ":14000")
		if err != nil {
			log.Fatal(err)
		}
	}

	mux := goji.NewMux()
	mux.Handle(pat.Get("/assets/*"), http.FileServer(AssetFileSystem{}))
	mux.Handle(pat.Get("/"), &appHandler{context, handleIndex})

	manners.Serve(l, mux)
}
