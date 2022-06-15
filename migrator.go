// package dh is a tool for maintaining RDBMS versions.
//
// # Synopsis
//
//	db := sqlx.NewDb(dbh, "sqlite3")
//	e := dh.NewMigrator(dh.ExtensionMigrator{})
//	if err := e.MigrateOne(db, dh.DHMigrations, "000-sqlite"); err != nil {
//		panic(err)
//	}
//	if err := e.MigrateAll(db, migrationDir); err != nil {
//		panic(err)
//	}
//
// # Description
//
// dh is inspired by my own [DBIx::Class::DeploymentHandler], which worked well
// but required the use of an ORM, and additionally had some other limitations,
// like needing to reliably split the SQL in the migration files into
// statements, which ended up being pretty frustrating.
//
// dh simplifies the situation dramatically.  You create directories of
// migrations; out of the box these migrations can contain SQL files (defined
// by having the `.sql` extension) or JSON files (defined by having the `.json`
// extension, containing simply an array of strings to be run as SQL
// statements.)
//
// In addition to directories of files, at the same level as the directories
// you define a plan (in `plan.txt`) that simply lists the migrations to be
// applied in order.  Comments (starting with `#`) are ignored, as are blank
// lines.
//
//	$ ls -F migrations
//	001/  002/  plan.txt
//	$ cat plan.txt
//	000-sqlite # built into dh
//	001
//	002
//
// ## The First Migration
//
// The first migration is special, as it must create the table that dh uses to
// store which migrations have been applied.  To work with the built in
// [MigrationStorage] it should be named `dh_migrations` and only needs two
// columns, `version`, which is the directory name applied, and `id`, which
// must increase with each applied version, so that queries for the last `id`
// will return the most recently applied version.
//
// dh includes a migration fit for the first migration that works for SQLite,
// which you can apply by doing something like this:
//
//	e := dh.NewMigrator(dh.ExtensionMigrator{})
//	err := e.MigrateOne(db, dh.DHMigrations, "000-sqlite")
//
// If people submit MRs for other databases I'll happily have them.
//
// [DBIx::Class::DeploymentHandler]: https://metacpan.org/pod/DBIx::Class::DeploymentHandler
package dh

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

// DoesMigrate are things that can apply migrations to the database.
//
// Built in examples include the [JSONMigrator], [SQLMigrator], and the
// related dispatcher, [ExtensionMigrator].  There's an example of a slightly
// more complicated migrator in the tests.
type DoesMigrate interface {
	// Migrate applies a file to the database within the current transaction.
	Migrate(*sql.Tx, fs.File) error
}

// Migrator is the main type to be used within dh.  It orchestrates the migrations
// based on the provided plan.
type Migrator struct {
	p Plan

	DoesMigrationStorage
	DoesMigrate
}

// NewMigrator returns a Migrator instance with [MigrationStorage] and the
// [ExtensionMigrator].  Create your own Migrator instace or replace those
// values if you need to.
func NewMigrator() Migrator {
	return Migrator{
		DoesMigrationStorage: MigrationStorage{},
		DoesMigrate:          ExtensionMigrator{},
	}
}

// MigrateOne applies a migration directory from d named version.
//
// Note that the version storage must exist by the time MigrateOne completes, or
// an error will be returned.
func (m Migrator) MigrateOne(dbh *sqlx.DB, d fs.FS, version string) error {
	tx, err := dbh.Begin()
	if err != nil {
		return fmt.Errorf("dbh.Begin: %w", err)
	}
	d, err = fs.Sub(d, version)
	if err != nil {
		return fmt.Errorf("fs.Sub: %w", err)
	}
	if err := m.migrateDir(tx, d); err != nil {
		return fmt.Errorf("m.migrateDir: %w", err)
	}
	if err := m.StoreVersion(tx, version); err != nil {
		return fmt.Errorf("m.ms.StoreVersion: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

// MigrateAll applies the migrations listed in `plan.txt`, starting
// with the migration after the one most recently applied to the
// database.
func (m Migrator) MigrateAll(dbh *sqlx.DB, d fs.FS) error {
	cv, err := m.CurrentVersion(dbh)
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
		found     bool
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

// migrateDir applies each migration within d.
func (m Migrator) migrateDir(dbh *sql.Tx, d fs.FS) error {
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
		if err := m.Migrate(dbh, f); err != nil {
			return fmt.Errorf("m.Migrate: %w", err)
		}
	}
	return nil
}

// ExtensionMigrator applies migrations based on their suffix.  It is
// intentionally simple; rather than submitting Pull Requests to add support
// for (eg) YAML, you are encouraged to maintain your own.
type ExtensionMigrator struct {
	ms MigrationStorage
	p  Plan
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

// SQLMigrator applies a file as a single Exec call.  If that will cause issues
// with your database, you can either switch to [JSONMigrator], or add a SQL
// Parser to your project to properly split statements.
type SQLMigrator struct{}

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

// JSONMigrator applies statements in a json file read as an array of strings.
// For example you could have a file that looks like this:
//
//	[
//	  "INSERT INTO foo (a, b, c) VALUES (1, 2, 3)",
//	  "INSERT INTO foo (a, b, c) VALUES (3, 2, 1)",
//	  "INSERT INTO foo (a, b, c) VALUES (9, 9, 9)"
//	]
type JSONMigrator struct{}

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
