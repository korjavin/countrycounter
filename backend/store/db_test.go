package store

import (
	"path/filepath"
	"testing"
)

func TestOpen_InMemory(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping() on opened DB failed: %v", err)
	}
}

func TestOpen_FileInTempDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", path, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping() on opened DB failed: %v", err)
	}
}

func TestOpen_UnwritablePath(t *testing.T) {
	// A path inside a non-existent directory cannot be created by SQLite.
	path := filepath.Join(t.TempDir(), "does", "not", "exist", "test.db")
	db, err := Open(path)
	if err == nil {
		db.Close()
		t.Fatalf("Open(%q) unexpectedly succeeded", path)
	}
}

func TestOpen_WALModeSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}
}
