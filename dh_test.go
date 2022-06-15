package dh_test

import (
	"database/sql"
	"embed"
	"io/fs"
	"testing"

	"github.com/frioux/dh"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func mustSub(fss fs.FS, dir string) fs.FS {
	ret, err := fs.Sub(fss, dir)
	if err != nil {
		panic(err)
	}
	return ret
}

//go:embed testdata/simple/*
var simpleFS embed.FS

func TestDH(t *testing.T) {
	migrationDir := mustSub(simpleFS, "testdata/simple")

	dbh, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	db := sqlx.NewDb(dbh, "sqlite3")
	e := dh.NewMigrator()
	if err := e.MigrateOne(db, dh.DHMigrations, "000-sqlite"); err != nil {
		panic(err)
	}
	if err := e.MigrateAll(db, migrationDir); err != nil {
		panic(err)
	}
}
