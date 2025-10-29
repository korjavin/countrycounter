package main

import (
	"sort"
)

// CountrySuggestion represents a suggested country with its priority score
type CountrySuggestion struct {
	Name          string
	Score         float64
	Priority      int // 1, 2, or 3
	Reason        string
	NeighborCount int
}

// GetCountrySuggestions returns a list of suggested countries based on visited countries
// Priority 1: Countries surrounded by visited countries (filling holes)
// Priority 2: Countries connected to most-visited regions (easy to reach)
// Priority 3: Large countries (good area coverage)
func GetCountrySuggestions(visitedCountries []string, maxSuggestions int) []CountrySuggestion {
	if len(visitedCountries) == 0 {
		return getDefaultSuggestions(maxSuggestions)
	}

	visitedSet := make(map[string]bool)
	for _, country := range visitedCountries {
		visitedSet[country] = true
	}

	var suggestions []CountrySuggestion

	// Priority 1: Find countries surrounded by visited countries
	priority1 := findSurroundedCountries(visitedSet)
	suggestions = append(suggestions, priority1...)

	// Priority 2: Find countries easily reachable (connected to many visited countries)
	priority2 := findEasyToReachCountries(visitedSet)
	suggestions = append(suggestions, priority2...)

	// Priority 3: Find large unvisited countries
	priority3 := findLargeCountries(visitedSet)
	suggestions = append(suggestions, priority3...)

	// Sort by priority (1 > 2 > 3), then by score within same priority
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Priority != suggestions[j].Priority {
			return suggestions[i].Priority < suggestions[j].Priority
		}
		return suggestions[i].Score > suggestions[j].Score
	})

	// Remove duplicates (a country might appear in multiple priority lists)
	seen := make(map[string]bool)
	var unique []CountrySuggestion
	for _, s := range suggestions {
		if !seen[s.Name] {
			seen[s.Name] = true
			unique = append(unique, s)
		}
	}

	// Limit results
	if len(unique) > maxSuggestions {
		unique = unique[:maxSuggestions]
	}

	return unique
}

// findSurroundedCountries finds countries that are bordered by many visited countries
// The more visited neighbors, the more "surrounded" it is
func findSurroundedCountries(visitedSet map[string]bool) []CountrySuggestion {
	var suggestions []CountrySuggestion

	for country, neighbors := range CountryBorders {
		if visitedSet[country] || len(neighbors) == 0 {
			continue
		}

		visitedNeighborCount := 0
		for _, neighbor := range neighbors {
			if visitedSet[neighbor] {
				visitedNeighborCount++
			}
		}

		// Only suggest if at least 50% of neighbors are visited
		surroundedRatio := float64(visitedNeighborCount) / float64(len(neighbors))
		if surroundedRatio >= 0.5 {
			score := float64(visitedNeighborCount) * surroundedRatio * 100
			reason := "Fills a gap in your map"
			if surroundedRatio == 1.0 {
				reason = "Completely surrounded by visited countries!"
			}
			suggestions = append(suggestions, CountrySuggestion{
				Name:          country,
				Score:         score,
				Priority:      1,
				Reason:        reason,
				NeighborCount: visitedNeighborCount,
			})
		}
	}

	return suggestions
}

// findEasyToReachCountries finds unvisited countries that border visited regions
// Countries with more visited neighbors are easier to reach
func findEasyToReachCountries(visitedSet map[string]bool) []CountrySuggestion {
	var suggestions []CountrySuggestion

	for country, neighbors := range CountryBorders {
		if visitedSet[country] {
			continue
		}

		visitedNeighborCount := 0
		for _, neighbor := range neighbors {
			if visitedSet[neighbor] {
				visitedNeighborCount++
			}
		}

		// Must have at least 1 visited neighbor but less than 50% (otherwise it's priority 1)
		if visitedNeighborCount > 0 && len(neighbors) > 0 {
			surroundedRatio := float64(visitedNeighborCount) / float64(len(neighbors))
			if surroundedRatio < 0.5 {
				// Score based on number of visited neighbors (more = easier to reach)
				score := float64(visitedNeighborCount) * 10
				reason := "Easy to reach from your visited countries"
				suggestions = append(suggestions, CountrySuggestion{
					Name:          country,
					Score:         score,
					Priority:      2,
					Reason:        reason,
					NeighborCount: visitedNeighborCount,
				})
			}
		}
	}

	return suggestions
}

// findLargeCountries finds large unvisited countries for good map coverage
func findLargeCountries(visitedSet map[string]bool) []CountrySuggestion {
	var suggestions []CountrySuggestion

	for country, area := range CountryAreas {
		if visitedSet[country] {
			continue
		}

		// Only suggest countries larger than 100,000 km²
		if area > 100000 {
			score := area / 1000 // Normalize score
			reason := "Large country for better map coverage"
			suggestions = append(suggestions, CountrySuggestion{
				Name:     country,
				Score:    score,
				Priority: 3,
				Reason:   reason,
			})
		}
	}

	return suggestions
}

// getDefaultSuggestions returns interesting countries to visit when user has no history
func getDefaultSuggestions(maxSuggestions int) []CountrySuggestion {
	popularCountries := []string{
		"France", "Italy", "Spain", "United Kingdom", "Germany",
		"Japan", "United States", "Thailand", "Australia", "Canada",
		"Mexico", "Brazil", "China", "India", "South Africa",
	}

	var suggestions []CountrySuggestion
	for i, country := range popularCountries {
		if i >= maxSuggestions {
			break
		}
		area := CountryAreas[country]
		suggestions = append(suggestions, CountrySuggestion{
			Name:     country,
			Score:    float64(len(popularCountries) - i),
			Priority: 3,
			Reason:   "Popular destination to start your journey",
		})
		_ = area // Keep area for potential use
	}

	return suggestions
}

// FormatSuggestions formats the suggestions into a human-readable message
func FormatSuggestions(suggestions []CountrySuggestion) string {
	if len(suggestions) == 0 {
		return "No suggestions available. You might have visited all reachable countries!"
	}

	message := "Here are some countries you should visit next:\n\n"

	for i, s := range suggestions {
		emoji := ""
		switch s.Priority {
		case 1:
			emoji = "⭐" // Highest priority - filling gaps
		case 2:
			emoji = "✈️" // Medium priority - easy to reach
		case 3:
			emoji = "🗺️" // Lower priority - large coverage
		}

		message += emoji + " " + s.Name

		if s.NeighborCount > 0 {
			message += " (" + formatNeighborCount(s.NeighborCount) + ")"
		}

		message += "\n   " + s.Reason + "\n"

		if i < len(suggestions)-1 {
			message += "\n"
		}
	}

	return message
}

func formatNeighborCount(count int) string {
	if count == 1 {
		return "1 visited neighbor"
	}
	return string(rune('0'+count)) + " visited neighbors"
}
