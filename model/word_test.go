package model

import (
	"encoding/json"
	"testing"
)

type topLevelFieldCheck struct {
	key   string
	value any
}

func TestWordAnnotationsEmbedding(t *testing.T) {
	// Test that embedded WordAnnotations fields appear at the top level in JSON
	resp := WordResponse{
		ID:       123,
		Headword: "test",
		WordAnnotations: WordAnnotations{
			CEFRLevel:      "B1",
			CETLevel:       4,
			OxfordLevel:    1,
			SchoolLevel:    2,
			FrequencyRank:  100,
			FrequencyCount: 5000,
			CollinsStars:   3,
			TranslationZH:  "测试；考试",
		},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify JSON structure - embedded fields should be at top level
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	assertTopLevelFields(t, result,
		topLevelFieldCheck{key: "cefr_level", value: "B1"},
		topLevelFieldCheck{key: "cet_level", value: 4},
		topLevelFieldCheck{key: "oxford_level", value: 1},
		topLevelFieldCheck{key: "school_level", value: 2},
		topLevelFieldCheck{key: "frequency_rank", value: 100},
		topLevelFieldCheck{key: "frequency_count", value: 5000},
		topLevelFieldCheck{key: "collins_stars", value: 3},
		topLevelFieldCheck{key: "translation_zh", value: "测试；考试"},
	)

	t.Logf("JSON output: %s", string(jsonData))
}

func TestSearchResultResponseEmbedding(t *testing.T) {
	resp := SearchResultResponse{
		ID:       456,
		Headword: "example",
		POS:      []string{"noun", "verb"},
		WordAnnotations: WordAnnotations{
			CEFRLevel:     "A2",
			FrequencyRank: 500,
		},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// All annotation fields should be at top level
	if _, ok := result["cefr_level"]; !ok {
		t.Error("cefr_level should be at top level")
	}
	if _, ok := result["frequency_rank"]; !ok {
		t.Error("frequency_rank should be at top level")
	}

	t.Logf("JSON output: %s", string(jsonData))
}

func TestSuggestResponseEmbedding(t *testing.T) {
	resp := SuggestResponse{
		Headword: "word",
		WordAnnotations: WordAnnotations{
			CEFRLevel:     "A1",
			FrequencyRank: 50,
			OxfordLevel:   1,
		},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Annotation fields should be at top level
	if _, ok := result["cefr_level"]; !ok {
		t.Error("cefr_level should be at top level")
	}
	if _, ok := result["frequency_rank"]; !ok {
		t.Error("frequency_rank should be at top level")
	}
	if _, ok := result["oxford_level"]; !ok {
		t.Error("oxford_level should be at top level")
	}

	t.Logf("JSON output: %s", string(jsonData))
}

func TestSenseResponseCEFRLevelSerialization(t *testing.T) {
	resp := SenseResponse{
		SenseID:      42,
		POS:          "verb",
		CEFRLevel:    "A2",
		CEFRSource:   "oxford",
		OxfordLevel:  1,
		DefinitionEN: "to examine something",
		DefinitionZH: "检查；审视",
		SenseOrder:   2,
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, ok := result["cefr_level"]; !ok {
		t.Fatal("cefr_level should be present")
	}
	if _, ok := result["sense_id"]; !ok {
		t.Fatal("sense_id should be present")
	}
	if result["cefr_level"].(string) != "A2" {
		t.Errorf("Expected cefr_level=A2, got %v", result["cefr_level"])
	}
	if result["pos"].(string) != "verb" {
		t.Errorf("Expected pos=verb, got %v", result["pos"])
	}
	if int(result["sense_id"].(float64)) != 42 {
		t.Errorf("Expected sense_id=42, got %v", result["sense_id"])
	}

	t.Logf("JSON output: %s", string(jsonData))
}

func assertTopLevelFields(t *testing.T, result map[string]interface{}, checks ...topLevelFieldCheck) {
	t.Helper()

	for _, check := range checks {
		actual, ok := result[check.key]
		if !ok {
			t.Errorf("%s should be at top level", check.key)
			continue
		}

		switch expected := check.value.(type) {
		case string:
			if actualString, ok := actual.(string); !ok || actualString != expected {
				t.Errorf("Expected %s=%v, got %v", check.key, expected, actual)
			}
		case int:
			actualNumber, ok := actual.(float64)
			if !ok || int(actualNumber) != expected {
				t.Errorf("Expected %s=%v, got %v", check.key, expected, actual)
			}
		default:
			t.Fatalf("unsupported expected type %T for key %s", check.value, check.key)
		}
	}
}
