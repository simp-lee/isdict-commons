package textutil

import (
	"strings"
)

var normalizationReplacer = strings.NewReplacer(
	"\u2018", "'",
	"\u2019", "'",
	"-", "",
	" ", "",
	"_", "",
)

// ToNormalized converts a word to its normalized form for lookup.
// Removes spaces, hyphens, underscores and converts to lowercase.
// PRESERVES apostrophes and slashes to maintain semantic distinctions.
// Normalizes common Unicode apostrophe variants to ASCII '.
// Examples:
//   - "air conditioning" -> "airconditioning"
//   - "air-conditioning" -> "airconditioning"
//   - "it's" -> "it's" (preserves apostrophe, distinct from "its")
//   - "we'll" -> "we'll" (preserves apostrophe, distinct from "well")
//   - "rock-'n'-roll" -> "rock'n'roll"
//   - "cooperate/co-operation" -> "cooperate/cooperation"
func ToNormalized(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return normalizationReplacer.Replace(s)
}
