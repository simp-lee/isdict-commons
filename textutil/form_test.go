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
		{"it’s", "it's"},                 // smart apostrophe should normalize to ASCII apostrophe
		{"it‘s", "it's"},                 // left single quotation mark should normalize to ASCII apostrophe
		{"its", "its"},                   // possessive: its
		{"we'll", "we'll"},               // contraction: we will
		{"we’ll", "we'll"},               // smart apostrophe should preserve normalized contraction form
		{"well", "well"},                 // adverb/noun: well
		{"rock-'n'-roll", "rock'n'roll"}, // special spelling, preserve apostrophe
		{"rock-’n’-roll", "rock'n'roll"}, // smart apostrophes normalize before punctuation stripping

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

func TestToNormalizedEquivalentApostrophes(t *testing.T) {
	if got, want := ToNormalized("it's"), ToNormalized("it’s"); got != want {
		t.Fatalf("ASCII and Unicode apostrophes should normalize equally: got %q, want %q", got, want)
	}

	if got, want := ToNormalized("we'll"), ToNormalized("we‘ll"); got != want {
		t.Fatalf("ASCII and Unicode apostrophes should normalize equally: got %q, want %q", got, want)
	}
}
