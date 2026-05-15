package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/korjavin/countrycounter/backend/store"
)

// legacyNameRenames maps pre-rename country names to their current canonical
// form. Mirrors migration 002_normalize_legacy_names.sql so a fresh DB seeded
// from a legacy data.json ends up in the same shape as a DB that was migrated
// in place. Without this, the migration runs before the JSON import (see
// main.go) and legacy names from data.json bypass normalization.
var legacyNameRenames = map[string]string{
	"Cape Verde": "Cabo Verde",
	"Palestine":  "Palestine State",
}

// MaybeImportJSON seeds an empty visits table from a legacy data.json file
// (map[user_id][]country_name).
//
// Behavior:
//   - If the visits table is non-empty, this is a no-op and returns (0, nil).
//   - If jsonPath does not exist, this is a no-op and returns (0, nil) —
//     a green-field deployment is not an error.
//   - If the file exists but is malformed, or if the import write fails,
//     the error is returned and the table is left empty (the import runs
//     inside a single transaction that rolls back on failure), so the next
//     start can safely retry once the underlying problem is fixed.
//
// Legacy country names in the JSON (e.g. "Cape Verde", "Palestine") are
// rewritten to their canonical equivalents before insertion so the imported
// rows match what the autocomplete and frontend duplicate check expect.
//
// The importer is intentionally a one-shot: once any row exists, subsequent
// calls short-circuit without touching the file system.
func MaybeImportJSON(repo *store.VisitsRepo, jsonPath string) (int, error) {
	empty, err := repo.IsEmpty()
	if err != nil {
		return 0, fmt.Errorf("check db emptiness: %w", err)
	}
	if !empty {
		return 0, nil
	}

	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read %s: %w", jsonPath, err)
	}

	var data map[int64][]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return 0, fmt.Errorf("parse %s: %w", jsonPath, err)
	}

	for uid, names := range data {
		for i, n := range names {
			if canonical, ok := legacyNameRenames[n]; ok {
				names[i] = canonical
			}
		}
		data[uid] = names
	}

	return repo.ImportPairs(data)
}
