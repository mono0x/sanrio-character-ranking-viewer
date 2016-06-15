package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mitchellh/cli"
)

//go:generate ego -package main templates/
//go:generate go-bindata -ignore .gitkeep assets/...

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	_ = godotenv.Load()

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
		"register": func() (cli.Command, error) {
			return &Register{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}
