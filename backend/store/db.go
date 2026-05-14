// Package store owns the shared SQLite connection and the visits repository.
//
// Open is called once at startup; the resulting *sql.DB is then passed to the
// repository constructors and to Migrate. MaxOpenConns is set to 1 so the
// single WAL writer never contends with itself.
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

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
