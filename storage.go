package dh

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/jmoiron/sqlx"
)

//go:embed sql
var dhMigrations embed.FS

// DHMigrations is an [fs.FS] containing pre-made migrations for the migration
// storage.  Currently contains:
//
//   - 000-sqlite
var DHMigrations fs.FS

func init() {
	var err error
	DHMigrations, err = fs.Sub(dhMigrations, "sql")
	if err != nil {
		panic(err)
	}
}

// DoesMigrationStorage is the interface a type must conform to to be able to
// be used by dh.
type DoesMigrationStorage interface {
	// StoreVersion inserts v in db.
	StoreVersion(db sqlx.Execer, version string) error

	// CurrentVersion returns the version stored in db.
	CurrentVersion(db sqlx.Ext) (string, error)
}

type MigrationStorage struct{}

func (s MigrationStorage) StoreVersion(dbh sqlx.Execer, v string) error {
	_, err := dbh.Exec("INSERT INTO dh_migrations (version) VALUES (?)", v)
	return err
}

func (s MigrationStorage) CurrentVersion(dbh sqlx.Ext) (string, error) {
	var found struct{ Version string }
	if err := sqlx.Get(dbh, &found, "SELECT version FROM dh_migrations ORDER BY id DESC LIMIT 1"); err != nil {
		return "", fmt.Errorf("dbh.Select: %w", err)
	}

	return found.Version, nil
}
