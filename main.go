package main

import (
	"log"
	"net/http"
	"os"

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
		"crawler": func() (cli.Command, error) {
			return &Crawler{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}
