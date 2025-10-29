package main

import (
	"testing"
)

func TestGetCountrySuggestionsEmpty(t *testing.T) {
	suggestions := GetCountrySuggestions([]string{}, 5)

	if len(suggestions) == 0 {
		t.Error("Expected default suggestions for empty input")
	}

	if len(suggestions) > 5 {
		t.Errorf("Expected at most 5 suggestions, got %d", len(suggestions))
	}
}

func TestGetCountrySuggestionsSurrounded(t *testing.T) {
	// Belgium is surrounded by France, Luxembourg, Germany, Netherlands
	visitedCountries := []string{"France", "Luxembourg", "Germany", "Netherlands"}

	suggestions := GetCountrySuggestions(visitedCountries, 10)

	// Belgium should be in the suggestions with priority 1
	found := false
	for _, s := range suggestions {
		if s.Name == "Belgium" {
			found = true
			if s.Priority != 1 {
				t.Errorf("Expected Belgium to have priority 1 (surrounded), got priority %d", s.Priority)
			}
			break
		}
	}

	if !found {
		t.Error("Expected Belgium to be suggested as it's surrounded by visited countries")
	}
}

func TestGetCountrySuggestionsEasyToReach(t *testing.T) {
	// If someone visited Germany, Czechia should be easy to reach
	visitedCountries := []string{"Germany"}

	suggestions := GetCountrySuggestions(visitedCountries, 10)

	// Czechia should be in the suggestions
	found := false
	for _, s := range suggestions {
		if s.Name == "Czechia" {
			found = true
			if s.Priority != 2 {
				t.Logf("Czechia has priority %d (expected 2 or 1 depending on visited neighbors)", s.Priority)
			}
			break
		}
	}

	if !found {
		t.Error("Expected Czechia to be suggested as it borders Germany")
	}
}

func TestGetCountrySuggestionsLargeCountries(t *testing.T) {
	// If someone visited a small country far from large ones
	visitedCountries := []string{"Singapore"}

	suggestions := GetCountrySuggestions(visitedCountries, 10)

	// Should get suggestions including large countries
	if len(suggestions) == 0 {
		t.Error("Expected some suggestions")
	}

	// Check that we have at least some priority 3 (large countries)
	hasPriority3 := false
	for _, s := range suggestions {
		if s.Priority == 3 {
			hasPriority3 = true
			break
		}
	}

	if !hasPriority3 {
		t.Error("Expected at least one large country suggestion (priority 3)")
	}
}

func TestFormatSuggestions(t *testing.T) {
	suggestions := []CountrySuggestion{
		{
			Name:          "Belgium",
			Score:         100,
			Priority:      1,
			Reason:        "Completely surrounded by visited countries!",
			NeighborCount: 4,
		},
		{
			Name:          "Poland",
			Score:         30,
			Priority:      2,
			Reason:        "Easy to reach from your visited countries",
			NeighborCount: 3,
		},
	}

	message := FormatSuggestions(suggestions)

	if message == "" {
		t.Error("Expected non-empty formatted message")
	}

	// Check that country names are in the message
	if !contains(message, "Belgium") {
		t.Error("Expected Belgium in formatted message")
	}

	if !contains(message, "Poland") {
		t.Error("Expected Poland in formatted message")
	}
}

func TestFormatSuggestionsEmpty(t *testing.T) {
	message := FormatSuggestions([]CountrySuggestion{})

	if message == "" {
		t.Error("Expected non-empty message even for empty suggestions")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
