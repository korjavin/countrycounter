package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/korjavin/countrycounter/backend/store"
)

// newTestRepo opens an in-memory SQLite, runs migrations, and returns a fresh
// VisitsRepo isolated to this test.
func newTestRepo(t *testing.T) *store.VisitsRepo {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open(:memory:) failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("store.Migrate failed: %v", err)
	}
	return store.NewVisitsRepo(db)
}

// stubGeocoder swaps the package-level geocoder for the duration of a test.
func stubGeocoder(t *testing.T, fn func(lat, lng float64) (string, error)) {
	t.Helper()
	prev := geocodeLocation
	geocodeLocation = fn
	t.Cleanup(func() { geocodeLocation = prev })
}

func TestHandleLocation_AddsNewCountryAndReportsCount(t *testing.T) {
	repo := newTestRepo(t)
	stubGeocoder(t, func(lat, lng float64) (string, error) {
		return "France", nil
	})

	reply, err := handleLocation(repo, 100, 48.8566, 2.3522)
	if err != nil {
		t.Fatalf("handleLocation returned err: %v", err)
	}
	if !strings.Contains(reply, "France") {
		t.Errorf("reply %q should mention the country", reply)
	}
	if !strings.Contains(reply, "1") {
		t.Errorf("reply %q should mention the count of visited countries", reply)
	}
	if !strings.Contains(reply, "Added") {
		t.Errorf("reply %q should be the success/added phrasing", reply)
	}

	got, err := repo.List(100)
	if err != nil {
		t.Fatalf("repo.List: %v", err)
	}
	if len(got) != 1 || got[0] != "France" {
		t.Errorf("repo.List(100) = %v, want [France]", got)
	}
}

func TestHandleLocation_CountIncrementsAcrossMultipleAdds(t *testing.T) {
	repo := newTestRepo(t)

	countries := []string{"France", "Japan", "Brazil"}
	idx := 0
	stubGeocoder(t, func(lat, lng float64) (string, error) {
		c := countries[idx]
		idx++
		return c, nil
	})

	for i, expected := range countries {
		reply, err := handleLocation(repo, 200, 0, 0)
		if err != nil {
			t.Fatalf("handleLocation #%d: %v", i+1, err)
		}
		if !strings.Contains(reply, expected) {
			t.Errorf("reply #%d %q should mention %q", i+1, reply, expected)
		}
		// Count appears as "visited N countries"; check N matches i+1.
		marker := "visited " + countString(i+1) + " countries"
		if !strings.Contains(reply, marker) {
			t.Errorf("reply #%d %q should contain %q", i+1, reply, marker)
		}
	}
}

func countString(n int) string {
	// Tiny helper avoids depending on strconv import duplication and keeps
	// the test readable. Only needs single-digit cases for this test.
	return string(rune('0' + n))
}

func TestHandleLocation_AlreadyVisited(t *testing.T) {
	repo := newTestRepo(t)
	stubGeocoder(t, func(lat, lng float64) (string, error) {
		return "Italy", nil
	})

	if _, err := handleLocation(repo, 300, 41.9, 12.5); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	reply, err := handleLocation(repo, 300, 41.9, 12.5)
	if err != nil {
		t.Fatalf("second call returned err: %v", err)
	}
	if !strings.Contains(reply, "already added") {
		t.Errorf("reply %q should indicate the country was already added", reply)
	}
	if !strings.Contains(reply, "Italy") {
		t.Errorf("reply %q should name the country", reply)
	}

	got, err := repo.List(300)
	if err != nil {
		t.Fatalf("repo.List: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("repo.List(300) = %v, want exactly one entry after duplicate", got)
	}
}

func TestHandleLocation_GeocodingFailureReturnsErrorReply(t *testing.T) {
	repo := newTestRepo(t)
	stubGeocoder(t, func(lat, lng float64) (string, error) {
		return "", errors.New("ocean — no country here")
	})

	reply, err := handleLocation(repo, 400, 0, 0)
	if err == nil {
		t.Errorf("expected non-nil err for geocoding failure (caller should log it)")
	}
	if !strings.Contains(reply, "couldn't determine") {
		t.Errorf("reply %q should explain the geocoding failure to the user", reply)
	}

	got, err := repo.List(400)
	if err != nil {
		t.Fatalf("repo.List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("repo.List(400) = %v, want empty (nothing should be persisted on geocode failure)", got)
	}
}
