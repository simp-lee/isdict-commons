package model

import "time"

type Entry struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	Headword           string `gorm:"type:text;not null;uniqueIndex:idx_entries_headword_pos_etymology_index,priority:1;index:idx_entries_headword"`
	NormalizedHeadword string `gorm:"type:text;not null;index:idx_entries_normalized_headword"`
	Pos                string `gorm:"type:text;not null;uniqueIndex:idx_entries_headword_pos_etymology_index,priority:2;index:idx_entries_pos"`
	EtymologyIndex     int    `gorm:"type:integer;not null;default:0;check:etymology_index >= 0;uniqueIndex:idx_entries_headword_pos_etymology_index,priority:3"`
	IsMultiword        bool   `gorm:"type:boolean;not null;default:false"`
	SourceRunID        int64  `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entries_source_run_id"`

	SourceRun ImportRun `gorm:"foreignKey:SourceRunID;references:ID;constraint:OnDelete:RESTRICT"`

	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Entry) TableName() string {
	return "entries"
}

type Sense struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID    int64 `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_senses_entry_id_sense_order,priority:1"`
	SenseOrder int16 `gorm:"type:smallint;not null;check:sense_order >= 1;uniqueIndex:idx_senses_entry_id_sense_order,priority:2"`

	Entry Entry `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Sense) TableName() string {
	return "senses"
}

type SenseGlossEN struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	SenseID    int64  `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_sense_glosses_en_sense_id_gloss_order,priority:1"`
	GlossOrder int16  `gorm:"type:smallint;not null;check:gloss_order >= 1;uniqueIndex:idx_sense_glosses_en_sense_id_gloss_order,priority:2"`
	TextEN     string `gorm:"column:text_en;type:text;not null"`

	Sense Sense `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:CASCADE"`
}

func (SenseGlossEN) TableName() string {
	return "sense_glosses_en"
}

type SenseGlossZH struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	SenseID      int64   `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_sense_glosses_zh_sense_id_source_gloss_order,priority:1;index:idx_sense_glosses_zh_sense_id_gloss_order,priority:1"`
	Source       string  `gorm:"type:text;not null;uniqueIndex:idx_sense_glosses_zh_sense_id_source_gloss_order,priority:2"`
	SourceRunID  int64   `gorm:"type:bigint;autoIncrement:false;not null;index:idx_sense_glosses_zh_source_run_id"`
	GlossOrder   int16   `gorm:"type:smallint;not null;check:gloss_order >= 1;uniqueIndex:idx_sense_glosses_zh_sense_id_source_gloss_order,priority:3;index:idx_sense_glosses_zh_sense_id_gloss_order,priority:2"`
	TextZHHans   string  `gorm:"column:text_zh_hans;type:text;not null"`
	DialectCode  *string `gorm:"type:text"`
	Romanization *string `gorm:"type:text"`
	IsPrimary    bool    `gorm:"type:boolean;not null;default:false"`

	Sense     Sense     `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:CASCADE"`
	SourceRun ImportRun `gorm:"foreignKey:SourceRunID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (SenseGlossZH) TableName() string {
	return "sense_glosses_zh"
}

type SenseLabel struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	SenseID    int64  `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_sense_labels_sense_id_label_type_label_code,priority:1;index:idx_sense_labels_sense_id_label_type_label_order,priority:1"`
	LabelType  string `gorm:"type:text;not null;check:label_type IN ('grammar','register','region','temporal','domain','attitude');uniqueIndex:idx_sense_labels_sense_id_label_type_label_code,priority:2;index:idx_sense_labels_sense_id_label_type_label_order,priority:2;index:idx_sense_labels_label_type_label_code,priority:1"`
	LabelCode  string `gorm:"type:text;not null;uniqueIndex:idx_sense_labels_sense_id_label_type_label_code,priority:3;index:idx_sense_labels_label_type_label_code,priority:2"`
	LabelOrder int16  `gorm:"type:smallint;not null;default:1;index:idx_sense_labels_sense_id_label_type_label_order,priority:3"`

	Sense Sense `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:CASCADE"`
}

func (SenseLabel) TableName() string {
	return "sense_labels"
}

type SenseExample struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	SenseID      int64  `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_sense_examples_sense_id_source_example_order,priority:1;index:idx_sense_examples_sense_id_example_order,priority:1"`
	Source       string `gorm:"type:text;not null;default:'wiktionary';uniqueIndex:idx_sense_examples_sense_id_source_example_order,priority:2"`
	ExampleOrder int16  `gorm:"type:smallint;not null;check:example_order >= 1;uniqueIndex:idx_sense_examples_sense_id_source_example_order,priority:3;index:idx_sense_examples_sense_id_example_order,priority:2"`
	SentenceEN   string `gorm:"column:sentence_en;type:text;not null"`

	Sense Sense `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:CASCADE"`
}

func (SenseExample) TableName() string {
	return "sense_examples"
}
