package model

import (
	"time"

	"github.com/lib/pq"
)

// Word represents the database Word model - the main dictionary entry
//
// Field Usage Guidelines:
//
// 1. Headword (case-sensitive):
//   - Preserves the original case form of the word
//   - Used for exact matching in data import and API response display
//   - Distinguishes proper nouns from common words (e.g., "Polish" vs "polish")
//   - Unique index: idx_word_headword_unique
//
// 2. HeadwordNormalized (case-insensitive):
//   - Lowercase normalized form for search and autocomplete
//   - Removes spaces/hyphens/underscores and converts to lowercase
//   - Same normalized form may correspond to multiple entries (e.g., "Polish"/"polish" -> "polish")
//   - Regular index: idx_word_headword_normalized
//
// 3. Query Patterns:
//   - Exact query: WHERE headword = ? (case-sensitive)
//   - Search/autocomplete: WHERE headword_normalized = ? (case-insensitive)
//   - Batch query: WHERE headword_normalized IN (?) (performance optimization)
type Word struct {
	ID uint `gorm:"primaryKey"`

	// Headword is the exact form of the word (case-sensitive)
	Headword string `gorm:"type:varchar(255);not null;uniqueIndex:idx_word_headword_unique"`

	// HeadwordNormalized is the normalized form for case-insensitive lookups
	HeadwordNormalized string    `gorm:"type:varchar(255);not null;index:idx_words_headword_normalized"`
	CEFRLevel          int       `gorm:"type:smallint;not null;default:0;index:idx_words_cefr_level;check:cefr_level BETWEEN 0 AND 6"`
	CEFRSource         string    `gorm:"type:varchar(10);not null;default:'';check:cefr_source IN ('', 'oxford', 'cefrj', 'both')"` // CEFR data source
	CETLevel           int       `gorm:"type:smallint;not null;default:0;index:idx_words_cet_level;check:cet_level BETWEEN 0 AND 2"`
	OxfordLevel        int       `gorm:"type:smallint;not null;default:0;index:idx_words_oxford_level;check:oxford_level BETWEEN 0 AND 2"`
	SchoolLevel        int       `gorm:"type:smallint;not null;default:0;index:idx_words_school_level;check:school_level BETWEEN 0 AND 3"`
	FrequencyCount     int       `gorm:"not null;default:0;check:frequency_count >= 0"`
	FrequencyRank      int       `gorm:"not null;default:0;index:idx_words_frequency_rank;check:frequency_rank >= 0"`
	CollinsStars       int       `gorm:"type:smallint;not null;default:0;index:idx_words_collins_stars;check:collins_stars BETWEEN 0 AND 5"`
	TranslationZH      string    `gorm:"type:text;not null;default:''"`
	CreatedAt          time.Time `gorm:"<-:create"`
	UpdatedAt          time.Time `gorm:"<-:update"`

	Pronunciations []Pronunciation `gorm:"constraint:OnDelete:CASCADE"`
	Senses         []Sense         `gorm:"constraint:OnDelete:CASCADE"`
	WordVariants   []WordVariant   `gorm:"constraint:OnDelete:CASCADE"`
}

// Pronunciation represents pronunciation information for a word
//
// Accent codes: 0=Unknown, 1=British(RP), 2=American(GA), 3=Australian, 4=NewZealand,
// 5=Canadian, 6=Irish, 7=Scottish, 8=Indian, 9=SouthAfrican, 10=Other
//
// Note: A partial unique index on (word_id, accent) WHERE is_primary is created
// manually in migration.go to ensure only one primary pronunciation per word per accent.
// GORM's uniqueIndex tag cannot handle partial indexes with WHERE clauses.
type Pronunciation struct {
	ID        uint   `gorm:"primaryKey"`
	WordID    uint   `gorm:"not null;index:idx_pronunciations_word_id;uniqueIndex:idx_pronunciation_unique,priority:1"`
	Word      Word   `gorm:"constraint:OnDelete:CASCADE"`
	Accent    int    `gorm:"type:smallint;not null;check:accent BETWEEN 0 AND 10;uniqueIndex:idx_pronunciation_unique,priority:2"`
	IPA       string `gorm:"type:varchar(200);not null;uniqueIndex:idx_pronunciation_unique,priority:3"`
	IsPrimary bool   `gorm:"not null;default:false"`
}

// Sense represents a word sense/meaning with part of speech
type Sense struct {
	ID           uint   `gorm:"primaryKey"`
	WordID       uint   `gorm:"not null;index:idx_senses_word_id;uniqueIndex:idx_sense_unique,priority:1"`
	Word         Word   `gorm:"constraint:OnDelete:CASCADE"`
	POS          int    `gorm:"type:smallint;not null;check:pos BETWEEN 0 AND 22;uniqueIndex:idx_sense_unique,priority:2"`
	CEFRLevel    int    `gorm:"type:smallint;not null;default:0;check:cefr_level BETWEEN 0 AND 6"`
	CEFRSource   string `gorm:"type:varchar(10);not null;default:'';check:cefr_source IN ('', 'oxford', 'cefrj', 'both')"` // CEFR data source
	OxfordLevel  int    `gorm:"type:smallint;not null;default:0;check:oxford_level BETWEEN 0 AND 2"`                       // Sense-level Oxford annotation
	DefinitionEN string `gorm:"type:text;not null"`
	DefinitionZH string `gorm:"type:text;not null"`
	SenseOrder   int    `gorm:"type:smallint;not null;default:1;check:sense_order >= 1;uniqueIndex:idx_sense_unique,priority:3"`

	Examples []Example `gorm:"constraint:OnDelete:CASCADE"`
}

// Example represents an example sentence for a word sense
type Example struct {
	ID           uint   `gorm:"primaryKey"`
	SenseID      uint   `gorm:"not null;index:idx_examples_sense_id;uniqueIndex:idx_example_unique,priority:1"`
	Sense        Sense  `gorm:"constraint:OnDelete:CASCADE"`
	SentenceEN   string `gorm:"type:text;not null"`
	SentenceZH   string `gorm:"type:text"`
	ExampleOrder int    `gorm:"type:smallint;not null;default:1;check:example_order >= 1;uniqueIndex:idx_example_unique,priority:2"`
}

// VariantKind represents the type of word variant
type VariantKind uint8

const (
	VariantForm  VariantKind = 1 // Morphological variant (corresponds to FormType, from form-of tags)
	VariantAlias VariantKind = 2 // Spelling variant/orthographic variant (from alt-of tags, etc.)
)

// WordVariant represents alternative forms of a word (inflections, spellings, etc.)
//
// Note: The unique constraint on (word_id, variant_text, kind, form_type) is created
// manually in migration.go using COALESCE to handle NULL form_type values properly.
// GORM's uniqueIndex tag cannot handle expression-based indexes.
type WordVariant struct {
	ID                 uint           `gorm:"primaryKey"`
	WordID             uint           `gorm:"not null;index:idx_word_variants_word_id"`
	Word               Word           `gorm:"constraint:OnDelete:CASCADE"`
	VariantText        string         `gorm:"type:varchar(255);not null;index:idx_word_variants_variant_text"`
	HeadwordNormalized string         `gorm:"type:varchar(255);not null;default:'';index:idx_word_variants_headword_normalized"`
	Kind               VariantKind    `gorm:"type:smallint;not null;check:kind BETWEEN 1 AND 2"`
	FormType           *int           `gorm:"type:smallint;check:form_type BETWEEN 1 AND 9"`                                       // Only filled for morphological variants
	Tags               pq.StringArray `gorm:"type:text[]"`                                                                         // PostgreSQL native array for additional tags
	FrequencyCount     int            `gorm:"not null;default:0;check:frequency_count >= 0"`                                       // Frequency count for this variant
	FrequencyRank      int            `gorm:"not null;default:0;index:idx_word_variants_frequency_rank;check:frequency_rank >= 0"` // Frequency rank for this variant
	CreatedAt          time.Time      `gorm:"<-:create"`
	UpdatedAt          time.Time      `gorm:"<-:update"`
}
