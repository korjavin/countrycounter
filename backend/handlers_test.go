package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/korjavin/countrycounter/backend/store"
)

// newTestServer wires a server to a fresh in-memory SQLite DB so each test
// gets an isolated schema and dataset.
func newTestServer(t *testing.T) *server {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open(:memory:) failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := store.Migrate(db); err != nil {
		t.Fatalf("store.Migrate failed: %v", err)
	}

	return &server{repo: store.NewVisitsRepo(db)}
}

func TestGetCountries_NewUserReturnsEmptyList(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/countries?userId=123", nil)
	rec := httptest.NewRecorder()

	srv.handleCountries(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got []string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v (body=%q)", err, rec.Body.String())
	}
	if got == nil {
		t.Fatalf("decoded nil slice, want empty JSON array")
	}
	if len(got) != 0 {
		t.Errorf("got %v, want empty slice", got)
	}

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestGetCountries_ReturnsRecordedCountries(t *testing.T) {
	srv := newTestServer(t)

	if err := srv.repo.Add(7, "France"); err != nil {
		t.Fatalf("seed Add: %v", err)
	}
	if err := srv.repo.Add(7, "Japan"); err != nil {
		t.Fatalf("seed Add: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/countries?userId=7", nil)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got []string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	want := map[string]bool{"France": true, "Japan": true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want 2 entries matching %v", got, want)
	}
	for _, c := range got {
		if !want[c] {
			t.Errorf("unexpected country %q in response", c)
		}
	}
}

func TestGetCountries_MissingUserId(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/countries", nil)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestGetCountries_InvalidUserId(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/countries?userId=abc", nil)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAddCountry_CreatesAndIsIdempotent(t *testing.T) {
	srv := newTestServer(t)

	body := strings.NewReader(`{"userId":42,"country":"Brazil"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/countries", body)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first POST status = %d, want 201", rec.Code)
	}

	// Second POST with the same payload — should also succeed (idempotent).
	body2 := strings.NewReader(`{"userId":42,"country":"Brazil"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/countries", body2)
	rec2 := httptest.NewRecorder()
	srv.handleCountries(rec2, req2)

	if rec2.Code != http.StatusCreated {
		t.Fatalf("second POST status = %d, want 201", rec2.Code)
	}

	got, err := srv.repo.List(42)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0] != "Brazil" {
		t.Errorf("after duplicate POSTs got %v, want [Brazil] (single entry)", got)
	}
}

func TestAddCountry_MissingBody(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/countries", nil)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAddCountry_MissingFields(t *testing.T) {
	srv := newTestServer(t)

	cases := []string{
		`{"userId":0,"country":"X"}`,
		`{"userId":5,"country":""}`,
		`{}`,
	}
	for _, body := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/countries", strings.NewReader(body))
		rec := httptest.NewRecorder()
		srv.handleCountries(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
	}
}

func TestDeleteCountry_RemovesExisting(t *testing.T) {
	srv := newTestServer(t)
	if err := srv.repo.Add(99, "Italy"); err != nil {
		t.Fatalf("seed Add: %v", err)
	}

	body := strings.NewReader(`{"userId":99,"country":"Italy"}`)
	req := httptest.NewRequest(http.MethodDelete, "/api/countries", body)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	got, err := srv.repo.List(99)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("after delete got %v, want empty", got)
	}
}

func TestDeleteCountry_MissingReturns404(t *testing.T) {
	srv := newTestServer(t)

	body := strings.NewReader(`{"userId":99,"country":"Italy"}`)
	req := httptest.NewRequest(http.MethodDelete, "/api/countries", body)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestDeleteCountry_MissingBody(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/countries", nil)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCountries_MethodNotAllowed(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/countries", nil)
	rec := httptest.NewRecorder()
	srv.handleCountries(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
