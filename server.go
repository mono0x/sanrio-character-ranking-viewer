package main

import "log"

type Server struct{}

func (s *Server) Help() string {
	return `sanrio-character-ranking-viewer server`
}

func (s *Server) Run(args []string) int {
	controller, err := NewController()
	if err != nil {
		log.Fatal(err)
	}
	defer controller.Close()

	if err := controller.Serve(); err != nil {
		log.Fatal(err)
	}
	return 0
}

func (s *Server) Synopsis() string {
	return `Start server`
}
