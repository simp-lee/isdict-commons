package model

import "testing"

func TestGetPOSName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "unknown"},
		{1, "noun"},
		{2, "verb"},
		{3, "adjective"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		result := GetPOSName(tt.code)
		if result != tt.expected {
			t.Errorf("GetPOSName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestGetAccentName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "unknown"},
		{1, "british"},
		{2, "american"},
		{3, "australian"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		result := GetAccentName(tt.code)
		if result != tt.expected {
			t.Errorf("GetAccentName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestParsePOS(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		ok       bool
	}{
		{"noun", 1, true},
		{"verb", 2, true},
		{"adjective", 3, true},
		{"invalid", 0, false},
	}

	for _, tt := range tests {
		code, ok := ParsePOS(tt.name)
		if ok != tt.ok {
			t.Errorf("ParsePOS(%s) ok = %v; want %v", tt.name, ok, tt.ok)
		}
		if ok && code != tt.expected {
			t.Errorf("ParsePOS(%s) = %d; want %d", tt.name, code, tt.expected)
		}
	}
}

func TestParseAccent(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		ok       bool
	}{
		{"british", 1, true},
		{"american", 2, true},
		{"australian", 3, true},
		{"invalid", 0, false},
	}

	for _, tt := range tests {
		code, ok := ParseAccent(tt.name)
		if ok != tt.ok {
			t.Errorf("ParseAccent(%s) ok = %v; want %v", tt.name, ok, tt.ok)
		}
		if ok && code != tt.expected {
			t.Errorf("ParseAccent(%s) = %d; want %d", tt.name, code, tt.expected)
		}
	}
}

func TestGetFormTypeName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{1, "past"},
		{2, "past_participle"},
		{5, "plural"},
		{999, ""},
	}

	for _, tt := range tests {
		result := GetFormTypeName(tt.code)
		if result != tt.expected {
			t.Errorf("GetFormTypeName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestGetVariantKindName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{1, "form"},
		{2, "alias"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		result := GetVariantKindName(tt.code)
		if result != tt.expected {
			t.Errorf("GetVariantKindName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestGetOxfordLevelName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, ""},
		{1, "Oxford 3000"},
		{2, "Oxford 5000"},
		{999, ""},
	}

	for _, tt := range tests {
		result := GetOxfordLevelName(tt.code)
		if result != tt.expected {
			t.Errorf("GetOxfordLevelName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestOxfordLevelFromString(t *testing.T) {
	tests := []struct {
		source   string
		expected int
	}{
		{"oxford_3000", 1},
		{"oxford_5000", 2},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := OxfordLevelFromString(tt.source)
		if result != tt.expected {
			t.Errorf("OxfordLevelFromString(%s) = %d; want %d", tt.source, result, tt.expected)
		}
	}
}
