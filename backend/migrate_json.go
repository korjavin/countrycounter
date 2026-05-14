package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/korjavin/countrycounter/backend/store"
)

// MaybeImportJSON seeds an empty visits table from a legacy data.json file
// (map[user_id][]country_name).
//
// Behavior:
//   - If the visits table is non-empty, this is a no-op and returns (0, nil).
//   - If jsonPath does not exist, this is a no-op and returns (0, nil) —
//     a green-field deployment is not an error.
//   - If the file exists but is malformed, or if a repo write fails mid-import,
//     the error is returned so the caller can refuse to start with partial
//     state.
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

	imported := 0
	for userID, countries := range data {
		for _, country := range countries {
			if err := repo.Add(userID, country); err != nil {
				return imported, fmt.Errorf("import (user %d, country %q): %w", userID, country, err)
			}
			imported++
		}
	}
	return imported, nil
}
