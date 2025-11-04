package textutil

import "testing"

func TestToNormalized(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Remove spaces and hyphens
		{"air conditioning", "airconditioning"},
		{"air-conditioning", "airconditioning"},
		{"Air-Conditioning", "airconditioning"},
		{"  air  conditioning  ", "airconditioning"},
		{"co-operate", "cooperate"},

		// Preserve apostrophes - key test cases
		{"it's", "it's"},                 // contraction: it is
		{"its", "its"},                   // possessive: its
		{"we'll", "we'll"},               // contraction: we will
		{"well", "well"},                 // adverb/noun: well
		{"rock-'n'-roll", "rock'n'roll"}, // special spelling, preserve apostrophe

		// Preserve slashes
		{"cooperate/co-operation", "cooperate/cooperation"},
		{"and/or", "and/or"},

		// Remove underscores
		{"air_conditioning", "airconditioning"},

		// Combined cases
		{"rock 'n' roll", "rock'n'roll"},
		{"20/20", "20/20"},

		// Edge cases
		{"", ""},
		{"   ", ""},
		{"hello", "hello"},
		{"HELLO", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToNormalized(tt.input)
			if result != tt.expected {
				t.Errorf("ToNormalized(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
