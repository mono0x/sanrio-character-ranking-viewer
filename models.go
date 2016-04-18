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
	if err := dbMap.SelectOne(&character,
		`SELECT * FROM character WHERE name = $1 LIMIT 1`); err != nil {
		return nil, err
	}
	return &character, nil
}

func FindEntryByName(dbMap *gorp.DbMap, rankingId int, name string) (*Entry, error) {
	var entry Entry
	if err := dbMap.SelectOne(&entry, `
		SELECT entry.* FROM entry
		JOIN character ON character.id = entry.character_id
		WHERE
			entry.ranking_id = $1
			AND character.name = $2
		LIMIT 1
	`, rankingId, name); err != nil {
		return nil, err
	}
	return &entry, nil
}

func GetEntryCharacters(dbMap *gorp.DbMap, rankingId int) ([]Character, error) {
	var characters []Character
	if _, err := dbMap.Select(&characters, `
		SELECT character.* FROM character
		JOIN entry ON
			entry.character_id = character.id
			AND entry.ranking_id = $1
	`, rankingId); err != nil {
		return nil, err
	}
	return characters, nil
}
