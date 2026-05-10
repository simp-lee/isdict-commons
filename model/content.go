package model

import "time"

type EntryDefinition struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID             int64   `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entry_definitions_entry_id;index:idx_entry_definitions_entry_id_definition_order,priority:1;uniqueIndex:idx_entry_definitions_entry_id_pos_normalized_zh_hans_key,priority:1"`
	SenseID             *int64  `gorm:"type:bigint;autoIncrement:false;index:idx_entry_definitions_sense_id"`
	POS                 string  `gorm:"column:pos;type:text;not null;default:'';uniqueIndex:idx_entry_definitions_entry_id_pos_normalized_zh_hans_key,priority:2"`
	Source              string  `gorm:"type:text;not null;index:idx_entry_definitions_source_updated_at,priority:1"`
	SourceRunID         int64   `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entry_definitions_source_run_id"`
	DefinitionOrder     int16   `gorm:"type:smallint;not null;check:definition_order >= 1;index:idx_entry_definitions_entry_id_definition_order,priority:2"`
	TextZHHans          string  `gorm:"column:text_zh_hans;type:text;not null;check:text_zh_hans <> ''"`
	TextEN              *string `gorm:"column:text_en;type:text"`
	NormalizedZHHansKey string  `gorm:"column:normalized_zh_hans_key;type:text;not null;check:normalized_zh_hans_key <> '';uniqueIndex:idx_entry_definitions_entry_id_pos_normalized_zh_hans_key,priority:3"`
	NormalizedENKey     string  `gorm:"column:normalized_en_key;type:text;not null;default:''"`

	Entry     Entry     `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
	Sense     *Sense    `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:SET NULL"`
	SourceRun ImportRun `gorm:"foreignKey:SourceRunID;references:ID;constraint:OnDelete:RESTRICT"`

	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now();index:idx_entry_definitions_source_updated_at,priority:2,sort:desc"`
}

func (EntryDefinition) TableName() string {
	return "entry_definitions"
}

type EntryExample struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID                 int64   `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entry_examples_entry_id;index:idx_entry_examples_entry_id_example_order,priority:1;uniqueIndex:idx_entry_examples_entry_id_normalized_sentence_en_key,priority:1"`
	SenseID                 *int64  `gorm:"type:bigint;autoIncrement:false;index:idx_entry_examples_sense_id"`
	Source                  string  `gorm:"type:text;not null;index:idx_entry_examples_source_updated_at,priority:1"`
	SourceRunID             int64   `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entry_examples_source_run_id"`
	ExampleOrder            int16   `gorm:"type:smallint;not null;check:example_order >= 1;index:idx_entry_examples_entry_id_example_order,priority:2"`
	SentenceEN              string  `gorm:"column:sentence_en;type:text;not null;check:sentence_en <> ''"`
	SentenceZHHans          *string `gorm:"column:sentence_zh_hans;type:text"`
	NormalizedSentenceENKey string  `gorm:"column:normalized_sentence_en_key;type:text;not null;check:normalized_sentence_en_key <> '';uniqueIndex:idx_entry_examples_entry_id_normalized_sentence_en_key,priority:2"`
	NormalizedSentenceZHKey string  `gorm:"column:normalized_sentence_zh_key;type:text;not null;default:''"`

	Entry     Entry     `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
	Sense     *Sense    `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:SET NULL"`
	SourceRun ImportRun `gorm:"foreignKey:SourceRunID;references:ID;constraint:OnDelete:RESTRICT"`

	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now();index:idx_entry_examples_source_updated_at,priority:2,sort:desc"`
}

func (EntryExample) TableName() string {
	return "entry_examples"
}
