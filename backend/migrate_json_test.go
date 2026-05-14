package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/korjavin/countrycounter/backend/store"
)

// newImporterRepo opens a fresh in-memory DB, migrates it, and returns a wired
// repo. Each test gets its own DB.
func newImporterRepo(t *testing.T) *store.VisitsRepo {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:): %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return store.NewVisitsRepo(db)
}

// writeJSON writes raw bytes to a temp file in t's tempdir and returns its path.
func writeJSON(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp json: %v", err)
	}
	return path
}

func TestMaybeImportJSON_EmptyDB_ValidJSON(t *testing.T) {
	repo := newImporterRepo(t)

	const payload = `{
		"1": ["Germany", "France", "Japan"],
		"2": ["Spain", "Italy"],
		"3": ["Brazil", "Argentina"]
	}`
	path := writeJSON(t, payload)

	n, err := MaybeImportJSON(repo, path)
	if err != nil {
		t.Fatalf("MaybeImportJSON: %v", err)
	}
	if n != 7 {
		t.Errorf("imported = %d, want 7", n)
	}

	cases := map[int64][]string{
		1: {"France", "Germany", "Japan"},
		2: {"Italy", "Spain"},
		3: {"Argentina", "Brazil"},
	}
	for userID, want := range cases {
		got, err := repo.List(userID)
		if err != nil {
			t.Fatalf("List(%d): %v", userID, err)
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%d) = %v, want %v", userID, got, want)
		}
	}
}

func TestMaybeImportJSON_EmptyDB_NoFile(t *testing.T) {
	repo := newImporterRepo(t)

	missing := filepath.Join(t.TempDir(), "does-not-exist.json")
	n, err := MaybeImportJSON(repo, missing)
	if err != nil {
		t.Fatalf("MaybeImportJSON: %v", err)
	}
	if n != 0 {
		t.Errorf("imported = %d, want 0", n)
	}
}

func TestMaybeImportJSON_EmptyDB_EmptyJSON(t *testing.T) {
	repo := newImporterRepo(t)
	path := writeJSON(t, `{}`)

	n, err := MaybeImportJSON(repo, path)
	if err != nil {
		t.Fatalf("MaybeImportJSON: %v", err)
	}
	if n != 0 {
		t.Errorf("imported = %d, want 0", n)
	}

	empty, err := repo.IsEmpty()
	if err != nil {
		t.Fatalf("IsEmpty: %v", err)
	}
	if !empty {
		t.Errorf("db not empty after importing empty JSON")
	}
}

func TestMaybeImportJSON_EmptyDB_MalformedJSON(t *testing.T) {
	repo := newImporterRepo(t)
	path := writeJSON(t, `{this is not json`)

	n, err := MaybeImportJSON(repo, path)
	if err == nil {
		t.Fatalf("MaybeImportJSON returned nil error for malformed JSON, want error")
	}
	if n != 0 {
		t.Errorf("imported = %d on parse error, want 0", n)
	}
}

func TestMaybeImportJSON_PopulatedDB_SkipsAndPreserves(t *testing.T) {
	repo := newImporterRepo(t)
	if err := repo.Add(1, "Germany"); err != nil {
		t.Fatalf("seed Add: %v", err)
	}

	path := writeJSON(t, `{"1": ["France"], "2": ["Japan"]}`)

	n, err := MaybeImportJSON(repo, path)
	if err != nil {
		t.Fatalf("MaybeImportJSON: %v", err)
	}
	if n != 0 {
		t.Errorf("imported = %d, want 0 when DB already populated", n)
	}

	got1, err := repo.List(1)
	if err != nil {
		t.Fatalf("List(1): %v", err)
	}
	if !reflect.DeepEqual(got1, []string{"Germany"}) {
		t.Errorf("List(1) = %v, want [Germany] (existing rows must not be merged/overwritten)", got1)
	}
	got2, err := repo.List(2)
	if err != nil {
		t.Fatalf("List(2): %v", err)
	}
	if len(got2) != 0 {
		t.Errorf("List(2) = %v, want empty (JSON should be ignored)", got2)
	}
}

func TestMaybeImportJSON_IdempotentSecondCall(t *testing.T) {
	repo := newImporterRepo(t)
	path := writeJSON(t, `{"1": ["Germany", "France"]}`)

	n1, err := MaybeImportJSON(repo, path)
	if err != nil {
		t.Fatalf("first MaybeImportJSON: %v", err)
	}
	if n1 != 2 {
		t.Errorf("first import = %d, want 2", n1)
	}

	n2, err := MaybeImportJSON(repo, path)
	if err != nil {
		t.Fatalf("second MaybeImportJSON: %v", err)
	}
	if n2 != 0 {
		t.Errorf("second import = %d, want 0 (idempotent — DB now non-empty)", n2)
	}

	got, err := repo.List(1)
	if err != nil {
		t.Fatalf("List(1): %v", err)
	}
	sort.Strings(got)
	want := []string{"France", "Germany"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List(1) = %v, want %v", got, want)
	}
}
