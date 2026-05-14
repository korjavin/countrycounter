// Package store owns the shared SQLite connection and the visits repository.
//
// Open is called once at startup; the resulting *sql.DB is then passed to the
// repository constructors and to Migrate. MaxOpenConns is capped at 1 so all
// reads and writes serialize through a single connection — combined with WAL
// mode and the 5s busy_timeout this matches the medicationtrackerbot pattern
// and removes the need for application-level locking.
package store

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Open opens a SQLite database at the given path with the project's standard
// pragmas (WAL journal, 5s busy_timeout) and MaxOpenConns=1.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}

	db.SetMaxOpenConns(1)

	return db, nil
}

// Migrate runs the embedded goose migrations against db. Calling Migrate more
// than once is safe — goose tracks applied versions in goose_db_version and
// only runs pending migrations.
func Migrate(db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
