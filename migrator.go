package dh

import (
	"fmt"
	"path/filepath"
	"encoding/json"
	"io"
	"io/fs"
	"database/sql"
)

type Migrator interface {
	Migrate(sql.Tx, fs.File) error
}

type ExtensionMigrator struct{}

func (m ExtensionMigrator) MigrateAll(dbh *sql.Tx, d fs.FS) error {
	// 1. find out what version the database is at
	// 2. if none, deploy max deploy
	// 3. MigrateDir 
	return nil
}

func (m ExtensionMigrator) MigrateDir(dbh *sql.Tx, d fs.FS) error {
	des, err := fs.ReadDir(d, ".")
	if err != nil {
		return fmt.Errorf("fs.ReadDir: %w", err)
	}

	for _, de := range des {
		f, err := d.Open(de.Name())
		if err != nil {
			return fmt.Errorf("fs.Open: %w", err)
		}
		defer f.Close()
		m.Migrate(dbh, f)
	}
	return nil
}

func (m ExtensionMigrator) Migrate(dbh *sql.Tx, f fs.File) error {
	s, err := f.Stat()
	if err != nil {
		return fmt.Errorf("fs.File.Stat: %w", err)
	}

	switch e := filepath.Ext(s.Name()); e {
		case ".sql":
			return SQLMigrator{}.Migrate(dbh, f)
		case ".json":
			return JSONMigrator{}.Migrate(dbh, f)
		default:
			return fmt.Errorf("Unknown extension: %s", e)
	}
}

type SQLMigrator struct {}

func (m SQLMigrator) Migrate(dbh *sql.Tx, f fs.File) error {
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if _, err := dbh.Exec(string(b)); err != nil {
		return fmt.Errorf("dbh.Exec: %w", err)
	}

	return nil
}

type JSONMigrator struct {}

func (m JSONMigrator) Migrate(dbh *sql.Tx, f fs.File) error {
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	var statements []string
	if err := json.Unmarshal(b, &statements); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	for i, stmt := range statements {
		if _, err := dbh.Exec(stmt); err != nil {
			return fmt.Errorf("dbh.Exec (%d): %w", i, err)
		}
	}

	return nil
}
