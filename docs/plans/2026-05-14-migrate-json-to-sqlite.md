# Migrate JSON storage to SQLite with Goose migrations

## Overview

Replace the current `backend/data.json` file-based storage with SQLite, managed by `pressly/goose/v3` migrations. The mapping `map[int64][]string` (Telegram user ID → visited country names) becomes a single `visits` table. On startup, if the DB is empty and a `data.json` file exists, the backend auto-imports it into the fresh DB — no manual step, no flag required. The pattern mirrors `../medicationtrackerbot`: pure-Go driver (`modernc.org/sqlite`), embedded SQL migrations auto-applied on startup, plain `database/sql` repository, WAL journal with `busy_timeout=5000` and `MaxOpenConns=1`.

Benefits: durable transactional writes (no more re-marshaling the whole map on every change), schema evolution via versioned migrations, removal of the global `sync.Mutex` (SQLite handles serialization), and proper unit-testable storage isolated from HTTP/bot handlers.

## Context (from discovery)

- Files involved:
  - `backend/main.go` — owns the global `UserData map[int64][]string`, `loadData()`/`saveData()`, all HTTP handlers (`getCountries`/`addCountry`/`deleteCountry`), and the Telegram bot loop that mutates `UserData` directly.
  - `backend/data.json` — current persisted store.
  - `backend/go.mod` — Go 1.24; no DB deps yet.
  - `backend/suggestions_test.go` — only existing test file; covers suggestion algorithm, not storage.
- Reference patterns from `../medicationtrackerbot`:
  - `internal/store/db/db.go` — `Open(path)` with WAL + busy_timeout + `MaxOpenConns(1)`.
  - `internal/store/store.go` — `//go:embed migrations/*.sql` and goose auto-migrate on startup.
  - `internal/store/migrations/*.sql` — `-- +goose Up` / `-- +goose Down` directives.
  - `internal/store/<domain>/repo.go` — plain `database/sql`, hand-written parameterized SQL.
- Data model is trivial: one entity (`visits`), composite key (user_id, country_name).

## Development Approach

- **Testing approach**: Regular (code first, then tests) — matches the style in `docs/plans/2026-05-01-improve-map-image-quality.md`
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change: `cd backend && go test ./...`
- Maintain backward compatibility for the HTTP API surface (request/response shapes do not change)

## Testing Strategy

- **Unit tests**: required for every task. Storage tests open an in-memory SQLite DB (`":memory:"`) and run goose migrations against it; handler tests use `httptest.NewRecorder` against a repo wired to an in-memory DB.
- **E2E tests**: project has Playwright tests under `e2e/`. The migration is backend-only and the HTTP contract is unchanged, so existing e2e tests should still pass — running them is part of Task 8 verification.

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code, schema, migrations, tests, docs
- **Post-Completion** (no checkboxes): operational steps that require human action — running the import on production data, swapping the data volume in deployment, deleting the legacy `data.json`

## Implementation Steps

### Task 1: Add SQLite + goose dependencies and DB connection layer

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`
- Create: `backend/store/db.go`

- [x] Add `modernc.org/sqlite` and `github.com/pressly/goose/v3` via `go get` in `backend/`
- [x] Create `backend/store/db.go` with an `Open(path string) (*sql.DB, error)` that: opens `sqlite` driver, pings, runs `PRAGMA journal_mode=WAL`, `PRAGMA busy_timeout = 5000`, and `db.SetMaxOpenConns(1)` — mirroring `medicationtrackerbot/internal/store/db/db.go`
- [x] Add a blank import `_ "modernc.org/sqlite"` in `backend/store/db.go`
- [x] Write `backend/store/db_test.go` covering: Open succeeds for `:memory:`; Open returns error for an unwritable path; the returned DB responds to Ping
- [x] Run `cd backend && go test ./...` — must pass before task 2

### Task 2: Add initial migration for the visits table

**Files:**
- Create: `backend/store/migrations/001_init.sql`
- Modify: `backend/store/db.go` (add `Migrate` function and embed FS)

- [x] Create `backend/store/migrations/001_init.sql` with:
  - `-- +goose Up`: `CREATE TABLE visits (user_id INTEGER NOT NULL, country_name TEXT NOT NULL, added_at DATETIME DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, country_name));` plus `CREATE INDEX idx_visits_user_id ON visits(user_id);`
  - `-- +goose Down`: `DROP INDEX idx_visits_user_id; DROP TABLE visits;`
- [x] In `backend/store/db.go` add `//go:embed migrations/*.sql` declaring `embedMigrations embed.FS` and a `Migrate(db *sql.DB) error` that calls `goose.SetBaseFS(embedMigrations)`, `goose.SetDialect("sqlite3")`, and `goose.Up(db, "migrations")`
- [x] Write `backend/store/migrate_test.go` covering: running Migrate on a fresh in-memory DB succeeds; the `visits` table exists afterward (query `sqlite_master`); calling Migrate twice is idempotent
- [x] Run `cd backend && go test ./...` — must pass before task 3

### Task 3: Implement the visits repository

**Files:**
- Create: `backend/store/visits.go`
- Create: `backend/store/visits_test.go`

- [x] Define `type VisitsRepo struct { db *sql.DB }` and `func NewVisitsRepo(db *sql.DB) *VisitsRepo`
- [x] Implement `List(userID int64) ([]string, error)` — `SELECT country_name FROM visits WHERE user_id = ? ORDER BY added_at` — returns empty slice (not nil) when no rows
- [x] Implement `Add(userID int64, country string) error` — `INSERT OR IGNORE INTO visits (user_id, country_name) VALUES (?, ?)` — idempotent (matches current "already visited" UX in the bot)
- [x] Implement `Delete(userID int64, country string) (bool, error)` — `DELETE FROM visits WHERE user_id = ? AND country_name = ?`; return `true` if a row was deleted via `RowsAffected`
- [x] Implement `Has(userID int64, country string) (bool, error)` — `SELECT 1 FROM visits WHERE user_id = ? AND country_name = ? LIMIT 1`
- [x] Write table-driven tests for each method using a per-test in-memory DB + Migrate; cover empty user, multiple countries, duplicate Add (no-op), Delete of missing row (returns false), unicode country names
- [x] Run `cd backend && go test ./...` — must pass before task 4

### Task 4: Refactor HTTP handlers to use the repository

**Files:**
- Modify: `backend/main.go`
- Create: `backend/handlers_test.go` (if not adding to existing test file)

- [x] Remove the global `var UserData map[int64][]string` and `var mutex` from `backend/main.go`
- [x] Remove `loadData()` and `saveData()` from `backend/main.go`
- [x] Introduce a `type server struct { repo *store.VisitsRepo }` and convert `handleCountries`, `getCountries`, `addCountry`, `deleteCountry` into methods on `*server`
- [x] In `main()`, open the DB via `store.Open(dbPath)`, call `store.Migrate(db)`, construct the `server`, and register `srv.handleCountries` on the mux — `dbPath` from env `DB_PATH` with default `backend/data.db`
- [x] Update `getCountries` to call `repo.List(userID)`
- [x] Update `addCountry` to call `repo.Add(userID, country)` and return 201
- [x] Update `deleteCountry` to call `repo.Delete(userID, country)`; return 404 if the bool from Delete is false
- [x] Write `backend/handlers_test.go` covering each handler with `httptest.NewRecorder` against an in-memory repo: GET returns `[]` for new user, GET returns recorded countries, POST creates, POST same country twice succeeds (idempotent), DELETE removes, DELETE of missing returns 404, bad userId → 400, missing body → 400
- [x] Run `cd backend && go test ./...` — must pass before task 5

### Task 5: Refactor the Telegram bot to use the repository

**Files:**
- Modify: `backend/main.go` (the `startTelegramBot` function)

- [x] Change `startTelegramBot` to accept `*store.VisitsRepo` (or move it onto `*server`) and call it from `main` accordingly
- [x] Replace the location-handler block (currently `mutex.Lock` + map read + `UserData[userID] = append(...)` + `saveData()`) with `repo.Has(...)` → `repo.Add(...)` → `repo.List(...)` to recompute the count for the reply
- [x] Replace `/map` handler's `UserData[userID]` read with `repo.List(userID)`
- [x] Replace `/list` handler's `UserData[userID]` read with `repo.List(userID)`
- [x] Replace `/suggest` handler's `UserData[userID]` read with `repo.List(userID)`
- [x] Extract a small helper (e.g. `handleLocation(repo, userID, lat, lng) (replyText string, err error)`) so the location flow is unit-testable without a real Telegram client
- [x] Write tests for `handleLocation` covering: new country added → reply mentions count; same country twice → "already added" path; geocoding failure → error reply
- [x] Run `cd backend && go test ./...` — must pass before task 6

### Task 6: Auto-import data.json on first startup

**Files:**
- Create: `backend/migrate_json.go`
- Create: `backend/migrate_json_test.go`
- Modify: `backend/main.go` (wire the auto-import into startup)

- [x] In `backend/migrate_json.go` implement `func MaybeImportJSON(repo *store.VisitsRepo, jsonPath string) (importedRows int, err error)` that: checks if the `visits` table is empty (`SELECT 1 FROM visits LIMIT 1` or a count query); if non-empty, returns (0, nil); if empty, checks whether `jsonPath` exists; if missing, returns (0, nil); if present, reads the file, unmarshals into `map[int64][]string`, iterates and calls `repo.Add` for each (user, country) pair, returns total inserted count
- [x] Expose a helper on `*store.VisitsRepo` like `IsEmpty() (bool, error)` to support the auto-detection (keeps the JSON layer agnostic of schema details)
- [x] Returns clear error only for *unexpected* failures (corrupt JSON, mid-import DB error). Missing JSON is treated as "nothing to do", not an error — this is the green-field deployment case
- [x] In `backend/main.go`, after `store.Open` + `store.Migrate`, call `MaybeImportJSON(repo, "backend/data.json")` (the path used today by `loadData`); log either `Auto-imported N rows from data.json` or `No data.json found or DB already populated — skipping auto-import`
- [x] On import error, log a clear message and `log.Fatalf` — refusing to start with partial state is safer than continuing with an unknown DB
- [x] Write `backend/migrate_json_test.go` covering: empty DB + valid JSON (3 users / 7 visits) → 7 imported, all queryable via `repo.List`; empty DB + no JSON file → (0, nil) no error; empty DB + empty JSON `{}` → (0, nil); empty DB + malformed JSON → error; populated DB + JSON exists → (0, nil) and existing rows are untouched (i.e. no overwrite, no merge); idempotency: second call after a successful import returns (0, nil)
- [x] Run `cd backend && go test ./...` — must pass before task 7

### Task 7: Update deployment & build configuration

**Files:**
- Modify: `Dockerfile`, `docker-compose.yml`, `docker-compose.prod.yml`

- [x] Ensure the Docker image includes the new SQLite DB file location on a persistent volume (replace or sit alongside the existing `data.json` volume mapping). Default in-container path: `/app/backend/data.db`
- [x] In `docker-compose.yml` and `docker-compose.prod.yml`, add a volume mount for the DB file/directory (a named volume or bind mount), preserving any existing `data.json` mount during the cutover window
- [x] Set env var `DB_PATH=/app/backend/data.db` in compose files for clarity
- [x] No production code changes here, so no new test code — but: run `docker compose config` to validate the YAML is well-formed
- [x] Run `cd backend && go test ./...` — must pass before task 8

### Task 8: Verify acceptance criteria

- [ ] Verify all requirements from Overview are implemented (SQLite storage, goose migrations, auto-import on first startup, mutex removed)
- [ ] Verify edge cases are handled (empty user list, duplicate Add, Delete of missing, malformed JSON in auto-import, populated-DB-skips-import, missing-JSON-not-an-error, idempotent re-migration)
- [ ] Manually verify the auto-import flow end-to-end: with a real `data.json` present and no `data.db`, start the server, observe the log line, query a known user via `GET /api/countries?userId=...` and confirm the data matches the JSON
- [ ] Run full test suite: `cd backend && go test ./...`
- [ ] Run e2e tests if reachable: `npx playwright test` from repo root (HTTP API surface unchanged, expected to pass)
- [ ] Run linter: `cd backend && go vet ./...` — all issues must be fixed
- [ ] Verify with `go test -cover ./...` that storage and handler packages have meaningful coverage (target 80%+ for new code)
- [ ] Grep for any remaining references to `UserData`, `loadData`, `saveData`, `data.json` in `backend/` — none should remain outside the migrator and tests

### Task 9: Update documentation

**Files:**
- Modify: `readme.md`, `DEPLOYMENT.md`

- [ ] Update `readme.md` storage section: replace any mention of `data.json` with SQLite (`backend/data.db`); note that `data.json`, if still present on first start, is auto-imported once and then ignored
- [ ] Update `DEPLOYMENT.md` upgrade procedure: keep `data.json` mounted alongside the new DB volume for one release so the first start of the new image can auto-import it; after a confidence window, drop the `data.json` mount
- [ ] No code tests for docs, but verify the documented flow actually works by starting the server locally with a copy of `data.json` present and a fresh `data.db` path

*Note: ralphex automatically moves completed plans to `docs/plans/completed/`*

## Technical Details

**Schema (migration 001):**
```sql
-- +goose Up
CREATE TABLE visits (
    user_id      INTEGER  NOT NULL,
    country_name TEXT     NOT NULL,
    added_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, country_name)
);
CREATE INDEX idx_visits_user_id ON visits(user_id);

-- +goose Down
DROP INDEX idx_visits_user_id;
DROP TABLE visits;
```

**Why composite PK on `(user_id, country_name)`**: matches the implicit invariant of the current code (a user can't have the same country twice — the bot's "already visited" check enforces this in memory). Makes `INSERT OR IGNORE` the natural idempotent operation for both writes and the JSON import.

**Repository surface:**
```go
type VisitsRepo struct{ db *sql.DB }
func NewVisitsRepo(db *sql.DB) *VisitsRepo
func (r *VisitsRepo) List(userID int64) ([]string, error)
func (r *VisitsRepo) Add(userID int64, country string) error           // idempotent
func (r *VisitsRepo) Delete(userID int64, country string) (bool, error) // bool: did a row get removed?
func (r *VisitsRepo) Has(userID int64, country string) (bool, error)
```

**Package layout (`backend/`):**
```
backend/
  main.go                    (HTTP server + bot, now talks to repo)
  migrate_json.go            (one-shot importer)
  migrate_json_test.go
  handlers_test.go
  suggestions.go             (unchanged)
  suggestions_test.go        (unchanged)
  country_data.go            (unchanged)
  iso_mapping.go             (unchanged)
  store/
    db.go                    (Open + Migrate + embed)
    db_test.go
    migrate_test.go
    visits.go
    visits_test.go
    migrations/
      001_init.sql
```

**Concurrency:** the existing `sync.Mutex` is removed entirely. `MaxOpenConns=1` + `busy_timeout=5000ms` serializes writes at the SQL layer; reads still go through the same single connection. This matches the medicationtrackerbot strategy and is fine for the expected load of a personal Telegram bot.

**CLI surface added to `backend` binary:**
- `-db <path>` — SQLite path; default `backend/data.db`; also honored via `DB_PATH` env

**Auto-import behavior on startup (no flag needed):**
1. Open DB at `-db` path (creates the file if missing)
2. Run goose migrations (creates the `visits` table on first run)
3. If `visits` is empty AND `backend/data.json` exists → import every (user, country) pair from the JSON and log `Auto-imported N rows from data.json`
4. If `visits` is non-empty OR `data.json` is missing → log skip reason and continue normally
5. Any import *failure* (malformed JSON, DB error mid-import) is fatal — better than booting with partial state

## Post-Completion

*Items requiring manual intervention or external systems — informational only*

**Production cutover:**
1. Deploy the new image with the SQLite volume mounted *and* the existing `data.json` mounted at its current path
2. On first start, the backend auto-detects the empty DB + present `data.json` and imports — observe the `Auto-imported N rows from data.json` log line
3. Verify in the UI/bot that a known user's country list looks right
4. After a confidence window (1 week), remove the `data.json` mount and delete the file from the host
5. Subsequent restarts simply see a populated DB and skip the import path

**Manual verification:**
- Smoke test: open the web UI for a real user, confirm visited countries render
- Smoke test: send `/list` and `/map` to the bot for a real user
- Smoke test: send a Telegram location and confirm the country gets added
- Backup: take a copy of `data.db` after the import (sqlite is a single file — `cp` works)

**External system updates:**
- None — this change is internal to `backend/`. Frontend and e2e tests are unaffected because the HTTP API contract is unchanged.
