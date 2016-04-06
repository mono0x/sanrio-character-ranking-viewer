package main

import (
	"database/sql"

	_ "github.com/lib/pq"

	"gopkg.in/gorp.v1"
)

type appContext struct {
	dbMap *gorp.DbMap
}

func newContext() (*appContext, error) {
	context := &appContext{}

	db, err := sql.Open("postgres", "user=app dbname=sanrio-character-ranking sslmode=disable")
	if err != nil {
		return nil, err
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbMap.AddTableWithName(Character{}, "character").SetKeys(true, "Id")
	dbMap.AddTableWithName(Status{}, "status").SetKeys(true, "Id")
	dbMap.AddTableWithName(Ranking{}, "ranking").SetKeys(true, "Id")
	dbMap.AddTableWithName(Entry{}, "entry")
	dbMap.AddTableWithName(Vote{}, "vote").SetKeys(true, "status_id")

	context.dbMap = dbMap

	return context, nil
}

func (c *appContext) Close() {
	c.dbMap.Db.Close()
}
