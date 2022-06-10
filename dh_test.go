package dh_test

import (
	"testing"
	"database/sql"
	"embed"
	"io/fs"

	"github.com/frioux/dh"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed testdata/simple/*
var simpleFS embed.FS

func TestDH(t *testing.T) {
	dbh, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	e := dh.ExtensionMigrator{}
	tx, err := dbh.Begin()
	if err != nil {
		panic(err)
	}
	l, err := fs.Sub(simpleFS, "testdata/simple")
	if err != nil {
		panic(err)
	}
	e.MigrateDir(tx, l)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}
