package model

import "testing"

func TestFormTypeMappingAliases(t *testing.T) {
	tests := []struct {
		name     string
		variants []string
		expected int
	}{
		{name: "past aliases", variants: []string{"past", "past_tense", "preterite"}, expected: 1},
		{name: "past participle aliases", variants: []string{"past_participle", "past_part"}, expected: 2},
		{name: "present third aliases", variants: []string{"present_3rd", "3rd_person_singular", "third_person_singular"}, expected: 3},
		{name: "gerund aliases", variants: []string{"gerund", "present_participle", "ing_form"}, expected: 4},
		{name: "possessive aliases", variants: []string{"possessive", "genitive"}, expected: 8},
		{name: "infinitive aliases", variants: []string{"infinitive", "to_infinitive"}, expected: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, variant := range tt.variants {
				code, ok := FormTypeMapping[variant]
				if !ok {
					t.Fatalf("FormTypeMapping[%q] missing", variant)
				}
				if code != tt.expected {
					t.Fatalf("FormTypeMapping[%q] = %d; want %d", variant, code, tt.expected)
				}
			}
		})
	}
}

func TestVariantKindConstantsMatchSchema(t *testing.T) {
	if VariantForm != 1 {
		t.Fatalf("VariantForm = %d; want 1", VariantForm)
	}
	if VariantAlias != 2 {
		t.Fatalf("VariantAlias = %d; want 2", VariantAlias)
	}
}

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

func TestGetCEFRLevelName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, ""},
		{1, "A1"},
		{2, "A2"},
		{3, "B1"},
		{4, "B2"},
		{5, "C1"},
		{6, "C2"},
		{-1, ""},
		{999, ""},
	}

	for _, tt := range tests {
		result := GetCEFRLevelName(tt.code)
		if result != tt.expected {
			t.Errorf("GetCEFRLevelName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestGetSchoolLevelName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "unknown"},
		{1, "初中"},
		{2, "高中"},
		{3, "大学"},
		{-1, "unknown"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		result := GetSchoolLevelName(tt.code)
		if result != tt.expected {
			t.Errorf("GetSchoolLevelName(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestParseCEFRLevel(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		ok       bool
	}{
		{"A1", 1, true},
		{"A2", 2, true},
		{"B1", 3, true},
		{"B2", 4, true},
		{"C1", 5, true},
		{"C2", 6, true},
		{"", 0, true},
		{"X1", 0, false},
		{"a1", 0, false},
		{"invalid", 0, false},
	}

	for _, tt := range tests {
		code, ok := ParseCEFRLevel(tt.name)
		if ok != tt.ok {
			t.Errorf("ParseCEFRLevel(%s) ok = %v; want %v", tt.name, ok, tt.ok)
		}
		if ok && code != tt.expected {
			t.Errorf("ParseCEFRLevel(%s) = %d; want %d", tt.name, code, tt.expected)
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
