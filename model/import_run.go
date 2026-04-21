package model

import "time"

// ImportRun tracks a single import or enrichment execution for provenance and replay.
type ImportRun struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	SourceName string `gorm:"type:text;not null;index:idx_import_runs_source_name_started_at,priority:1"`
	SourcePath string `gorm:"type:text;not null"`

	SourceDumpID   string     `gorm:"type:text;not null;default:''"`
	SourceDumpDate *time.Time `gorm:"type:date"`

	RawFileSHA256   string `gorm:"type:text;not null;default:''"`
	ErrorCount      int64  `gorm:"type:bigint;not null;default:0"`
	PipelineVersion string `gorm:"type:text;not null"`
	Status          string `gorm:"type:text;not null;check:status IN ('running','completed','failed');index:idx_import_runs_status"`
	RowCount        int64  `gorm:"type:bigint;not null;default:0"`
	EntryCount      int64  `gorm:"type:bigint;not null;default:0"`
	Note            string `gorm:"type:text;not null;default:''"`

	StartedAt  time.Time  `gorm:"type:timestamptz;not null;default:now();index:idx_import_runs_source_name_started_at,priority:2,sort:desc"`
	FinishedAt *time.Time `gorm:"type:timestamptz"`
}

func (ImportRun) TableName() string {
	return "import_runs"
}
