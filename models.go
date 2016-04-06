package main

import (
	"time"

	"gopkg.in/gorp.v1"
)

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

type Entry struct {
	RankingId   int `db:"ranking_id"`
	CharacterId int `db:"character_id"`
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

func FindCharacterByName(dbMap *gorp.DbMap, name string) (*Character, error) {
	var character Character
	if err := dbMap.SelectOne(&character, `SELECT * FROM character WHERE name = :name LIMIT 1`,
		map[string]string{"name": name}); err != nil {
		return nil, err
	}
	return &character, nil
}

func FindEntryByName(dbMap *gorp.DbMap, rankingId int, name string) (*Entry, error) {
	var entry Entry
	if err := dbMap.SelectOne(&entry, `
		SELECT * FROM entry
		JOIN character ON character.id = entry.character_id
		WHERE
			entry.ranking_id = :ranking_id
			AND character.name = :name
		LIMIT 1
	`, map[string]string{
		"ranking_id": string(rankingId),
		"name":       name,
	}); err != nil {
		return nil, err
	}
	return &entry, nil
}
