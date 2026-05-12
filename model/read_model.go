package model

type EntrySearchTerm struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID        int64  `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entry_search_terms_entry_id"`
	Headword       string `gorm:"type:text;not null"`
	TermText       string `gorm:"type:text;not null"`
	NormalizedTerm string `gorm:"type:text;not null;index:idx_entry_search_terms_normalized_term;index:idx_entry_search_terms_is_multiword_normalized_term,priority:2;index:idx_entry_search_terms_normalized_term_frequency_rank,priority:1"`
	TermKind       string `gorm:"type:text;not null;check:term_kind IN ('headword','form','alias')"`
	TermRank       int16  `gorm:"type:smallint;not null;check:term_rank IN (1,2)"`
	Pos            string `gorm:"type:text;not null;index:idx_entry_search_terms_pos"`
	IsMultiword    bool   `gorm:"type:boolean;not null;index:idx_entry_search_terms_is_multiword_normalized_term,priority:1"`

	FrequencyRank  int   `gorm:"type:integer;not null;default:0;check:frequency_rank >= 0;index:idx_entry_search_terms_normalized_term_frequency_rank,priority:2"`
	FrequencyCount int   `gorm:"type:integer;not null;default:0;check:frequency_count >= 0"`
	CEFRLevel      int16 `gorm:"type:smallint;not null;default:0;check:cefr_level >= 0 AND cefr_level <= 6"`
	OxfordLevel    int16 `gorm:"type:smallint;not null;default:0;check:oxford_level >= 0 AND oxford_level <= 2"`
	CETLevel       int16 `gorm:"type:smallint;not null;default:0;check:cet_level >= 0 AND cet_level <= 2"`
	CollinsStars   int16 `gorm:"type:smallint;not null;default:0;check:collins_stars >= 0 AND collins_stars <= 5"`
	SchoolLevel    int16 `gorm:"type:smallint;not null;default:0;check:school_level >= 0 AND school_level <= 3"`

	Entry Entry `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
}

func (EntrySearchTerm) TableName() string {
	return "entry_search_terms"
}

type FeaturedCandidate struct {
	EntryID int64 `gorm:"primaryKey;autoIncrement:false;type:bigint"`

	Headword           string `gorm:"type:text;not null"`
	NormalizedHeadword string `gorm:"type:text;not null;uniqueIndex:idx_featured_candidates_normalized_headword"`
	IsMultiword        bool   `gorm:"type:boolean;not null;index:idx_featured_candidates_is_multiword;index:idx_featured_candidates_is_multiword_quality_rank,priority:1"`
	Pos                string `gorm:"type:text;not null"`
	FrequencyRank      int    `gorm:"type:integer;not null;default:0;check:frequency_rank >= 0;index:idx_featured_candidates_frequency_rank"`
	CEFRLevel          int16  `gorm:"type:smallint;not null;default:0;check:cefr_level >= 0 AND cefr_level <= 6"`
	OxfordLevel        int16  `gorm:"type:smallint;not null;default:0;check:oxford_level >= 0 AND oxford_level <= 2"`
	CETLevel           int16  `gorm:"type:smallint;not null;default:0;check:cet_level >= 0 AND cet_level <= 2"`
	CollinsStars       int16  `gorm:"type:smallint;not null;default:0;check:collins_stars >= 0 AND collins_stars <= 5"`
	SchoolLevel        int16  `gorm:"type:smallint;not null;default:0;check:school_level >= 0 AND school_level <= 3"`
	QualityRank        int    `gorm:"type:integer;not null;check:quality_rank > 0;index:idx_featured_candidates_quality_rank;index:idx_featured_candidates_is_multiword_quality_rank,priority:2"`

	Entry Entry `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
}

func (FeaturedCandidate) TableName() string {
	return "featured_candidates"
}
