package main

type Crawler struct{}

func (c *Crawler) Help() string {
	return `sanrio-character-ranking-viewer crawler`
}

func (c *Crawler) Run(args []string) int {
	return 0
}

func (c *Crawler) Synopsis() string {
	return `Start crawler`
}
