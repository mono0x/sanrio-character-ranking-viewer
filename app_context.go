package main

import (
	"database/sql"

	_ "github.com/lib/pq"

	"gopkg.in/gorp.v1"
)

type AppContext struct {
	dbMap *gorp.DbMap
}

func NewAppContext() (*AppContext, error) {
	context := &AppContext{}

	db, err := sql.Open("postgres", "user=app host=127.0.0.1 dbname=sanrio-character-ranking sslmode=disable")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbMap.AddTableWithName(Character{}, "character").SetKeys(true, "Id")
	dbMap.AddTableWithName(Status{}, "status").SetKeys(false, "Id")
	dbMap.AddTableWithName(Ranking{}, "ranking").SetKeys(true, "Id")
	dbMap.AddTableWithName(Entry{}, "entry").SetKeys(false, "RankingId", "CharacterId")
	dbMap.AddTableWithName(Vote{}, "vote").SetKeys(false, "StatusId")

	context.dbMap = dbMap

	return context, nil
}

func (c *AppContext) Close() {
	c.dbMap.Db.Close()
}
