package model

import "time"

type EntryLearningSignal struct {
	EntryID int64 `gorm:"primaryKey;autoIncrement:false;type:bigint"`

	CEFRLevel      int16  `gorm:"type:smallint;not null;default:0;check:cefr_level >= 0 AND cefr_level <= 6;index:idx_entry_learning_signals_cefr_level"`
	CEFRSource     string `gorm:"type:text;not null;default:'';check:cefr_source IN ('','oxford','cefrj','both')"`
	CEFRRunID      *int64 `gorm:"type:bigint;autoIncrement:false"`
	OxfordLevel    int16  `gorm:"type:smallint;not null;default:0;check:oxford_level >= 0 AND oxford_level <= 2;index:idx_entry_learning_signals_oxford_level"`
	OxfordRunID    *int64 `gorm:"type:bigint;autoIncrement:false"`
	CETLevel       int16  `gorm:"type:smallint;not null;default:0;check:cet_level >= 0 AND cet_level <= 2;index:idx_entry_learning_signals_cet_level"`
	CETRunID       *int64 `gorm:"type:bigint;autoIncrement:false"`
	SchoolLevel    int16  `gorm:"type:smallint;not null;default:0;check:school_level >= 0 AND school_level <= 3;index:idx_entry_learning_signals_school_level"`
	FrequencyRank  int    `gorm:"type:integer;not null;default:0;check:frequency_rank >= 0;index:idx_entry_learning_signals_frequency_rank"`
	FrequencyCount int    `gorm:"type:integer;not null;default:0;check:frequency_count >= 0"`
	FrequencyRunID *int64 `gorm:"type:bigint;autoIncrement:false"`
	CollinsStars   int16  `gorm:"type:smallint;not null;default:0;check:collins_stars >= 0 AND collins_stars <= 5;index:idx_entry_learning_signals_collins_stars"`
	CollinsRunID   *int64 `gorm:"type:bigint;autoIncrement:false"`

	Entry        Entry      `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
	CEFRRun      *ImportRun `gorm:"foreignKey:CEFRRunID;references:ID;constraint:OnDelete:RESTRICT"`
	OxfordRun    *ImportRun `gorm:"foreignKey:OxfordRunID;references:ID;constraint:OnDelete:RESTRICT"`
	CETRun       *ImportRun `gorm:"foreignKey:CETRunID;references:ID;constraint:OnDelete:RESTRICT"`
	FrequencyRun *ImportRun `gorm:"foreignKey:FrequencyRunID;references:ID;constraint:OnDelete:RESTRICT"`
	CollinsRun   *ImportRun `gorm:"foreignKey:CollinsRunID;references:ID;constraint:OnDelete:RESTRICT"`

	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (EntryLearningSignal) TableName() string {
	return "entry_learning_signals"
}

type SenseLearningSignal struct {
	SenseID int64 `gorm:"primaryKey;autoIncrement:false;type:bigint"`

	CEFRLevel   int16  `gorm:"type:smallint;not null;default:0;check:cefr_level >= 0 AND cefr_level <= 6;index:idx_sense_learning_signals_cefr_level"`
	CEFRSource  string `gorm:"type:text;not null;default:'';check:cefr_source IN ('','oxford','cefrj','both')"`
	CEFRRunID   *int64 `gorm:"type:bigint;autoIncrement:false"`
	OxfordLevel int16  `gorm:"type:smallint;not null;default:0;check:oxford_level >= 0 AND oxford_level <= 2;index:idx_sense_learning_signals_oxford_level"`
	OxfordRunID *int64 `gorm:"type:bigint;autoIncrement:false"`

	Sense     Sense      `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:CASCADE"`
	CEFRRun   *ImportRun `gorm:"foreignKey:CEFRRunID;references:ID;constraint:OnDelete:RESTRICT"`
	OxfordRun *ImportRun `gorm:"foreignKey:OxfordRunID;references:ID;constraint:OnDelete:RESTRICT"`

	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (SenseLearningSignal) TableName() string {
	return "sense_learning_signals"
}
