package main

import (
	"time"

	"gopkg.in/gorp.v1"
)

type Character struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

func FindCharacterByName(dbMap *gorp.DbMap, name string) (*Character, error) {
	var character Character
	if err := dbMap.SelectOne(&character,
		`SELECT * FROM character WHERE name = :name LIMIT 1`,
		map[string]string{"name": name}); err != nil {
		return nil, err
	}
	return &character, nil
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
