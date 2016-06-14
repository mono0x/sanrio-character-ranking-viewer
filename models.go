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
	Rank  int
	Name  string
	Count int
}

func GetCurrentRanking(executor gorp.SqlExecutor, time time.Time) (*Ranking, error) {
	var ranking Ranking
	if err := executor.SelectOne(&ranking, `
		SELECT * FROM ranking
		WHERE $1 BETWEEN started_on AND ended_on
		ORDER BY id DESC
		LIMIT 1
	`, time); err != nil {
		return nil, err
	}
	return &ranking, nil
}

func GetLatestRanking(executor gorp.SqlExecutor, time time.Time) (*Ranking, error) {
	var ranking Ranking
	if err := executor.SelectOne(&ranking, `
		SELECT * FROM ranking
		WHERE $1 >= started_on
		ORDER BY id DESC
		LIMIT 1
	`, time); err != nil {
		return nil, err
	}
	return &ranking, nil
}

func FindCharacterByName(executor gorp.SqlExecutor, name string) (*Character, error) {
	var character Character
	if err := executor.SelectOne(&character,
		`SELECT * FROM character WHERE name = $1 LIMIT 1`); err != nil {
		return nil, err
	}
	return &character, nil
}

func FindEntryByName(executor gorp.SqlExecutor, rankingId int, name string) (*Entry, error) {
	var entry Entry
	if err := executor.SelectOne(&entry, `
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

func GetEntryCharacters(executor gorp.SqlExecutor, rankingId int) ([]Character, error) {
	var characters []Character
	if _, err := executor.Select(&characters, `
		SELECT character.* FROM character
		JOIN entry ON
			entry.character_id = character.id
			AND entry.ranking_id = $1
	`, rankingId); err != nil {
		return nil, err
	}
	return characters, nil
}

func GetRankingItems(executor gorp.SqlExecutor, rankingId int) ([]RankingItem, error) {
	var rankingItems []RankingItem
	if _, err := executor.Select(&rankingItems, `
		SELECT
			RANK() OVER (ORDER BY x.count DESC) AS Rank,
			character.name AS Name,
			x.count AS Count
		FROM (
			SELECT
				vote.character_id, COUNT(*) AS count
			FROM vote
			JOIN entry ON
				entry.ranking_id = vote.ranking_id
				AND entry.character_id = vote.character_id
			WHERE vote.ranking_id = $1
			GROUP BY vote.character_id
		) x
		JOIN character ON character.id = x.character_id
		ORDER BY x.count DESC, character.name ASC
	`, rankingId); err != nil {
		return nil, err
	}
	return rankingItems, nil
}
