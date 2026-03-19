package model

// WordAnnotations contains all annotation/level fields for a word
// These fields are used across different response types to indicate
// word difficulty, frequency, and categorization levels
type WordAnnotations struct {
	CEFRLevel      string `json:"cefr_level"`               // CEFR level: "A1","A2","B1","B2","C1","C2", or "" (unknown)
	CEFRSource     string `json:"cefr_source,omitempty"`    // CEFR data source: "oxford", "cefrj", "both", or "" (empty if no CEFR data)
	CETLevel       int    `json:"cet_level"`                // China College English Test (4=CET4, 6=CET6, 0=unknown)
	OxfordLevel    int    `json:"oxford_level"`             // Oxford wordlist (1=Oxford3000, 2=Oxford5000, 0=unknown)
	SchoolLevel    int    `json:"school_level"`             // Recommended learning stage for Chinese English learners (0=unknown, 1=初中, 2=高中, 3=大学)
	FrequencyRank  int    `json:"frequency_rank"`           // Word frequency ranking (lower = more common, 0=unknown)
	FrequencyCount int    `json:"frequency_count"`          // Raw frequency count (0=unknown)
	CollinsStars   int    `json:"collins_stars"`            // Collins COBUILD star rating (1-5, 0=unknown)
	TranslationZH  string `json:"translation_zh,omitempty"` // General Chinese translation from ECDICT
}

// WordResponse represents a word with all its details
type WordResponse struct {
	ID              uint                    `json:"id"`
	Headword        string                  `json:"headword"`
	WordAnnotations                         // Embedded: all annotation fields
	QueriedVariant  *QueriedVariantInfo     `json:"queried_variant,omitempty"` // If queried via variant, return variant frequency info
	Pronunciations  []PronunciationResponse `json:"pronunciations,omitempty"`
	Senses          []SenseResponse         `json:"senses,omitempty"`
	Variants        []VariantResponse       `json:"variants,omitempty"`
}

// PronunciationResponse represents pronunciation information
type PronunciationResponse struct {
	Accent    string `json:"accent"`
	IPA       string `json:"ipa"`
	IsPrimary bool   `json:"is_primary"`
}

// SenseResponse represents a word sense with examples
type SenseResponse struct {
	SenseID      uint              `json:"sense_id"`
	POS          string            `json:"pos"`
	CEFRLevel    string            `json:"cefr_level"`
	CEFRSource   string            `json:"cefr_source,omitempty"` // CEFR data source
	OxfordLevel  int               `json:"oxford_level"`          // Sense-level Oxford annotation
	DefinitionEN string            `json:"definition_en,omitempty"`
	DefinitionZH string            `json:"definition_zh,omitempty"`
	SenseOrder   int               `json:"sense_order"`
	Examples     []ExampleResponse `json:"examples,omitempty"`
}

// ExampleResponse represents an example sentence
type ExampleResponse struct {
	ExampleID    uint   `json:"example_id"`
	SentenceEN   string `json:"sentence_en,omitempty"`
	SentenceZH   string `json:"sentence_zh,omitempty"`
	ExampleOrder int    `json:"example_order"`
}

// VariantResponse represents a word variant
type VariantResponse struct {
	VariantText    string   `json:"variant_text"`
	Kind           string   `json:"kind"`
	FormType       string   `json:"form_type,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	FrequencyRank  int      `json:"frequency_rank,omitempty"`  // Variant frequency rank
	FrequencyCount int      `json:"frequency_count,omitempty"` // Variant frequency count
}

// QueriedVariantInfo represents the frequency info of the queried variant
// Used when a user queries a variant (e.g., "transplanted") but gets the main word
type QueriedVariantInfo struct {
	Text           string  `json:"text"`
	FrequencyRank  int     `json:"frequency_rank"`
	FrequencyCount int     `json:"frequency_count"`
	UsageRatio     float64 `json:"usage_ratio,omitempty"` // Percentage: variant_count / main_word_count * 100
}

// VariantReverseResponse represents a word found by variant lookup
type VariantReverseResponse struct {
	ID              uint                    `json:"id"`
	Headword        string                  `json:"headword"`
	WordAnnotations                         // Embedded: all annotation fields
	VariantInfo     []VariantResponse       `json:"variant_info"`
	Pronunciations  []PronunciationResponse `json:"pronunciations,omitempty"`
	Senses          []SenseResponse         `json:"senses,omitempty"`
}

// SearchResultResponse represents a search result item
type SearchResultResponse struct {
	ID              uint     `json:"id"`
	Headword        string   `json:"headword"`
	POS             []string `json:"pos"`
	WordAnnotations          // Embedded: all annotation fields
}

// SuggestResponse represents a suggestion item
type SuggestResponse struct {
	Headword        string `json:"headword"`
	WordAnnotations        // Embedded: all annotation fields
}

// BatchRequest represents batch query request
type BatchRequest struct {
	Words                 []string `json:"words" binding:"required"`
	IncludeVariants       *bool    `json:"include_variants,omitempty"`
	IncludePronunciations *bool    `json:"include_pronunciations,omitempty"`
	IncludeSenses         *bool    `json:"include_senses,omitempty"`
}
