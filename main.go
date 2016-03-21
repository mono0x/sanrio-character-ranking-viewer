package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goji.io"
	"goji.io/pat"

	"github.com/arschles/go-bindata-html-template"
	"github.com/braintree/manners"
	"github.com/lestrrat/go-server-starter/listener"
	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

//go:generate go-bindata assets/... tmpl/...

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

type appContext struct {
	dbMap     *gorp.DbMap
	templates map[string]*template.Template
}

type appHandler struct {
	*appContext
	handler func(*appContext, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.handler(ah.appContext, w, r)
}

func handleIndex(c *appContext, w http.ResponseWriter, r *http.Request) {
	type rankingView struct {
		Name  string
		Count int
	}

	var ranking []rankingView
	if _, err := c.dbMap.Select(&ranking, `
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
	`, map[string]string{
		"ranking_id": "1",
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.templates["index"].Execute(w, map[string]interface{}{
		"Ranking": &ranking,
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
	context.templates = make(map[string]*template.Template, 0)

	for _, name := range []string{
		"index",
	} {
		t, err := template.New("main", Asset).ParseFiles("tmpl/layout.html", "tmpl/"+name+".html")
		if err != nil {
			log.Fatal(err)
		}
		context.templates[name] = t
	}

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
