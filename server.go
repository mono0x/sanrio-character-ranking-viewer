package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goji.io"

	"github.com/braintree/manners"
	"github.com/elithrar/goji-logger"
	"github.com/lestrrat/go-server-starter/listener"
	"goji.io/pat"
)

type Server struct{}

func (s *Server) Help() string {
	return `sanrio-character-ranking-viewer server`
}

func (s *Server) Run(args []string) int {
	context, err := newContext()
	if err != nil {
		log.Fatal(err)
	}
	defer context.Close()

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
	mux.UseC(logger.RequestID)
	mux.UseC(logger.Logger)
	mux.Handle(pat.Get("/assets/*"), http.FileServer(AssetFileSystem{}))
	mux.Handle(pat.Get("/"), &appHandler{context, handleIndex})

	manners.Serve(l, mux)
	return 0
}

func (s *Server) Synopsis() string {
	return `Start server`
}

func handleIndex(c *appContext, w http.ResponseWriter, r *http.Request) {
	ranking, err := GetLatestRanking(c.dbMap, time.Now())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rankingItems, err := GetRankingItems(c.dbMap, ranking.Id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = renderLayout(w, "Sanrio Character Ranking Viewer", func(w io.Writer) error {
		return renderIndex(w, ranking, rankingItems)
	})
}
