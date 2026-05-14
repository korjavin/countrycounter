package store

import (
	"reflect"
	"sort"
	"testing"
)

// newTestRepo opens a fresh in-memory DB, runs migrations, and returns a repo
// wired to it. The DB is closed via t.Cleanup so each test gets an isolated
// schema.
func newTestRepo(t *testing.T) *VisitsRepo {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	return NewVisitsRepo(db)
}

func TestVisitsRepo_List_EmptyUser(t *testing.T) {
	repo := newTestRepo(t)

	got, err := repo.List(42)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if got == nil {
		t.Fatalf("List returned nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("List returned %v, want empty slice", got)
	}
}

func TestVisitsRepo_AddAndList(t *testing.T) {
	repo := newTestRepo(t)

	countries := []string{"Germany", "France", "Japan"}
	for _, c := range countries {
		if err := repo.Add(1, c); err != nil {
			t.Fatalf("Add(%q) failed: %v", c, err)
		}
	}

	got, err := repo.List(1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	// Order is added_at-ascending; since the inserts happen in the same
	// transaction-less rapid sequence, added_at may tie. Compare as sets.
	sortedGot := append([]string(nil), got...)
	sort.Strings(sortedGot)
	sortedWant := append([]string(nil), countries...)
	sort.Strings(sortedWant)
	if !reflect.DeepEqual(sortedGot, sortedWant) {
		t.Errorf("List = %v, want %v (any order)", got, countries)
	}
}

func TestVisitsRepo_Add_IsIdempotent(t *testing.T) {
	repo := newTestRepo(t)

	if err := repo.Add(7, "Spain"); err != nil {
		t.Fatalf("first Add failed: %v", err)
	}
	if err := repo.Add(7, "Spain"); err != nil {
		t.Fatalf("second Add (should be no-op) failed: %v", err)
	}

	got, err := repo.List(7)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"Spain"}) {
		t.Errorf("List = %v, want [Spain] (single row, not duplicated)", got)
	}
}

func TestVisitsRepo_List_IsolatedPerUser(t *testing.T) {
	repo := newTestRepo(t)

	if err := repo.Add(1, "Germany"); err != nil {
		t.Fatalf("Add(1, Germany): %v", err)
	}
	if err := repo.Add(2, "France"); err != nil {
		t.Fatalf("Add(2, France): %v", err)
	}

	got1, err := repo.List(1)
	if err != nil {
		t.Fatalf("List(1): %v", err)
	}
	if !reflect.DeepEqual(got1, []string{"Germany"}) {
		t.Errorf("List(1) = %v, want [Germany]", got1)
	}

	got2, err := repo.List(2)
	if err != nil {
		t.Fatalf("List(2): %v", err)
	}
	if !reflect.DeepEqual(got2, []string{"France"}) {
		t.Errorf("List(2) = %v, want [France]", got2)
	}
}

func TestVisitsRepo_Delete(t *testing.T) {
	repo := newTestRepo(t)
	if err := repo.Add(1, "Germany"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	removed, err := repo.Delete(1, "Germany")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !removed {
		t.Errorf("Delete returned false, want true for existing row")
	}

	got, err := repo.List(1)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("List after delete = %v, want empty", got)
	}
}

func TestVisitsRepo_Delete_MissingRow(t *testing.T) {
	repo := newTestRepo(t)

	removed, err := repo.Delete(1, "Nowhere")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if removed {
		t.Errorf("Delete returned true for missing row, want false")
	}
}

func TestVisitsRepo_Delete_OnlyTargetedRow(t *testing.T) {
	repo := newTestRepo(t)
	if err := repo.Add(1, "Germany"); err != nil {
		t.Fatalf("Add(1, Germany): %v", err)
	}
	if err := repo.Add(1, "France"); err != nil {
		t.Fatalf("Add(1, France): %v", err)
	}
	if err := repo.Add(2, "Germany"); err != nil {
		t.Fatalf("Add(2, Germany): %v", err)
	}

	removed, err := repo.Delete(1, "Germany")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !removed {
		t.Errorf("Delete returned false, want true")
	}

	got1, err := repo.List(1)
	if err != nil {
		t.Fatalf("List(1): %v", err)
	}
	if !reflect.DeepEqual(got1, []string{"France"}) {
		t.Errorf("List(1) = %v, want [France]", got1)
	}

	got2, err := repo.List(2)
	if err != nil {
		t.Fatalf("List(2): %v", err)
	}
	if !reflect.DeepEqual(got2, []string{"Germany"}) {
		t.Errorf("List(2) = %v, want [Germany] (other user's row should survive)", got2)
	}
}

func TestVisitsRepo_Has(t *testing.T) {
	repo := newTestRepo(t)
	if err := repo.Add(1, "Germany"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	cases := []struct {
		name    string
		userID  int64
		country string
		want    bool
	}{
		{"exact match", 1, "Germany", true},
		{"different user", 2, "Germany", false},
		{"different country", 1, "France", false},
		{"both different", 99, "Atlantis", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.Has(tc.userID, tc.country)
			if err != nil {
				t.Fatalf("Has(%d, %q): %v", tc.userID, tc.country, err)
			}
			if got != tc.want {
				t.Errorf("Has(%d, %q) = %v, want %v", tc.userID, tc.country, got, tc.want)
			}
		})
	}
}

func TestVisitsRepo_UnicodeCountryNames(t *testing.T) {
	repo := newTestRepo(t)

	names := []string{"日本", "België", "Côte d'Ivoire", "Россия"}
	for _, n := range names {
		if err := repo.Add(1, n); err != nil {
			t.Fatalf("Add(%q): %v", n, err)
		}
	}

	for _, n := range names {
		has, err := repo.Has(1, n)
		if err != nil {
			t.Fatalf("Has(%q): %v", n, err)
		}
		if !has {
			t.Errorf("Has(%q) = false, want true", n)
		}
	}

	got, err := repo.List(1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != len(names) {
		t.Errorf("List length = %d, want %d (%v)", len(got), len(names), got)
	}
}
