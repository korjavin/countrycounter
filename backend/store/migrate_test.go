package store

import (
	"testing"
)

func TestMigrate_FreshDB(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var name string
	row := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'visits'`)
	if err := row.Scan(&name); err != nil {
		t.Fatalf("visits table not found after Migrate: %v", err)
	}
	if name != "visits" {
		t.Errorf("expected table name 'visits', got %q", name)
	}

	var idxName string
	row = db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'idx_visits_user_id'`)
	if err := row.Scan(&idxName); err != nil {
		t.Fatalf("idx_visits_user_id index not found after Migrate: %v", err)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("first Migrate failed: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}

	var count int
	row := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'visits'`)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query visits table count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 visits table, got %d", count)
	}
}

func TestMigrate_VisitsSchema(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO visits (user_id, country_name) VALUES (?, ?)`, int64(1), "Germany"); err != nil {
		t.Fatalf("insert into visits: %v", err)
	}

	_, err = db.Exec(`INSERT INTO visits (user_id, country_name) VALUES (?, ?)`, int64(1), "Germany")
	if err == nil {
		t.Errorf("expected primary key violation on duplicate insert, got nil")
	}
}
