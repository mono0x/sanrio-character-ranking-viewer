package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/braintree/manners"
	"github.com/elithrar/goji-logger"
	"github.com/lestrrat/go-server-starter/listener"
	"goji.io"
	"goji.io/pat"
)

type Controller struct {
	*AppContext
}

func NewController() (*Controller, error) {
	context, err := NewAppContext()
	if err != nil {
		return nil, err
	}

	return &Controller{
		AppContext: context,
	}, nil
}

func (c *Controller) Close() {
	c.AppContext.Close()
}

func (c *Controller) Serve() error {
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
		return err
	}

	var l net.Listener
	if len(listeners) > 0 {
		l = listeners[0]
	} else {
		l, err = net.Listen("tcp", ":14000")
		if err != nil {
			return err
		}
	}

	mux := goji.NewMux()
	mux.UseC(logger.RequestID)
	mux.UseC(logger.Logger)
	mux.Handle(pat.Get("/assets/*"), http.FileServer(AssetFileSystem{}))
	mux.Handle(pat.Get("/"), http.HandlerFunc(c.HandleIndex))

	manners.Serve(l, mux)
	return nil
}

func (c *Controller) HandleIndex(w http.ResponseWriter, r *http.Request) {
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
