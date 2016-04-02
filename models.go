package main

import "time"

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
