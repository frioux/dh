package dh_test

import (
	"database/sql"
	"embed"
	"io/fs"
	"math/rand"
	"testing"
	"path/filepath"

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

//go:embed testdata/*
var testdataFS embed.FS

func TestDH(t *testing.T) {
	migrationDir := mustSub(testdataFS, "testdata/simple")

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

type exampleGoMigrator struct {}

func (m exampleGoMigrator) Migrate(db *sql.Tx, f fs.File) error {
	s, err := f.Stat()
	if err != nil {
		return err
	}

	if filepath.Ext(s.Name()) == ".sql" {
		return dh.SQLMigrator{}.Migrate(db, f)
	}

	if s.Name() == "001-populate.gom" {
		stmt, err := db.Prepare("INSERT INTO frew (a, b, c) VALUES (?, ?, ?)")
		if err != nil {
			return err
		}

		for i := 0; i < 1000; i++ {
			if _, err := stmt.Exec(rand.Int(), rand.Int(), rand.Int()); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestGoMigrator(t *testing.T) {
	migrationDir := mustSub(testdataFS, "testdata/complex")

	dbh, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	db := sqlx.NewDb(dbh, "sqlite3")
	e := dh.NewMigrator()
	e.DoesMigrate = exampleGoMigrator{}
	if err := e.MigrateOne(db, dh.DHMigrations, "000-sqlite"); err != nil {
		panic(err)
	}
	if err := e.MigrateAll(db, migrationDir); err != nil {
		panic(err)
	}
}
