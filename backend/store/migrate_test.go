package store

import (
	"testing"

	"github.com/pressly/goose/v3"
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

// TestMigrate_NormalizesLegacyNames verifies that migration 002 rewrites
// legacy `Cape Verde` / `Palestine` rows to the canonical `Cabo Verde` /
// `Palestine State` names used by the autocomplete and iso_mapping.go. Users
// who already have both legacy and canonical rows keep only the canonical one.
func TestMigrate_NormalizesLegacyNames(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.UpTo(db, "migrations", 1); err != nil {
		t.Fatalf("apply migration 001: %v", err)
	}

	seed := []struct {
		uid  int64
		name string
	}{
		{1, "Cape Verde"},      // legacy only — should be renamed
		{2, "Palestine"},       // legacy only — should be renamed
		{3, "Cape Verde"},      // user has both legacy + canonical
		{3, "Cabo Verde"},      //   → legacy row dropped, canonical kept
		{4, "Palestine"},       // user has both legacy + canonical
		{4, "Palestine State"}, //   → legacy row dropped, canonical kept
		{5, "Germany"},         // unrelated row — should be untouched
	}
	for _, r := range seed {
		if _, err := db.Exec(`INSERT INTO visits (user_id, country_name) VALUES (?, ?)`, r.uid, r.name); err != nil {
			t.Fatalf("seed insert (%d, %q): %v", r.uid, r.name, err)
		}
	}

	if err := goose.UpTo(db, "migrations", 2); err != nil {
		t.Fatalf("apply migration 002: %v", err)
	}

	var leftover int
	if err := db.QueryRow(`SELECT COUNT(*) FROM visits WHERE country_name IN ('Cape Verde', 'Palestine')`).Scan(&leftover); err != nil {
		t.Fatalf("query leftover legacy rows: %v", err)
	}
	if leftover != 0 {
		t.Errorf("expected 0 legacy rows after migration, got %d", leftover)
	}

	expected := map[int64]string{
		1: "Cabo Verde",
		2: "Palestine State",
		3: "Cabo Verde",
		4: "Palestine State",
		5: "Germany",
	}
	for uid, want := range expected {
		var got string
		if err := db.QueryRow(`SELECT country_name FROM visits WHERE user_id = ?`, uid).Scan(&got); err != nil {
			t.Fatalf("query user %d: %v", uid, err)
		}
		if got != want {
			t.Errorf("user %d: got %q, want %q", uid, got, want)
		}
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM visits WHERE user_id = ?`, uid).Scan(&count); err != nil {
			t.Fatalf("count user %d: %v", uid, err)
		}
		if count != 1 {
			t.Errorf("user %d: expected 1 row, got %d", uid, count)
		}
	}
}
