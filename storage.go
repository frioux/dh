package dh

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Version struct {
	Version string
}

type MigrationStorage struct {}

func (s MigrationStorage) StoreVersion(dbh *sqlx.DB, v Version) error {
	_, err := dbh.Exec("INSERT INTO dh_migrations (version) VALUES (?)", v.Version)
	return err
}

func (s MigrationStorage) CurrentVersion(dbh *sqlx.DB) (string, error) {
	var found Version
	if err := dbh.Get(&found, "SELECT MAX(version) as version FROM dh_migrations"); err != nil {
		return "", fmt.Errorf("dbh.Select: %w", err)
	}

	return found.Version, nil
}

