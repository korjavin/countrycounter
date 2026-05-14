package store

import (
	"database/sql"
	"fmt"
)

// VisitsRepo is the storage layer for the visits table. It owns no state of
// its own beyond the shared *sql.DB; multiple goroutines may call its methods
// concurrently — Open caps the connection pool at 1, so all reads and writes
// serialize through the same SQLite connection.
type VisitsRepo struct {
	db *sql.DB
}

// NewVisitsRepo wires a repo to an already-opened, already-migrated *sql.DB.
func NewVisitsRepo(db *sql.DB) *VisitsRepo {
	return &VisitsRepo{db: db}
}

// List returns the country names visited by userID, ordered by when they were
// added. The result is a non-nil empty slice when the user has no visits.
func (r *VisitsRepo) List(userID int64) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT country_name FROM visits WHERE user_id = ? ORDER BY added_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query visits: %w", err)
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan visit row: %w", err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate visit rows: %w", err)
	}
	return out, nil
}

// Add records a visit. It is idempotent: a second call with the same
// (userID, country) is a no-op rather than an error, matching the bot's
// existing "already visited" UX.
func (r *VisitsRepo) Add(userID int64, country string) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO visits (user_id, country_name) VALUES (?, ?)`,
		userID, country,
	)
	if err != nil {
		return fmt.Errorf("insert visit: %w", err)
	}
	return nil
}

// Delete removes a visit. The bool is true when a row was actually removed,
// false when no matching (userID, country) pair existed — handlers use this
// to distinguish 200 from 404.
func (r *VisitsRepo) Delete(userID int64, country string) (bool, error) {
	res, err := r.db.Exec(
		`DELETE FROM visits WHERE user_id = ? AND country_name = ?`,
		userID, country,
	)
	if err != nil {
		return false, fmt.Errorf("delete visit: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return affected > 0, nil
}

// ImportPairs inserts every (userID, country) pair in data within a single
// transaction. If any insert fails the transaction is rolled back so the
// table is left in its prior state, letting the caller safely retry on the
// next start. Returns the number of pairs processed (counting duplicates,
// which become no-ops because of INSERT OR IGNORE).
func (r *VisitsRepo) ImportPairs(data map[int64][]string) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin import tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO visits (user_id, country_name) VALUES (?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare import insert: %w", err)
	}
	defer stmt.Close()

	imported := 0
	for userID, countries := range data {
		for _, country := range countries {
			if _, err := stmt.Exec(userID, country); err != nil {
				return 0, fmt.Errorf("import (user %d, country %q): %w", userID, country, err)
			}
			imported++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit import tx: %w", err)
	}
	return imported, nil
}

// IsEmpty reports whether the visits table contains zero rows. Used by the
// JSON auto-importer to decide whether to seed a fresh DB from the legacy
// data.json file without leaking schema details into the importer.
func (r *VisitsRepo) IsEmpty() (bool, error) {
	var one int
	err := r.db.QueryRow(`SELECT 1 FROM visits LIMIT 1`).Scan(&one)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("query visits emptiness: %w", err)
	}
	return false, nil
}

// Has reports whether (userID, country) is already recorded.
func (r *VisitsRepo) Has(userID int64, country string) (bool, error) {
	var one int
	err := r.db.QueryRow(
		`SELECT 1 FROM visits WHERE user_id = ? AND country_name = ? LIMIT 1`,
		userID, country,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query visit existence: %w", err)
	}
	return true, nil
}
