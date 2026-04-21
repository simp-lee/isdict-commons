package model

import "time"

type EntrySummaryZH struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID     int64  `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_entry_summaries_zh_entry_id_source,priority:1;index:idx_entry_summaries_zh_entry_id"`
	Source      string `gorm:"type:text;not null;uniqueIndex:idx_entry_summaries_zh_entry_id_source,priority:2;index:idx_entry_summaries_zh_source_updated_at,priority:1"`
	SourceRunID int64  `gorm:"type:bigint;autoIncrement:false;not null"`
	SummaryText string `gorm:"column:summary_text;type:text;not null"`

	Entry     Entry     `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
	SourceRun ImportRun `gorm:"foreignKey:SourceRunID;references:ID;constraint:OnDelete:RESTRICT"`

	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now();index:idx_entry_summaries_zh_source_updated_at,priority:2,sort:desc"`
}

func (EntrySummaryZH) TableName() string {
	return "entry_summaries_zh"
}

type EntryEtymology struct {
	EntryID int64 `gorm:"primaryKey;autoIncrement:false;type:bigint"`

	Source             string  `gorm:"type:text;not null;index:idx_entry_etymologies_source_updated_at,priority:1"`
	SourceRunID        int64   `gorm:"type:bigint;autoIncrement:false;not null"`
	EtymologyTextRaw   string  `gorm:"column:etymology_text_raw;type:text;not null"`
	EtymologyTextClean *string `gorm:"column:etymology_text_clean;type:text"`

	Entry     Entry     `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
	SourceRun ImportRun `gorm:"foreignKey:SourceRunID;references:ID;constraint:OnDelete:RESTRICT"`

	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now();index:idx_entry_etymologies_source_updated_at,priority:2,sort:desc"`
}

func (EntryEtymology) TableName() string {
	return "entry_etymologies"
}
