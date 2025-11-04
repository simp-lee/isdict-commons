package textutil

import (
	"strings"
)

// ToNormalized converts a word to its normalized form for lookup.
// Removes spaces, hyphens, underscores and converts to lowercase.
// PRESERVES apostrophes and slashes to maintain semantic distinctions.
// Examples:
//   - "air conditioning" -> "airconditioning"
//   - "air-conditioning" -> "airconditioning"
//   - "it's" -> "it's" (preserves apostrophe, distinct from "its")
//   - "we'll" -> "we'll" (preserves apostrophe, distinct from "well")
//   - "rock-'n'-roll" -> "rock'n'roll"
//   - "cooperate/co-operation" -> "cooperate/cooperation"
func ToNormalized(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Only remove spaces, hyphens, and underscores
	// Preserve apostrophes (') to distinguish contractions from possessives
	// Preserve slashes (/) to maintain alternative forms
	replacers := []string{"-", " ", "_"}
	for _, r := range replacers {
		s = strings.ReplaceAll(s, r, "")
	}
	return s
}
