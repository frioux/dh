package dh

import (
	"fmt"
	"path/filepath"
	"encoding/json"
	"io"
	"io/fs"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DoesMigrate interface {
	Migrate(*sql.Tx, fs.File) error
}

type Migrator struct {
	ms MigrationStorage
	p Plan
	m DoesMigrate
}

func NewMigrator(m DoesMigrate) Migrator { return Migrator{m: m} }

func (m Migrator) MigrateOne(dbh *sqlx.DB, d fs.FS, version string) error {
	tx, err := dbh.Begin()
	if err != nil {
		return fmt.Errorf("dbh.Begin: %w", err)
	}
	d, err = fs.Sub(d, version)
	if err != nil {
		return fmt.Errorf("fs.Sub: %w", err)
	}
	if err := m.MigrateDir(tx, d); err != nil {
		return fmt.Errorf("m.MigrateDir: %w", err)
	}
	if err := m.ms.StoreVersion(tx, version); err != nil {
		return fmt.Errorf("m.ms.StoreVersion: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (m Migrator) MigrateAll(dbh *sqlx.DB, d fs.FS) error {
	cv, err := m.ms.CurrentVersion(dbh)
	if err != nil {
		return err
	}
	f, err := d.Open("plan.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	ps, err := m.p.Parse(f)
	if err != nil {
		return err
	}

	var (
		planStart int
		found bool
	)
	for i, plan := range ps {
		if plan == cv {
			planStart = i
			found = true
		}
	}
	if !found {
		return fmt.Errorf("current version (%s) not found in plan.txt", cv)
	}

	for _, plan := range ps[planStart+1:] {
		if err := m.MigrateOne(dbh, d, plan); err != nil {
			return fmt.Errorf("m.MigrateOne: %w", err)
		}
	}

	return nil
}

func (m Migrator) MigrateDir(dbh *sql.Tx, d fs.FS) error {
	des, err := fs.ReadDir(d, ".")
	if err != nil {
		return fmt.Errorf("fs.ReadDir: %w", err)
	}

	for _, de := range des {
		// fmt.Println("migrating", de.Name())
		f, err := d.Open(de.Name())
		if err != nil {
			return fmt.Errorf("fs.Open: %w", err)
		}
		defer f.Close()
		if err := m.m.Migrate(dbh, f); err != nil {
			return fmt.Errorf("m.Migrate: %w", err)
		}
	}
	return nil
}

type ExtensionMigrator struct{
	ms MigrationStorage
	p Plan
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
