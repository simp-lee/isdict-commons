package model

import (
	"encoding/json"
	"testing"
)

func TestWordAnnotationsEmbedding(t *testing.T) {
	// Test that embedded WordAnnotations fields appear at the top level in JSON
	resp := WordResponse{
		ID:       123,
		Headword: "test",
		WordAnnotations: WordAnnotations{
			CEFRLevel:      3,
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

	// Check that annotation fields are at the top level (not nested)
	if _, ok := result["cefr_level"]; !ok {
		t.Error("cefr_level should be at top level")
	}
	if _, ok := result["cet_level"]; !ok {
		t.Error("cet_level should be at top level")
	}
	if _, ok := result["oxford_level"]; !ok {
		t.Error("oxford_level should be at top level")
	}
	if _, ok := result["school_level"]; !ok {
		t.Error("school_level should be at top level")
	}
	if _, ok := result["frequency_rank"]; !ok {
		t.Error("frequency_rank should be at top level")
	}
	if _, ok := result["frequency_count"]; !ok {
		t.Error("frequency_count should be at top level")
	}
	if _, ok := result["collins_stars"]; !ok {
		t.Error("collins_stars should be at top level")
	}
	if _, ok := result["translation_zh"]; !ok {
		t.Error("translation_zh should be at top level")
	}

	// Verify values
	if int(result["cefr_level"].(float64)) != 3 {
		t.Errorf("Expected cefr_level=3, got %v", result["cefr_level"])
	}
	if int(result["cet_level"].(float64)) != 4 {
		t.Errorf("Expected cet_level=4, got %v", result["cet_level"])
	}
	if int(result["oxford_level"].(float64)) != 1 {
		t.Errorf("Expected oxford_level=1, got %v", result["oxford_level"])
	}
	if int(result["collins_stars"].(float64)) != 3 {
		t.Errorf("Expected collins_stars=3, got %v", result["collins_stars"])
	}
	if result["translation_zh"].(string) != "测试；考试" {
		t.Errorf("Expected translation_zh='测试；考试', got %v", result["translation_zh"])
	}

	t.Logf("JSON output: %s", string(jsonData))
}

func TestSearchResultResponseEmbedding(t *testing.T) {
	resp := SearchResultResponse{
		ID:       456,
		Headword: "example",
		POS:      []string{"noun", "verb"},
		WordAnnotations: WordAnnotations{
			CEFRLevel:     2,
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
			CEFRLevel:     1,
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
