package dh

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type StorageVersion struct {
	Version string
}

func (v StorageVersion) String() string {
	return v.Version
}

type MigrationStorage struct {}

func (s MigrationStorage) StoreVersion(dbh sqlx.Execer, v string) error {
	_, err := dbh.Exec("INSERT INTO dh_migrations (version) VALUES (?)", v)
	return err
}

func (s MigrationStorage) CurrentVersion(dbh sqlx.Ext) (string, error) {
	var found StorageVersion
	if err := sqlx.Get(dbh, &found, "SELECT version FROM dh_migrations ORDER BY id DESC LIMIT 1"); err != nil {
		return "", fmt.Errorf("dbh.Select: %w", err)
	}

	return found.Version, nil
}

