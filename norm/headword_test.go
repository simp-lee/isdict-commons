package norm

import "testing"

func TestNormalizeHeadword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "spaces_removed_and_lowercased", input: "Asian American", want: "asianamerican"},
		{name: "ascii_apostrophe_preserved", input: "it's", want: "it's"},
		{name: "typographic_apostrophe_normalized", input: "It’s", want: "it's"},
		{name: "hyphen_removed", input: "co-op", want: "coop"},
		{name: "underscore_removed", input: "ice_cream", want: "icecream"},
		{name: "slash_preserved", input: "and/or", want: "and/or"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizeHeadword(tt.input); got != tt.want {
				t.Fatalf("NormalizeHeadword(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsMultiword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		headword string
		want     bool
	}{
		{name: "space_separated", headword: "ice cream", want: true},
		{name: "hyphenated", headword: "well-known", want: true},
		{name: "single_word", headword: "book", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsMultiword(tt.headword); got != tt.want {
				t.Fatalf("IsMultiword(%q) = %t; want %t", tt.headword, got, tt.want)
			}
		})
	}
}
