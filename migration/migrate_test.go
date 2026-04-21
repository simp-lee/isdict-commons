package migration

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/simp-lee/isdict-commons/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

const (
	postgresIntegrationDSNEnv                  = "ISDICT_TEST_POSTGRES_DSN"
	postgresIntegrationDestructiveResetEnv     = "ISDICT_TEST_POSTGRES_ALLOW_DESTRUCTIVE_RESET"
	postgresIntegrationDestructiveResetConfirm = "drop-public-migration-tables"
	postgresIntegrationRemoteOverrideEnv       = "ISDICT_TEST_POSTGRES_ALLOW_REMOTE_DSN"
	postgresIntegrationRemoteOverrideConfirm   = "allow-remote-disposable-instance"
	postgresIntegrationHostEnv                 = "PGHOST"
	postgresIntegrationServiceEnv              = "PGSERVICE"
	postgresIntegrationServiceFileEnv          = "PGSERVICEFILE"
)

var postgresIntegrationDisposableDatabaseNames = []string{
	"isdict_test",
	"isdict_test_db",
	"isdict_integration_test",
	"isdict_integration_test_db",
	"isdict_migration_test",
	"isdict_migration_test_db",
}

var postgresIntegrationDisposableDatabaseNameSet = func() map[string]struct{} {
	accepted := make(map[string]struct{}, len(postgresIntegrationDisposableDatabaseNames))
	for _, name := range postgresIntegrationDisposableDatabaseNames {
		accepted[name] = struct{}{}
	}

	return accepted
}()

var postgresIntegrationLocalUnixSocketDirs = []string{
	"/var/run/postgresql",
	"/run/postgresql",
	"/tmp",
	"/private/tmp",
}

var postgresIntegrationLocalUnixSocketDirSet = func() map[string]struct{} {
	accepted := make(map[string]struct{}, len(postgresIntegrationLocalUnixSocketDirs))
	for _, dir := range postgresIntegrationLocalUnixSocketDirs {
		accepted[filepath.Clean(dir)] = struct{}{}
	}

	return accepted
}()

type postgresIntegrationTarget struct {
	DatabaseName  string `gorm:"column:database_name"`
	CurrentSchema string `gorm:"column:current_schema"`
	SearchPath    string `gorm:"column:search_path"`
}

type migrationFixture struct {
	EntryID      int64
	SenseID      int64
	FormID       int64
	ImportRunIDs map[string]int64
}

const (
	importRunKeyEntry                  = "entry"
	importRunKeySenseGlossZH           = "sense_gloss_zh"
	importRunKeyEntrySummary           = "entry_summary"
	importRunKeyEntryEtymology         = "entry_etymology"
	importRunKeyEntryLearningCEFR      = "entry_learning_cefr"
	importRunKeyEntryLearningOxford    = "entry_learning_oxford"
	importRunKeyEntryLearningCET       = "entry_learning_cet"
	importRunKeyEntryLearningFrequency = "entry_learning_frequency"
	importRunKeyEntryLearningCollins   = "entry_learning_collins"
	importRunKeySenseLearningCEFR      = "sense_learning_cefr"
	importRunKeySenseLearningOxford    = "sense_learning_oxford"
)

var postgresIntegrationAllImportRunKeys = []string{
	importRunKeyEntry,
	importRunKeySenseGlossZH,
	importRunKeyEntrySummary,
	importRunKeyEntryEtymology,
	importRunKeyEntryLearningCEFR,
	importRunKeyEntryLearningOxford,
	importRunKeyEntryLearningCET,
	importRunKeyEntryLearningFrequency,
	importRunKeyEntryLearningCollins,
	importRunKeySenseLearningCEFR,
	importRunKeySenseLearningOxford,
}

var postgresIntegrationEntryOwnedImportRunKeys = []string{
	importRunKeyEntry,
	importRunKeyEntrySummary,
	importRunKeyEntryEtymology,
	importRunKeyEntryLearningCEFR,
	importRunKeyEntryLearningOxford,
	importRunKeyEntryLearningCET,
	importRunKeyEntryLearningFrequency,
	importRunKeyEntryLearningCollins,
}

var postgresIntegrationSenseOwnedImportRunKeys = []string{
	importRunKeySenseGlossZH,
	importRunKeySenseLearningCEFR,
	importRunKeySenseLearningOxford,
}

type columnExpectation struct {
	TableName  string
	ColumnName string
	DataType   string
	UDTName    string
}

type columnMetadata struct {
	DataType string `gorm:"column:data_type"`
	UDTName  string `gorm:"column:udt_name"`
}

type postgresIntegrationIndexCatalog struct {
	TableName  string         `gorm:"column:table_name"`
	IndexName  string         `gorm:"column:index_name"`
	IsUnique   bool           `gorm:"column:is_unique"`
	Method     string         `gorm:"column:method"`
	Predicate  string         `gorm:"column:predicate"`
	Columns    pq.StringArray `gorm:"column:columns;type:text[]"`
	Definition string         `gorm:"column:definition"`
}

type indexExpectation struct {
	TableName string
	IndexName string
}

// These test-side contracts intentionally duplicate the accepted 15-table schema.
var postgresIntegrationExpectedTables = []string{
	"import_runs",
	"entries",
	"senses",
	"sense_glosses_en",
	"sense_glosses_zh",
	"sense_labels",
	"sense_examples",
	"pronunciation_ipas",
	"pronunciation_audios",
	"entry_forms",
	"lexical_relations",
	"entry_summaries_zh",
	"entry_learning_signals",
	"sense_learning_signals",
	"entry_etymologies",
}

var postgresIntegrationExpectedIndexes = []indexExpectation{
	{TableName: "import_runs", IndexName: "idx_import_runs_source_name_started_at"},
	{TableName: "import_runs", IndexName: "idx_import_runs_status"},
	{TableName: "entries", IndexName: "idx_entries_headword_pos_etymology_index"},
	{TableName: "entries", IndexName: "idx_entries_headword"},
	{TableName: "entries", IndexName: "idx_entries_normalized_headword"},
	{TableName: "entries", IndexName: "idx_entries_pos"},
	{TableName: "entries", IndexName: "idx_entries_source_run_id"},
	{TableName: "entries", IndexName: "idx_entries_normalized_headword_trgm"},
	{TableName: "senses", IndexName: "idx_senses_entry_id_sense_order"},
	{TableName: "sense_glosses_en", IndexName: "idx_sense_glosses_en_sense_id_gloss_order"},
	{TableName: "sense_glosses_zh", IndexName: "idx_sense_glosses_zh_sense_id_source_gloss_order"},
	{TableName: "sense_glosses_zh", IndexName: "idx_sense_glosses_zh_sense_id_gloss_order"},
	{TableName: "sense_glosses_zh", IndexName: "idx_sense_glosses_zh_source_run_id"},
	{TableName: "sense_glosses_zh", IndexName: "idx_sense_glosses_zh_sense_id_source_primary"},
	{TableName: "sense_labels", IndexName: "idx_sense_labels_sense_id_label_type_label_code"},
	{TableName: "sense_labels", IndexName: "idx_sense_labels_sense_id_label_type_label_order"},
	{TableName: "sense_labels", IndexName: "idx_sense_labels_label_type_label_code"},
	{TableName: "sense_examples", IndexName: "idx_sense_examples_sense_id_source_example_order"},
	{TableName: "sense_examples", IndexName: "idx_sense_examples_sense_id_example_order"},
	{TableName: "pronunciation_ipas", IndexName: "idx_pronunciation_ipas_entry_id_accent_code_ipa"},
	{TableName: "pronunciation_ipas", IndexName: "idx_pronunciation_ipas_entry_id_accent_code_display_order"},
	{TableName: "pronunciation_ipas", IndexName: "idx_pronunciation_ipas_entry_id_accent_code_primary"},
	{TableName: "pronunciation_audios", IndexName: "idx_pronunciation_audios_entry_id_accent_code_audio_filename"},
	{TableName: "pronunciation_audios", IndexName: "idx_pronunciation_audios_entry_id_accent_code_display_order"},
	{TableName: "pronunciation_audios", IndexName: "idx_pronunciation_audios_audio_filename"},
	{TableName: "pronunciation_audios", IndexName: "idx_pronunciation_audios_entry_id_accent_code_primary"},
	{TableName: "entry_forms", IndexName: "idx_entry_forms_entry_id_relation_kind"},
	{TableName: "entry_forms", IndexName: "idx_entry_forms_normalized_form"},
	{TableName: "entry_forms", IndexName: "idx_entry_forms_entry_id_relation_kind_form_text_form_type"},
	{TableName: "lexical_relations", IndexName: "idx_lexical_relations_entry_id_relation_type"},
	{TableName: "lexical_relations", IndexName: "idx_lexical_relations_sense_id_relation_type"},
	{TableName: "lexical_relations", IndexName: "idx_lexical_relations_entry_id_sense_id_rel_type_target_norm"},
	{TableName: "entry_summaries_zh", IndexName: "idx_entry_summaries_zh_entry_id_source"},
	{TableName: "entry_summaries_zh", IndexName: "idx_entry_summaries_zh_entry_id"},
	{TableName: "entry_summaries_zh", IndexName: "idx_entry_summaries_zh_source_updated_at"},
	{TableName: "entry_learning_signals", IndexName: "idx_entry_learning_signals_cefr_level"},
	{TableName: "entry_learning_signals", IndexName: "idx_entry_learning_signals_oxford_level"},
	{TableName: "entry_learning_signals", IndexName: "idx_entry_learning_signals_cet_level"},
	{TableName: "entry_learning_signals", IndexName: "idx_entry_learning_signals_school_level"},
	{TableName: "entry_learning_signals", IndexName: "idx_entry_learning_signals_frequency_rank"},
	{TableName: "entry_learning_signals", IndexName: "idx_entry_learning_signals_collins_stars"},
	{TableName: "sense_learning_signals", IndexName: "idx_sense_learning_signals_cefr_level"},
	{TableName: "sense_learning_signals", IndexName: "idx_sense_learning_signals_oxford_level"},
	{TableName: "entry_etymologies", IndexName: "idx_entry_etymologies_source_updated_at"},
}

var postgresIntegrationExpectedSQLManagedIndexDefinitions = []sqlIndexDefinitionTarget{
	{
		TableName: "pronunciation_ipas",
		IndexName: "idx_pronunciation_ipas_entry_id_accent_code_primary",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "accent_code"},
		Predicate: "is_primary = true",
	},
	{
		TableName: "pronunciation_audios",
		IndexName: "idx_pronunciation_audios_entry_id_accent_code_primary",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "accent_code"},
		Predicate: "is_primary = true",
	},
	{
		TableName: "sense_glosses_zh",
		IndexName: "idx_sense_glosses_zh_sense_id_source_primary",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"sense_id", "source"},
		Predicate: "is_primary = true",
	},
	{
		TableName: "entry_forms",
		IndexName: "idx_entry_forms_entry_id_relation_kind_form_text_form_type",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "relation_kind", "form_text", "COALESCE(form_type, '')"},
	},
	{
		TableName: "lexical_relations",
		IndexName: "idx_lexical_relations_entry_id_sense_id_rel_type_target_norm",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "COALESCE(sense_id, 0)", "relation_type", "target_text_normalized"},
	},
	{
		TableName: "entries",
		IndexName: "idx_entries_normalized_headword_trgm",
		Method:    "gin",
		Columns:   []string{"normalized_headword gin_trgm_ops"},
	},
}

var postgresIntegrationExpectedGORMIndexDefinitions = []sqlIndexDefinitionTarget{
	{
		TableName: "import_runs",
		IndexName: "idx_import_runs_source_name_started_at",
		Method:    "btree",
		Columns:   []string{"source_name", "started_at desc"},
	},
	{
		TableName: "import_runs",
		IndexName: "idx_import_runs_status",
		Method:    "btree",
		Columns:   []string{"status"},
	},
	{
		TableName: "entries",
		IndexName: "idx_entries_headword_pos_etymology_index",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"headword", "pos", "etymology_index"},
	},
	{
		TableName: "entries",
		IndexName: "idx_entries_headword",
		Method:    "btree",
		Columns:   []string{"headword"},
	},
	{
		TableName: "entries",
		IndexName: "idx_entries_normalized_headword",
		Method:    "btree",
		Columns:   []string{"normalized_headword"},
	},
	{
		TableName: "entries",
		IndexName: "idx_entries_pos",
		Method:    "btree",
		Columns:   []string{"pos"},
	},
	{
		TableName: "entries",
		IndexName: "idx_entries_source_run_id",
		Method:    "btree",
		Columns:   []string{"source_run_id"},
	},
	{
		TableName: "senses",
		IndexName: "idx_senses_entry_id_sense_order",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "sense_order"},
	},
	{
		TableName: "sense_glosses_en",
		IndexName: "idx_sense_glosses_en_sense_id_gloss_order",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"sense_id", "gloss_order"},
	},
	{
		TableName: "sense_glosses_zh",
		IndexName: "idx_sense_glosses_zh_sense_id_source_gloss_order",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"sense_id", "source", "gloss_order"},
	},
	{
		TableName: "sense_glosses_zh",
		IndexName: "idx_sense_glosses_zh_sense_id_gloss_order",
		Method:    "btree",
		Columns:   []string{"sense_id", "gloss_order"},
	},
	{
		TableName: "sense_glosses_zh",
		IndexName: "idx_sense_glosses_zh_source_run_id",
		Method:    "btree",
		Columns:   []string{"source_run_id"},
	},
	{
		TableName: "sense_labels",
		IndexName: "idx_sense_labels_sense_id_label_type_label_code",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"sense_id", "label_type", "label_code"},
	},
	{
		TableName: "sense_labels",
		IndexName: "idx_sense_labels_sense_id_label_type_label_order",
		Method:    "btree",
		Columns:   []string{"sense_id", "label_type", "label_order"},
	},
	{
		TableName: "sense_labels",
		IndexName: "idx_sense_labels_label_type_label_code",
		Method:    "btree",
		Columns:   []string{"label_type", "label_code"},
	},
	{
		TableName: "sense_examples",
		IndexName: "idx_sense_examples_sense_id_source_example_order",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"sense_id", "source", "example_order"},
	},
	{
		TableName: "sense_examples",
		IndexName: "idx_sense_examples_sense_id_example_order",
		Method:    "btree",
		Columns:   []string{"sense_id", "example_order"},
	},
	{
		TableName: "pronunciation_ipas",
		IndexName: "idx_pronunciation_ipas_entry_id_accent_code_ipa",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "accent_code", "ipa"},
	},
	{
		TableName: "pronunciation_ipas",
		IndexName: "idx_pronunciation_ipas_entry_id_accent_code_display_order",
		Method:    "btree",
		Columns:   []string{"entry_id", "accent_code", "display_order"},
	},
	{
		TableName: "pronunciation_audios",
		IndexName: "idx_pronunciation_audios_entry_id_accent_code_audio_filename",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "accent_code", "audio_filename"},
	},
	{
		TableName: "pronunciation_audios",
		IndexName: "idx_pronunciation_audios_entry_id_accent_code_display_order",
		Method:    "btree",
		Columns:   []string{"entry_id", "accent_code", "display_order"},
	},
	{
		TableName: "pronunciation_audios",
		IndexName: "idx_pronunciation_audios_audio_filename",
		Method:    "btree",
		Columns:   []string{"audio_filename"},
	},
	{
		TableName: "entry_forms",
		IndexName: "idx_entry_forms_entry_id_relation_kind",
		Method:    "btree",
		Columns:   []string{"entry_id", "relation_kind"},
	},
	{
		TableName: "entry_forms",
		IndexName: "idx_entry_forms_normalized_form",
		Method:    "btree",
		Columns:   []string{"normalized_form"},
	},
	{
		TableName: "lexical_relations",
		IndexName: "idx_lexical_relations_entry_id_relation_type",
		Method:    "btree",
		Columns:   []string{"entry_id", "relation_type"},
	},
	{
		TableName: "lexical_relations",
		IndexName: "idx_lexical_relations_sense_id_relation_type",
		Method:    "btree",
		Columns:   []string{"sense_id", "relation_type"},
	},
	{
		TableName: "entry_summaries_zh",
		IndexName: "idx_entry_summaries_zh_entry_id_source",
		Unique:    true,
		Method:    "btree",
		Columns:   []string{"entry_id", "source"},
	},
	{
		TableName: "entry_summaries_zh",
		IndexName: "idx_entry_summaries_zh_entry_id",
		Method:    "btree",
		Columns:   []string{"entry_id"},
	},
	{
		TableName: "entry_summaries_zh",
		IndexName: "idx_entry_summaries_zh_source_updated_at",
		Method:    "btree",
		Columns:   []string{"source", "updated_at desc"},
	},
	{
		TableName: "entry_learning_signals",
		IndexName: "idx_entry_learning_signals_cefr_level",
		Method:    "btree",
		Columns:   []string{"cefr_level"},
	},
	{
		TableName: "entry_learning_signals",
		IndexName: "idx_entry_learning_signals_oxford_level",
		Method:    "btree",
		Columns:   []string{"oxford_level"},
	},
	{
		TableName: "entry_learning_signals",
		IndexName: "idx_entry_learning_signals_cet_level",
		Method:    "btree",
		Columns:   []string{"cet_level"},
	},
	{
		TableName: "entry_learning_signals",
		IndexName: "idx_entry_learning_signals_school_level",
		Method:    "btree",
		Columns:   []string{"school_level"},
	},
	{
		TableName: "entry_learning_signals",
		IndexName: "idx_entry_learning_signals_frequency_rank",
		Method:    "btree",
		Columns:   []string{"frequency_rank"},
	},
	{
		TableName: "entry_learning_signals",
		IndexName: "idx_entry_learning_signals_collins_stars",
		Method:    "btree",
		Columns:   []string{"collins_stars"},
	},
	{
		TableName: "sense_learning_signals",
		IndexName: "idx_sense_learning_signals_cefr_level",
		Method:    "btree",
		Columns:   []string{"cefr_level"},
	},
	{
		TableName: "sense_learning_signals",
		IndexName: "idx_sense_learning_signals_oxford_level",
		Method:    "btree",
		Columns:   []string{"oxford_level"},
	},
	{
		TableName: "entry_etymologies",
		IndexName: "idx_entry_etymologies_source_updated_at",
		Method:    "btree",
		Columns:   []string{"source", "updated_at desc"},
	},
}

var embeddedSQLIndexDefinitionAllowList = []string{
	"sql/002_partial_indexes.sql",
	"sql/003_expression_indexes.sql",
	"sql/004_gin_indexes.sql",
}

var createIndexStatementPattern = regexp.MustCompile(`(?is)\bcreate\s+(unique\s+)?index\b`)

var sqlAssignmentTargetPattern = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*(?::=|=)\s*$`)

var sqlIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type postgresFormatCall struct {
	Template         string
	Arguments        []string
	AssignmentTarget string
}

func TestImportRunTableName_IsExplicitContract(t *testing.T) {
	t.Parallel()

	if got := (model.ImportRun{}).TableName(); got != "import_runs" {
		t.Fatalf("ImportRun.TableName() = %q; want %q", got, "import_runs")
	}
}

func TestMatchesIndexDefinition_AcceptsExpectedSQLManagedDefinitions(t *testing.T) {
	t.Parallel()

	for _, target := range postgresIntegrationExpectedSQLManagedIndexDefinitions {
		target := target
		t.Run(target.IndexName, func(t *testing.T) {
			t.Parallel()

			if !matchesIndexDefinition(buildCanonicalIndexDefinition(target, "public"), target) {
				t.Fatalf("matchesIndexDefinition(%q) = false; want true", target.IndexName)
			}
		})
	}
}

func TestMatchesIndexDefinition_RejectsDefinitionDrift(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		target          sqlIndexDefinitionTarget
		indexDefinition string
	}{
		{
			name:            "partial_index_missing_predicate",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_pronunciation_ipas_entry_id_accent_code_primary"),
			indexDefinition: "CREATE UNIQUE INDEX idx_pronunciation_ipas_entry_id_accent_code_primary ON public.pronunciation_ipas USING btree (entry_id, accent_code)",
		},
		{
			name:            "expression_index_wrong_coalesce_default",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_entry_forms_entry_id_relation_kind_form_text_form_type"),
			indexDefinition: "CREATE UNIQUE INDEX idx_entry_forms_entry_id_relation_kind_form_text_form_type ON public.entry_forms USING btree (entry_id, relation_kind, form_text, COALESCE(form_type, 'unknown'::text))",
		},
		{
			name:            "gin_index_wrong_method",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_entries_normalized_headword_trgm"),
			indexDefinition: "CREATE INDEX idx_entries_normalized_headword_trgm ON public.entries USING btree (normalized_headword)",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if matchesIndexDefinition(tt.indexDefinition, tt.target) {
				t.Fatalf("matchesIndexDefinition(%q) = true; want false", tt.target.IndexName)
			}
		})
	}
}

func TestMatchesIndexDefinition_AcceptsExpectedGORMManagedDefinitions(t *testing.T) {
	t.Parallel()

	for _, target := range postgresIntegrationExpectedGORMIndexDefinitions {
		target := target
		t.Run(target.IndexName, func(t *testing.T) {
			t.Parallel()

			if !matchesIndexDefinition(buildCanonicalIndexDefinition(target, "public"), target) {
				t.Fatalf("matchesIndexDefinition(%q) = false; want true", target.IndexName)
			}
		})
	}
}

func TestMatchesIndexDefinition_RejectsBusinessCriticalGORMDefinitionDrift(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		target          sqlIndexDefinitionTarget
		indexDefinition string
	}{
		{
			name:            "entries_unique_index_wrong_column_order",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_entries_headword_pos_etymology_index"),
			indexDefinition: "CREATE UNIQUE INDEX idx_entries_headword_pos_etymology_index ON public.entries USING btree (headword, etymology_index, pos)",
		},
		{
			name:            "senses_unique_index_missing_uniqueness",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_senses_entry_id_sense_order"),
			indexDefinition: "CREATE INDEX idx_senses_entry_id_sense_order ON public.senses USING btree (entry_id, sense_order)",
		},
		{
			name:            "summary_unique_index_wrong_column_set",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_entry_summaries_zh_entry_id_source"),
			indexDefinition: "CREATE UNIQUE INDEX idx_entry_summaries_zh_entry_id_source ON public.entry_summaries_zh USING btree (source, entry_id)",
		},
		{
			name:            "sense_examples_unique_index_wrong_terminal_column",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_sense_examples_sense_id_source_example_order"),
			indexDefinition: "CREATE UNIQUE INDEX idx_sense_examples_sense_id_source_example_order ON public.sense_examples USING btree (sense_id, source, sentence_en)",
		},
		{
			name:            "pronunciation_audio_unique_index_wrong_terminal_column",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_pronunciation_audios_entry_id_accent_code_audio_filename"),
			indexDefinition: "CREATE UNIQUE INDEX idx_pronunciation_audios_entry_id_accent_code_audio_filename ON public.pronunciation_audios USING btree (entry_id, accent_code, display_order)",
		},
		{
			name:            "import_runs_composite_index_missing_desc_sort",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_import_runs_source_name_started_at"),
			indexDefinition: "CREATE INDEX idx_import_runs_source_name_started_at ON public.import_runs USING btree (source_name, started_at)",
		},
		{
			name:            "entry_summaries_source_updated_at_missing_desc_sort",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_entry_summaries_zh_source_updated_at"),
			indexDefinition: "CREATE INDEX idx_entry_summaries_zh_source_updated_at ON public.entry_summaries_zh USING btree (source, updated_at)",
		},
		{
			name:            "entry_etymologies_source_updated_at_wrong_column_order",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_entry_etymologies_source_updated_at"),
			indexDefinition: "CREATE INDEX idx_entry_etymologies_source_updated_at ON public.entry_etymologies USING btree (updated_at DESC, source)",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if matchesIndexDefinition(tt.indexDefinition, tt.target) {
				t.Fatalf("matchesIndexDefinition(%q) = true; want false", tt.target.IndexName)
			}
		})
	}
}

func TestMatchesIndexDefinition_AcceptsWhitespaceNormalizedDefinitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		target          sqlIndexDefinitionTarget
		indexDefinition string
	}{
		{
			name:            "sql_managed_multiline_tabs_and_redundant_spaces",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_pronunciation_ipas_entry_id_accent_code_primary"),
			indexDefinition: "CREATE   UNIQUE   INDEX idx_pronunciation_ipas_entry_id_accent_code_primary\n\tON public.pronunciation_ipas USING btree (\n\t\tentry_id,\t accent_code\n\t)\nWHERE\t (  is_primary   =   true  )",
		},
		{
			name:            "gorm_managed_multiline_tabs_and_redundant_spaces",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_sense_examples_sense_id_source_example_order"),
			indexDefinition: "CREATE UNIQUE INDEX idx_sense_examples_sense_id_source_example_order\n\tON public.sense_examples USING btree (\n\t\tsense_id,\n\t\tsource,\t\texample_order\n\t)",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !matchesIndexDefinition(tt.indexDefinition, tt.target) {
				t.Fatalf("matchesIndexDefinition(%q) = false; want true", tt.target.IndexName)
			}
		})
	}
}

func TestMatchesIndexDefinition_AcceptsPostgresDeparserStyleDefinitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		target          sqlIndexDefinitionTarget
		indexDefinition string
	}{
		{
			name:            "partial_index_parenthesized_boolean_predicate",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_pronunciation_ipas_entry_id_accent_code_primary"),
			indexDefinition: "CREATE UNIQUE INDEX idx_pronunciation_ipas_entry_id_accent_code_primary ON public.pronunciation_ipas USING btree (entry_id, accent_code) WHERE (is_primary)",
		},
		{
			name:            "entry_forms_coalesce_with_text_casts",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_entry_forms_entry_id_relation_kind_form_text_form_type"),
			indexDefinition: "CREATE UNIQUE INDEX idx_entry_forms_entry_id_relation_kind_form_text_form_type ON public.entry_forms USING btree (entry_id, relation_kind, form_text, COALESCE((form_type)::text, ''::text))",
		},
		{
			name:            "lexical_relations_coalesce_with_bigint_casts_and_wrapping_parens",
			target:          mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_lexical_relations_entry_id_sense_id_rel_type_target_norm"),
			indexDefinition: "CREATE UNIQUE INDEX idx_lexical_relations_entry_id_sense_id_rel_type_target_norm ON public.lexical_relations USING btree (entry_id, (COALESCE((sense_id)::bigint, (0)::bigint)), relation_type, target_text_normalized)",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !matchesIndexDefinition(tt.indexDefinition, tt.target) {
				t.Fatalf("matchesIndexDefinition(%q) = false; want true", tt.target.IndexName)
			}
		})
	}
}

func TestMatchesIndexDefinition_AcceptsQuotedSchemaQualifiedDefinitions(t *testing.T) {
	t.Parallel()

	target := mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_entries_normalized_headword_trgm")
	indexDefinition := `CREATE INDEX idx_entries_normalized_headword_trgm ON "tenant-data".entries USING gin (normalized_headword "shared-extensions".gin_trgm_ops)`

	if !matchesIndexDefinition(indexDefinition, target) {
		t.Fatalf("matchesIndexDefinition(%q) = false; want true", target.IndexName)
	}
}

func TestPostgresIntegrationIndexMatchesExpectation_AcceptsQuotedSchemaQualifiedDefinitions(t *testing.T) {
	t.Parallel()

	target := mustIndexDefinitionTarget(t, postgresIntegrationExpectedSQLManagedIndexDefinitions, "idx_entries_normalized_headword_trgm")
	index := postgresIntegrationIndexCatalog{
		TableName:  "entries",
		IndexName:  "idx_entries_normalized_headword_trgm",
		Method:     "gin",
		Columns:    pq.StringArray{`normalized_headword "shared-extensions".gin_trgm_ops`},
		Definition: `CREATE INDEX idx_entries_normalized_headword_trgm ON "tenant-data".entries USING gin (normalized_headword "shared-extensions".gin_trgm_ops)`,
	}

	if !postgresIntegrationIndexMatchesExpectation(index, target) {
		t.Fatalf("postgresIntegrationIndexMatchesExpectation(%q) = false; want true", target.IndexName)
	}
}

func TestPostgresIntegrationIndexMatchesExpectation_AcceptsDescendingSortFromDefinition(t *testing.T) {
	t.Parallel()

	target := mustIndexDefinitionTarget(t, postgresIntegrationExpectedGORMIndexDefinitions, "idx_import_runs_source_name_started_at")
	index := postgresIntegrationIndexCatalog{
		TableName:  "import_runs",
		IndexName:  "idx_import_runs_source_name_started_at",
		Method:     "btree",
		Columns:    pq.StringArray{"source_name", "started_at"},
		Definition: "CREATE INDEX idx_import_runs_source_name_started_at ON public.import_runs USING btree (source_name, started_at DESC)",
	}

	if !postgresIntegrationIndexMatchesExpectation(index, target) {
		t.Fatalf("postgresIntegrationIndexMatchesExpectation(%q) = false; want true when definition preserves DESC sort", target.IndexName)
	}
}

func TestPostgresIntegrationExpectedTables_MirrorRuntimeMigrationTargets(t *testing.T) {
	t.Parallel()

	if !slices.Equal(postgresIntegrationExpectedTables, migrationTargetTableNames()) {
		t.Fatalf("postgresIntegrationExpectedTables = %#v; want %#v", postgresIntegrationExpectedTables, migrationTargetTableNames())
	}
}

func TestPostgresIntegrationExpectedIndexes_MirrorRuntimeVerificationTargets(t *testing.T) {
	t.Parallel()

	if !slices.EqualFunc(expectedIndexes, postgresIntegrationExpectedIndexes, equalIndexTargetAndExpectation) {
		t.Fatalf("postgresIntegrationExpectedIndexes = %#v; want %#v", postgresIntegrationExpectedIndexes, expectedIndexes)
	}
}

func TestPostgresIntegrationExpectedSQLManagedIndexDefinitions_MirrorRuntimeVerificationTargets(t *testing.T) {
	t.Parallel()

	if !slices.EqualFunc(sqlManagedIndexDefinitions, postgresIntegrationExpectedSQLManagedIndexDefinitions, equalIndexDefinitionTargets) {
		t.Fatalf("postgresIntegrationExpectedSQLManagedIndexDefinitions = %#v; want %#v", postgresIntegrationExpectedSQLManagedIndexDefinitions, sqlManagedIndexDefinitions)
	}
}

func TestPostgresIntegrationExpectedGORMIndexDefinitions_MirrorRuntimeVerificationTargets(t *testing.T) {
	t.Parallel()

	runtimeDefinitions := mustLoadGORMManagedIndexDefinitions(t)
	if !slices.EqualFunc(runtimeDefinitions, postgresIntegrationExpectedGORMIndexDefinitions, equalIndexDefinitionTargets) {
		t.Fatalf("postgresIntegrationExpectedGORMIndexDefinitions = %#v; want %#v", postgresIntegrationExpectedGORMIndexDefinitions, runtimeDefinitions)
	}
}

func TestExpectedIndexes_MatchModelTagsAndEmbeddedSQL(t *testing.T) {
	t.Parallel()

	sourceDerivedIndexes := sourceDerivedIndexExpectations(t)
	if got, want := sortedIndexTargetKeys(expectedIndexes), sortedIndexExpectationKeys(sourceDerivedIndexes); !slices.Equal(got, want) {
		t.Fatalf("expectedIndexes contracts = %v; want %v", got, want)
	}
	if got, want := sortedIndexExpectationKeys(postgresIntegrationExpectedIndexes), sortedIndexExpectationKeys(sourceDerivedIndexes); !slices.Equal(got, want) {
		t.Fatalf("postgresIntegrationExpectedIndexes contracts = %v; want %v", got, want)
	}
}

func TestSQLManagedIndexDefinitions_MatchEmbeddedSQL(t *testing.T) {
	t.Parallel()

	sourceDerivedDefinitions := deriveSQLManagedIndexDefinitionsFromEmbeddedSQL(t)
	if got, want := sortedIndexDefinitionKeys(sqlManagedIndexDefinitions), sortedIndexDefinitionKeys(sourceDerivedDefinitions); !slices.Equal(got, want) {
		t.Fatalf("sqlManagedIndexDefinitions contracts = %v; want %v", got, want)
	}
	if got, want := sortedIndexDefinitionKeys(postgresIntegrationExpectedSQLManagedIndexDefinitions), sortedIndexDefinitionKeys(sourceDerivedDefinitions); !slices.Equal(got, want) {
		t.Fatalf("postgresIntegrationExpectedSQLManagedIndexDefinitions contracts = %v; want %v", got, want)
	}
}

func TestParseGINIndexDefinitionsFromEmbeddedSQLFile_ExtractsAllCreateIndexTemplates(t *testing.T) {
	t.Parallel()

	sqlText := `
DO $$
DECLARE
	extension_schema text;
	qualified_table text;
BEGIN
	qualified_table := format('%I.%I', current_schema(), 'entries');
	EXECUTE format(
		'CREATE INDEX IF NOT EXISTS idx_entries_normalized_headword_trgm ON %s USING gin (normalized_headword %I.gin_trgm_ops)',
		qualified_table,
		extension_schema
	);

	qualified_table := format('%I.%I', current_schema(), 'entry_forms');
	EXECUTE format(
		'CREATE INDEX IF NOT EXISTS idx_entry_forms_normalized_form_trgm ON %s USING gin (normalized_form %I.gin_trgm_ops)',
		qualified_table,
		extension_schema
	);
END $$;
`

	got := parseGINIndexDefinitionsFromEmbeddedSQLFile(t, sqlText)
	want := []sqlIndexDefinitionTarget{
		{
			TableName: "entries",
			IndexName: "idx_entries_normalized_headword_trgm",
			Method:    "gin",
			Columns:   []string{"normalized_headword gin_trgm_ops"},
		},
		{
			TableName: "entry_forms",
			IndexName: "idx_entry_forms_normalized_form_trgm",
			Method:    "gin",
			Columns:   []string{"normalized_form gin_trgm_ops"},
		},
	}

	if !slices.EqualFunc(got, want, equalIndexDefinitionTargets) {
		t.Fatalf("parseGINIndexDefinitionsFromEmbeddedSQLFile() = %#v; want %#v", got, want)
	}
}

func TestEmbeddedSQLIndexDefinitionFiles_MatchDiscoveredEmbeddedSQLSet(t *testing.T) {
	t.Parallel()

	if got := discoverEmbeddedSQLIndexDefinitionFiles(t); !slices.Equal(got, embeddedSQLIndexDefinitionAllowList) {
		t.Fatalf("discoverEmbeddedSQLIndexDefinitionFiles() = %v; want %v", got, embeddedSQLIndexDefinitionAllowList)
	}
}

func TestManagedObjectContractDiff_ReportsMissingAndUnexpected(t *testing.T) {
	t.Parallel()

	missing, unexpected := managedObjectContractDiff(
		[]string{"entries", "import_runs", "senses"},
		[]string{"entries", "shadow_table"},
	)

	if !slices.Equal(missing, []string{"import_runs", "senses"}) {
		t.Fatalf("missing = %v; want %v", missing, []string{"import_runs", "senses"})
	}
	if !slices.Equal(unexpected, []string{"shadow_table"}) {
		t.Fatalf("unexpected = %v; want %v", unexpected, []string{"shadow_table"})
	}
}

func TestCurrentSchemaTableContractDiff_ReportsUnexpectedCurrentSchemaTables(t *testing.T) {
	t.Parallel()

	actualTables := append(append([]string(nil), postgresIntegrationExpectedTables...), "shadow_table")
	missing, unexpected := currentSchemaTableContractDiff(actualTables)
	if len(missing) > 0 {
		t.Fatalf("missing = %v; want nil", missing)
	}
	if !slices.Equal(unexpected, []string{"shadow_table"}) {
		t.Fatalf("unexpected = %v; want %v", unexpected, []string{"shadow_table"})
	}
}

func TestCurrentSchemaIndexContractDiff_ReportsUnexpectedSecondaryIndexes(t *testing.T) {
	t.Parallel()

	indexCatalog := make(map[string]postgresIntegrationIndexCatalog, len(postgresIntegrationExpectedIndexes)+2)
	for _, expectation := range postgresIntegrationExpectedIndexes {
		indexCatalog[postgresIntegrationIndexKey(expectation.TableName, expectation.IndexName)] = postgresIntegrationIndexCatalog{
			TableName: expectation.TableName,
			IndexName: expectation.IndexName,
			Method:    "btree",
		}
	}

	canonicalPrimaryKey := postgresIntegrationIndexKey("entries", "entries_pkey")
	unexpectedSecondaryIndex := postgresIntegrationIndexKey("entries", "idx_entries_sidecar_note")
	indexCatalog[canonicalPrimaryKey] = postgresIntegrationIndexCatalog{
		TableName: "entries",
		IndexName: "entries_pkey",
		IsUnique:  true,
		Method:    "btree",
	}
	indexCatalog[unexpectedSecondaryIndex] = postgresIntegrationIndexCatalog{
		TableName: "entries",
		IndexName: "idx_entries_sidecar_note",
		Method:    "btree",
	}

	filteredIndexCatalog, missing, unexpected := currentSchemaIndexContractDiff(indexCatalog)
	if len(missing) > 0 {
		t.Fatalf("missing = %v; want nil", formatManagedIndexKeys(missing))
	}
	if !slices.Equal(unexpected, []string{unexpectedSecondaryIndex}) {
		t.Fatalf("unexpected = %v; want %v", formatManagedIndexKeys(unexpected), formatManagedIndexKeys([]string{unexpectedSecondaryIndex}))
	}
	if _, ok := filteredIndexCatalog[canonicalPrimaryKey]; ok {
		t.Fatalf("currentSchemaIndexContractDiff() retained canonical primary-key index %q; want it filtered", canonicalPrimaryKey)
	}
	if _, ok := filteredIndexCatalog[unexpectedSecondaryIndex]; !ok {
		t.Fatalf("currentSchemaIndexContractDiff() removed unexpected secondary index %q; want it retained", unexpectedSecondaryIndex)
	}
}

func TestFilterAcceptedImplicitIndexes_OnlyDropsCanonicalPrimaryKeys(t *testing.T) {
	t.Parallel()

	indexCatalog := map[string]postgresIntegrationIndexCatalog{
		postgresIntegrationIndexKey("entries", "entries_pkey"): {
			TableName: "entries",
			IndexName: "entries_pkey",
			IsUnique:  true,
			Method:    "btree",
		},
		postgresIntegrationIndexKey("entries", "idx_entries_headword"): {
			TableName: "entries",
			IndexName: "idx_entries_headword",
			Method:    "btree",
		},
		postgresIntegrationIndexKey("entries", "shadow_pkey"): {
			TableName: "entries",
			IndexName: "shadow_pkey",
			IsUnique:  true,
			Method:    "btree",
		},
	}

	filtered := filterAcceptedImplicitIndexes(indexCatalog)
	if _, ok := filtered[postgresIntegrationIndexKey("entries", "entries_pkey")]; ok {
		t.Fatal("filterAcceptedImplicitIndexes() retained canonical primary-key index; want it filtered")
	}
	if _, ok := filtered[postgresIntegrationIndexKey("entries", "idx_entries_headword")]; !ok {
		t.Fatal("filterAcceptedImplicitIndexes() removed managed secondary index; want it retained")
	}
	if _, ok := filtered[postgresIntegrationIndexKey("entries", "shadow_pkey")]; !ok {
		t.Fatal("filterAcceptedImplicitIndexes() removed non-canonical *_pkey index name; want it retained")
	}
}

func TestPostgresIntegrationExpectedGORMIndexDefinitions_MatchModelTags(t *testing.T) {
	t.Parallel()

	sourceDerivedDefinitions := mustLoadGORMManagedIndexDefinitions(t)
	if got, want := sortedIndexDefinitionKeys(postgresIntegrationExpectedGORMIndexDefinitions), sortedIndexDefinitionKeys(sourceDerivedDefinitions); !slices.Equal(got, want) {
		t.Fatalf("postgresIntegrationExpectedGORMIndexDefinitions contracts = %v; want %v", got, want)
	}
}

func TestIdentityManagedTables_MatchModelDerivedIdentityScope(t *testing.T) {
	t.Parallel()

	derived := deriveIdentityManagedTablesFromModelPrimaryKeyShape(t)
	if !slices.Equal(identityManagedTables, derived) {
		t.Fatalf("identityManagedTables = %#v; want %#v", identityManagedTables, derived)
	}

	if got := extractIdentityManagedTablesFromEmbeddedSQL(t); !slices.Equal(got, derived) {
		t.Fatalf("005_identity_columns.sql identity targets = %#v; want %#v", got, derived)
	}
}

func TestLogAnalyzeWarning_AlwaysSignalsBestEffort(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	previousPrefix := log.Prefix()
	log.SetOutput(&output)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
	})

	logAnalyzeWarning(errors.New("planner stats unavailable"))

	logged := output.String()
	for _, fragment := range []string{"analyze tables warning", "best-effort", "migration continues", "planner stats unavailable"} {
		if !strings.Contains(logged, fragment) {
			t.Fatalf("logged warning = %q; want substring %q", logged, fragment)
		}
	}
}

func TestIdentityColumnStateIssue_RequiresGeneratedAlways(t *testing.T) {
	t.Parallel()

	issue := identityColumnStateIssue("entries", "public", identityColumnState{
		IsIdentity:         "YES",
		IdentityGeneration: "BY DEFAULT",
		SequenceSchema:     "public",
		SequenceName:       "entries_id_seq",
		SequenceLastValue:  sql.NullInt64{Int64: 42, Valid: true},
	})
	if issue == "" {
		t.Fatal("identityColumnStateIssue() = empty; want GENERATED ALWAYS contract failure")
	}

	for _, fragment := range []string{"identity generation", "BY DEFAULT", "ALWAYS"} {
		if !strings.Contains(issue, fragment) {
			t.Fatalf("identityColumnStateIssue() = %q; want substring %q", issue, fragment)
		}
	}
}

func TestIdentityColumnStateIssue_AcceptsGeneratedAlways(t *testing.T) {
	t.Parallel()

	issue := identityColumnStateIssue("entries", "public", identityColumnState{
		IsIdentity:         "YES",
		IdentityGeneration: "ALWAYS",
		SequenceSchema:     "public",
		SequenceName:       "entries_id_seq",
		SequenceLastValue:  sql.NullInt64{Int64: 42, Valid: true},
	})
	if issue != "" {
		t.Fatalf("identityColumnStateIssue() = %q; want empty issue", issue)
	}
}

func TestIdentityColumnStateIssue_AcceptsFreshGeneratedAlwaysWithoutLastValue(t *testing.T) {
	t.Parallel()

	issue := identityColumnStateIssue("entries", "public", identityColumnState{
		IsIdentity:         "YES",
		IdentityGeneration: "ALWAYS",
		SequenceSchema:     "public",
		SequenceName:       "entries_id_seq",
	})
	if issue != "" {
		t.Fatalf("identityColumnStateIssue() = %q; want empty issue for unused fresh sequence", issue)
	}
}

func TestIdentitySequencePositionIssue_AcceptsFreshSequenceWithoutLastValue(t *testing.T) {
	t.Parallel()

	issue := identitySequencePositionIssue("entries", identityColumnState{
		SequenceSchema: "public",
		SequenceName:   "entries_id_seq",
	}, 0)
	if issue != "" {
		t.Fatalf("identitySequencePositionIssue() = %q; want empty issue for fresh sequence before first insert", issue)
	}
}

func TestIdentitySequencePositionIssue_RejectsMissingLastValueWhenRowsExist(t *testing.T) {
	t.Parallel()

	issue := identitySequencePositionIssue("entries", identityColumnState{
		SequenceSchema: "public",
		SequenceName:   "entries_id_seq",
	}, 2)
	if issue == "" {
		t.Fatal("identitySequencePositionIssue() = empty; want drift failure when rows exist but sequence has no last_value")
	}

	for _, fragment := range []string{"has no last_value", "max(id)=2"} {
		if !strings.Contains(issue, fragment) {
			t.Fatalf("identitySequencePositionIssue() = %q; want substring %q", issue, fragment)
		}
	}
}

func TestPgTrgmSQL_ResolvesInstalledSchemaAtRuntime(t *testing.T) {
	t.Parallel()

	extensionsSQL := mustReadEmbeddedSQL(t, "sql/001_extensions.sql")
	if !strings.Contains(extensionsSQL, "CREATE EXTENSION IF NOT EXISTS pg_trgm") {
		t.Fatalf("001_extensions.sql = %q; want schema-agnostic CREATE EXTENSION", extensionsSQL)
	}
	for _, forbidden := range []string{"WITH SCHEMA public", "SET SCHEMA public"} {
		if strings.Contains(extensionsSQL, forbidden) {
			t.Fatalf("001_extensions.sql = %q; want no public-schema relocation fragment %q", extensionsSQL, forbidden)
		}
	}

	ginIndexSQL := mustReadEmbeddedSQL(t, "sql/004_gin_indexes.sql")
	for _, fragment := range []string{"FROM pg_extension ext", "ext.extname = 'pg_trgm'", "%I.gin_trgm_ops"} {
		if !strings.Contains(ginIndexSQL, fragment) {
			t.Fatalf("004_gin_indexes.sql = %q; want runtime schema resolution fragment %q", ginIndexSQL, fragment)
		}
	}
	if strings.Contains(ginIndexSQL, "public.gin_trgm_ops") {
		t.Fatalf("004_gin_indexes.sql = %q; want no public-specific operator class reference", ginIndexSQL)
	}
}

func mustReadEmbeddedSQL(t *testing.T, fileName string) string {
	t.Helper()

	sqlBytes, err := embeddedSQLFiles.ReadFile(fileName)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v; want nil", fileName, err)
	}

	return string(sqlBytes)
}

func TestIdentityColumnsSQL_UsesActiveSchema(t *testing.T) {
	t.Parallel()

	sqlText := mustReadEmbeddedSQL(t, "sql/005_identity_columns.sql")
	for _, forbidden := range []string{
		"table_schema = 'public'",
		"format('public.%I'",
		"FROM public.%I",
	} {
		if strings.Contains(sqlText, forbidden) {
			t.Fatalf("005_identity_columns.sql contains forbidden public-schema fragment %q", forbidden)
		}
	}

	if !strings.Contains(sqlText, "current_schema()") {
		t.Fatal("005_identity_columns.sql does not reference current_schema(); want active-schema resolution")
	}

	for _, fragment := range []string{"ADD GENERATED ALWAYS AS IDENTITY", "ALTER COLUMN id SET GENERATED ALWAYS"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("005_identity_columns.sql = %q; want repair fragment %q", sqlText, fragment)
		}
	}
}

func writePostgresIntegrationServiceFile(t *testing.T, serviceName, host, databaseName string) string {
	t.Helper()

	serviceFilePath := filepath.Join(t.TempDir(), "pg_service.conf")
	serviceFile := fmt.Sprintf(
		"[%s]\nhost=%s\nport=5432\ndbname=%s\nuser=isdict\npassword=secret\nsslmode=disable\n",
		serviceName,
		host,
		databaseName,
	)
	if err := os.WriteFile(serviceFilePath, []byte(serviceFile), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v; want nil", serviceFilePath, err)
	}

	return serviceFilePath
}

func TestValidatePostgresIntegrationDatabaseName_StrictContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		databaseName string
		wantErr      bool
	}{
		{name: "allows_exact_short_contract", databaseName: "isdict_test"},
		{name: "allows_exact_readme_contract", databaseName: "isdict_test_db"},
		{name: "allows_exact_integration_contract", databaseName: " isdict_integration_test "},
		{name: "allows_case_insensitive_match", databaseName: "ISDICT_MIGRATION_TEST_DB"},
		{name: "rejects_plain_database_name", databaseName: "isdict", wantErr: true},
		{name: "rejects_review_ab_testing_example", databaseName: "ab_testing", wantErr: true},
		{name: "rejects_review_customer_test_copy_example", databaseName: "customer-test-copy", wantErr: true},
		{name: "rejects_review_staging_tests_example", databaseName: "staging-tests", wantErr: true},
		{name: "rejects_isdict_prod_suffix", databaseName: "isdict_prod_test", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePostgresIntegrationDatabaseName(tt.databaseName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("validatePostgresIntegrationDatabaseName(%q) error = nil; want rejection", tt.databaseName)
				}
				if !strings.Contains(err.Error(), "disposable test database") {
					t.Fatalf("validatePostgresIntegrationDatabaseName(%q) error = %v; want disposable test database context", tt.databaseName, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("validatePostgresIntegrationDatabaseName(%q) error = %v; want nil", tt.databaseName, err)
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_StrictDatabaseContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name: "accepts_url_dsn_with_allowed_database",
			dsn:  "postgres:///isdict_test_db?sslmode=disable",
		},
		{
			name: "accepts_keyword_value_dsn_with_allowed_database",
			dsn:  "user=isdict password=secret dbname=isdict_integration_test sslmode=disable",
		},
		{
			name:    "rejects_url_dsn_with_review_ab_testing_database",
			dsn:     "postgres:///ab_testing?sslmode=disable",
			wantErr: true,
		},
		{
			name:    "rejects_url_dsn_with_review_customer_test_copy_database",
			dsn:     "postgres:///customer-test-copy?sslmode=disable",
			wantErr: true,
		},
		{
			name:    "rejects_keyword_value_dsn_with_review_staging_tests_database",
			dsn:     "user=isdict password=secret dbname=staging-tests sslmode=disable",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePostgresIntegrationDSNWithRemoteOverride(tt.dsn, "")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("validatePostgresIntegrationDSN(%q) error = nil; want rejection", tt.dsn)
				}
				if !strings.Contains(err.Error(), "disposable test database") {
					t.Fatalf("validatePostgresIntegrationDSN(%q) error = %v; want disposable test database context", tt.dsn, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("validatePostgresIntegrationDSN(%q) error = %v; want nil", tt.dsn, err)
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_DefaultAndUnixSocketLocalTargetContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "accepts_hostless_url_dsn",
			dsn:  "postgres:///isdict_migration_test_db?sslmode=disable",
		},
		{
			name: "accepts_hostless_keyword_value_dsn",
			dsn:  "user=isdict password=secret dbname=isdict_integration_test sslmode=disable",
		},
		{
			name: "accepts_hostless_url_with_explicit_socket_query",
			dsn:  "postgres:///isdict_test_db?host=/run/postgresql&sslmode=disable",
		},
		{
			name: "accepts_unix_socket_keyword_value_host",
			dsn:  "host=/var/run/postgresql port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := validatePostgresIntegrationDSNWithRemoteOverride(tt.dsn, ""); err != nil {
				t.Fatalf("validatePostgresIntegrationDSN(%q) error = %v; want nil", tt.dsn, err)
			}
		})
	}
}

func TestPostgresIntegrationDSNUsesExplicitHostSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dsn  string
		want bool
	}{
		{name: "hostless_url_dsn", dsn: "postgres:///isdict_test_db?sslmode=disable", want: false},
		{name: "hostless_keyword_value_dsn", dsn: "user=isdict password=secret dbname=isdict_test_db sslmode=disable", want: false},
		{name: "explicit_localhost_url_host", dsn: "postgres://user:pass@localhost:5432/isdict_test_db?sslmode=disable", want: true},
		{name: "explicit_keyword_value_loopback_host", dsn: "host=127.0.0.1 port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable", want: true},
		{name: "explicit_keyword_value_unix_socket_host", dsn: "host=/var/run/postgresql dbname=isdict_test_db", want: true},
		{name: "explicit_url_query_unix_socket_host", dsn: "postgres:///isdict_test_db?host=/run/postgresql", want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := postgresIntegrationDSNUsesExplicitHostSelection(tt.dsn)
			if err != nil {
				t.Fatalf("postgresIntegrationDSNUsesExplicitHostSelection(%q) error = %v; want nil", tt.dsn, err)
			}
			if got != tt.want {
				t.Fatalf("postgresIntegrationDSNUsesExplicitHostSelection(%q) = %t; want %t", tt.dsn, got, tt.want)
			}
		})
	}
}

func TestIsImplicitLocalPostgresIntegrationHost_DefaultConnectionClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		want bool
	}{
		{name: "accepts_empty_host", host: "", want: true},
		{name: "accepts_localhost", host: "localhost", want: true},
		{name: "accepts_ipv4_loopback", host: "127.0.0.1", want: true},
		{name: "accepts_ipv6_loopback", host: "[::1]", want: true},
		{name: "accepts_var_run_postgresql_socket_dir", host: "/var/run/postgresql", want: true},
		{name: "accepts_run_postgresql_socket_dir", host: "/run/postgresql", want: true},
		{name: "accepts_tmp_socket_dir", host: "/tmp", want: true},
		{name: "accepts_private_tmp_socket_dir", host: "/private/tmp", want: true},
		{name: "accepts_cleaned_socket_dir", host: "/var/run/postgresql/", want: true},
		{name: "rejects_proxy_like_cloudsql_socket_dir", host: "/cloudsql/project:region:instance", want: false},
		{name: "rejects_custom_absolute_socket_dir", host: "/srv/postgresql", want: false},
		{name: "rejects_remote_hostname", host: "prod-db.example.com", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isImplicitLocalPostgresIntegrationHost(tt.host); got != tt.want {
				t.Fatalf("isImplicitLocalPostgresIntegrationHost(%q) = %t; want %t", tt.host, got, tt.want)
			}
		})
	}
}

func TestIsPostgresIntegrationLocalUnixSocketHost_ExplicitHostClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		want bool
	}{
		{name: "rejects_empty_host", host: "", want: false},
		{name: "rejects_localhost", host: "localhost", want: false},
		{name: "rejects_ipv4_loopback", host: "127.0.0.1", want: false},
		{name: "rejects_ipv6_loopback", host: "[::1]", want: false},
		{name: "accepts_var_run_postgresql_socket_dir", host: "/var/run/postgresql", want: true},
		{name: "accepts_run_postgresql_socket_dir", host: "/run/postgresql", want: true},
		{name: "accepts_tmp_socket_dir", host: "/tmp", want: true},
		{name: "accepts_private_tmp_socket_dir", host: "/private/tmp", want: true},
		{name: "accepts_cleaned_socket_dir", host: "/var/run/postgresql/", want: true},
		{name: "rejects_proxy_like_cloudsql_socket_dir", host: "/cloudsql/project:region:instance", want: false},
		{name: "rejects_custom_absolute_socket_dir", host: "/srv/postgresql", want: false},
		{name: "rejects_remote_hostname", host: "prod-db.example.com", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isPostgresIntegrationLocalUnixSocketHost(tt.host); got != tt.want {
				t.Fatalf("isPostgresIntegrationLocalUnixSocketHost(%q) = %t; want %t", tt.host, got, tt.want)
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_RejectsExplicitHostTargetsWithoutRemoteOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dsn          string
		wantFragment string
	}{
		{
			name:         "rejects_explicit_localhost_url_host",
			dsn:          "postgres://user:pass@localhost:5432/isdict_test_db?sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_explicit_ipv4_loopback_keyword_value_host",
			dsn:          "host=127.0.0.1 port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_explicit_ipv6_loopback_url_host",
			dsn:          "postgres://user:pass@[::1]:5432/isdict_test_db?sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_forwarded_loopback_tcp_without_override",
			dsn:          "host=127.0.0.1 port=6543 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_remote_url_host",
			dsn:          "postgres://user:pass@prod-db.example.com:5432/isdict_test_db?sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_remote_keyword_value_host",
			dsn:          "host=prod-db.example.com port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_mixed_local_and_remote_fallback_hosts",
			dsn:          "host=localhost,prod-db.example.com port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
			wantFragment: "explicit host selection",
		},
		{
			name:         "rejects_proxy_like_unix_socket_host",
			dsn:          "host=/cloudsql/project:region:instance port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
			wantFragment: "explicit host selection",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePostgresIntegrationDSNWithRemoteOverride(tt.dsn, "")
			if err == nil {
				t.Fatalf("validatePostgresIntegrationDSN(%q) error = nil; want rejection", tt.dsn)
			}
			for _, fragment := range []string{tt.wantFragment, postgresIntegrationRemoteOverrideEnv} {
				if !strings.Contains(err.Error(), fragment) {
					t.Fatalf("validatePostgresIntegrationDSN(%q) error = %v; want fragment %q", tt.dsn, err, fragment)
				}
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_HostlessDSNRespectsPGHOSTSelection(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		wantErr       bool
		wantFragments []string
	}{
		{
			name:          "rejects_pg_host_localhost_without_override",
			host:          "localhost",
			wantErr:       true,
			wantFragments: []string{"host \"localhost\"", "explicit host selection", postgresIntegrationRemoteOverrideEnv},
		},
		{
			name:          "rejects_pg_host_ipv4_loopback_without_override",
			host:          "127.0.0.1",
			wantErr:       true,
			wantFragments: []string{"host \"127.0.0.1\"", "explicit host selection", postgresIntegrationRemoteOverrideEnv},
		},
		{
			name:    "allows_pg_host_local_unix_socket_dir",
			host:    "/var/run/postgresql",
			wantErr: false,
		},
		{
			name:          "rejects_pg_host_remote_hostname_without_override",
			host:          "prod-db.example.com",
			wantErr:       true,
			wantFragments: []string{"host \"prod-db.example.com\"", "explicit host selection", postgresIntegrationRemoteOverrideEnv},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(postgresIntegrationHostEnv, tt.host)
			t.Setenv(postgresIntegrationServiceEnv, "")
			t.Setenv(postgresIntegrationServiceFileEnv, "")

			err := validatePostgresIntegrationDSNWithRemoteOverride("postgres:///isdict_test_db?sslmode=disable", "")
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = %v; want nil for %s=%q", err, postgresIntegrationHostEnv, tt.host)
				}
				return
			}

			if err == nil {
				t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = nil; want rejection for %s=%q", postgresIntegrationHostEnv, tt.host)
			}

			for _, fragment := range tt.wantFragments {
				if !strings.Contains(err.Error(), fragment) {
					t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = %v; want fragment %q", err, fragment)
				}
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_RejectsEnvServiceIndirectionWithoutRemoteOverride(t *testing.T) {
	serviceFile := writePostgresIntegrationServiceFile(t, "ci_disposable", "prod-db.example.com", "isdict_test_db")
	tests := []struct {
		name   string
		setEnv func(t *testing.T)
	}{
		{
			name: "rejects_pgservice_from_env",
			setEnv: func(t *testing.T) {
				t.Setenv(postgresIntegrationServiceEnv, "ci_disposable")
				t.Setenv(postgresIntegrationServiceFileEnv, serviceFile)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnv(t)

			err := validatePostgresIntegrationDSNWithRemoteOverride("postgres:///isdict_test_db?sslmode=disable", "")
			if err == nil {
				t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = nil; want service rejection")
			}
			for _, fragment := range []string{"service/servicefile", postgresIntegrationServiceEnv, postgresIntegrationServiceFileEnv, postgresIntegrationRemoteOverrideEnv} {
				if !strings.Contains(err.Error(), fragment) {
					t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = %v; want fragment %q", err, fragment)
				}
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_AllowsLonePGSERVICEFILEEnvWithExplicitLocalDSN(t *testing.T) {
	serviceFile := writePostgresIntegrationServiceFile(t, "ci_disposable", "prod-db.example.com", "isdict_test_db")
	t.Setenv(postgresIntegrationServiceFileEnv, serviceFile)

	if err := validatePostgresIntegrationDSNWithRemoteOverride("host=/var/run/postgresql port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable", ""); err != nil {
		t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = %v; want nil when only %s is set", err, postgresIntegrationServiceFileEnv)
	}
}

func TestValidatePostgresIntegrationDSN_AllowsEnvServiceIndirectionWithExplicitOverride(t *testing.T) {
	serviceFile := writePostgresIntegrationServiceFile(t, "ci_disposable", "prod-db.example.com", "isdict_test_db")
	t.Setenv(postgresIntegrationServiceEnv, "ci_disposable")
	t.Setenv(postgresIntegrationServiceFileEnv, serviceFile)

	if err := validatePostgresIntegrationDSNWithRemoteOverride("postgres:///isdict_test_db?sslmode=disable", postgresIntegrationRemoteOverrideConfirm); err != nil {
		t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride() error = %v; want nil with explicit remote override", err)
	}
}

func TestValidatePostgresIntegrationDSN_RejectsServiceIndirectionWithoutRemoteOverride(t *testing.T) {
	t.Parallel()

	serviceFile := writePostgresIntegrationServiceFile(t, "ci_disposable", "prod-db.example.com", "isdict_test_db")
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "rejects_url_service_indirection",
			dsn:  fmt.Sprintf("postgres:///?servicefile=%s&service=ci_disposable", url.QueryEscape(serviceFile)),
		},
		{
			name: "rejects_keyword_value_service_indirection",
			dsn:  fmt.Sprintf("servicefile=%s service=ci_disposable", serviceFile),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePostgresIntegrationDSNWithRemoteOverride(tt.dsn, "")
			if err == nil {
				t.Fatalf("validatePostgresIntegrationDSN(%q) error = nil; want service rejection", tt.dsn)
			}
			for _, fragment := range []string{"service/servicefile", postgresIntegrationRemoteOverrideEnv} {
				if !strings.Contains(err.Error(), fragment) {
					t.Fatalf("validatePostgresIntegrationDSN(%q) error = %v; want fragment %q", tt.dsn, err, fragment)
				}
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_AllowsRemoteTargetsWithExplicitOverride(t *testing.T) {
	t.Parallel()

	serviceFile := writePostgresIntegrationServiceFile(t, "ci_disposable", "prod-db.example.com", "isdict_test_db")
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "allows_remote_url_host",
			dsn:  "postgres://user:pass@prod-db.example.com:5432/isdict_test_db?sslmode=disable",
		},
		{
			name: "allows_remote_keyword_value_host",
			dsn:  "host=prod-db.example.com port=5432 user=isdict password=secret dbname=isdict_test_db sslmode=disable",
		},
		{
			name: "allows_url_service_indirection",
			dsn:  fmt.Sprintf("postgres:///?servicefile=%s&service=ci_disposable", url.QueryEscape(serviceFile)),
		},
		{
			name: "allows_keyword_value_service_indirection",
			dsn:  fmt.Sprintf("servicefile=%s service=ci_disposable", serviceFile),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := validatePostgresIntegrationDSNWithRemoteOverride(tt.dsn, postgresIntegrationRemoteOverrideConfirm); err != nil {
				t.Fatalf("validatePostgresIntegrationDSNWithRemoteOverride(%q) error = %v; want nil", tt.dsn, err)
			}
		})
	}
}

func TestValidatePostgresIntegrationDSN_UsesExplicitRemoteOverrideEnv(t *testing.T) {
	t.Setenv(postgresIntegrationRemoteOverrideEnv, postgresIntegrationRemoteOverrideConfirm)

	err := validatePostgresIntegrationDSN("postgres://user:pass@prod-db.example.com:5432/isdict_test_db?sslmode=disable")
	if err != nil {
		t.Fatalf("validatePostgresIntegrationDSN() error = %v; want nil with explicit remote override env", err)
	}
}

func TestValidatePostgresIntegrationResetConfirmation_RequiresExplicitOptIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		confirmation string
		wantErr      bool
	}{
		{name: "rejects_blank_confirmation", confirmation: "", wantErr: true},
		{name: "rejects_incorrect_confirmation", confirmation: "yes", wantErr: true},
		{name: "accepts_exact_confirmation", confirmation: postgresIntegrationDestructiveResetConfirm},
		{name: "accepts_trimmed_confirmation", confirmation: " " + postgresIntegrationDestructiveResetConfirm + " "},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePostgresIntegrationResetConfirmation(tt.confirmation)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("validatePostgresIntegrationResetConfirmation(%q) error = nil; want rejection", tt.confirmation)
				}
				if !strings.Contains(err.Error(), postgresIntegrationDestructiveResetEnv) {
					t.Fatalf("validatePostgresIntegrationResetConfirmation(%q) error = %v; want env guidance", tt.confirmation, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("validatePostgresIntegrationResetConfirmation(%q) error = %v; want nil", tt.confirmation, err)
			}
		})
	}
}

func TestValidatePostgresIntegrationTarget_RejectsNonTestDatabase(t *testing.T) {
	t.Parallel()

	err := validatePostgresIntegrationTarget(postgresIntegrationTarget{
		DatabaseName:  "isdict",
		CurrentSchema: "public",
		SearchPath:    "public",
	})
	if err == nil {
		t.Fatal("validatePostgresIntegrationTarget() error = nil; want non-test database rejection")
	}

	if !strings.Contains(err.Error(), "disposable test database") {
		t.Fatalf("validatePostgresIntegrationTarget() error = %v; want test database context", err)
	}
}

func TestValidatePostgresIntegrationTarget_RejectsNonPublicSchemaSearchPath(t *testing.T) {
	t.Parallel()

	err := validatePostgresIntegrationTarget(postgresIntegrationTarget{
		DatabaseName:  "isdict_test",
		CurrentSchema: "tenant_a",
		SearchPath:    "tenant_a, public",
	})
	if err == nil {
		t.Fatal("validatePostgresIntegrationTarget() error = nil; want non-public schema rejection")
	}

	if !strings.Contains(err.Error(), "current_schema()") {
		t.Fatalf("validatePostgresIntegrationTarget() error = %v; want current_schema() context", err)
	}

	err = validatePostgresIntegrationTarget(postgresIntegrationTarget{
		DatabaseName:  "isdict_test",
		CurrentSchema: "public",
		SearchPath:    `"$user", public`,
	})
	if err == nil {
		t.Fatal("validatePostgresIntegrationTarget() error = nil; want search_path rejection")
	}

	if !strings.Contains(err.Error(), "search_path") {
		t.Fatalf("validatePostgresIntegrationTarget() error = %v; want search_path context", err)
	}
}

func TestValidatePostgresIntegrationTarget_AllowsTestDatabaseInPublicSchema(t *testing.T) {
	t.Parallel()

	err := validatePostgresIntegrationTarget(postgresIntegrationTarget{
		DatabaseName:  "isdict_test",
		CurrentSchema: "public",
		SearchPath:    "public",
	})
	if err != nil {
		t.Fatalf("validatePostgresIntegrationTarget() error = %v; want nil", err)
	}
}

func TestPostgresIntegrationResetStatements_TargetPublicSchemaOnly(t *testing.T) {
	t.Parallel()

	statements := postgresIntegrationResetStatements()
	want := postgresIntegrationExpectedResetStatements()
	if !slices.Equal(statements, want) {
		t.Fatalf("postgresIntegrationResetStatements() = %v; want %v", statements, want)
	}
}

func TestPostgresIntegrationResetStatements_UseRestrictWithoutCascade(t *testing.T) {
	t.Parallel()

	statements := postgresIntegrationResetStatements()
	if len(statements) != len(migrationTargets) {
		t.Fatalf("len(postgresIntegrationResetStatements()) = %d; want %d", len(statements), len(migrationTargets))
	}

	for _, stmt := range statements {
		if !strings.HasPrefix(stmt, "DROP TABLE IF EXISTS public.") {
			t.Fatalf("reset statement = %q; want public schema-qualified drop", stmt)
		}
		if !strings.HasSuffix(stmt, " RESTRICT") {
			t.Fatalf("reset statement = %q; want RESTRICT suffix", stmt)
		}
		if strings.Contains(stmt, " CASCADE") {
			t.Fatalf("reset statement = %q; want no CASCADE", stmt)
		}
	}
}

func TestResetPostgresIntegrationSchemaNoFail_PostgresIntegrationRejectsCrossSchemaDependents(t *testing.T) {
	db := newPostgresIntegrationDB(t)

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration() error = %v; want nil", err)
	}

	guardSchema, guardView := createPostgresIntegrationGuardView(
		t,
		db,
		"c07_reset_guard",
		"entry_etymologies_guard",
		"SELECT entry_id, source FROM public.entry_etymologies",
	)

	err := resetPostgresIntegrationSchemaNoFail(db)
	if err == nil {
		t.Fatal("resetPostgresIntegrationSchemaNoFail() error = nil; want cross-schema dependency rejection")
	}

	normalizedErr := strings.ToLower(err.Error())
	for _, fragment := range []string{"entry_etymologies", "sqlstate 2bp01"} {
		if !strings.Contains(normalizedErr, fragment) {
			t.Fatalf("resetPostgresIntegrationSchemaNoFail() error = %v; want fragment %q", err, fragment)
		}
	}

	assertSchemaRelationExists(t, db, guardSchema, guardView)
	assertSchemaRelationExists(t, db, "public", "entry_etymologies")
	if !db.Migrator().HasTable("entry_etymologies") {
		t.Fatal("entry_etymologies table missing after failed reset; want RESTRICT to preserve public table")
	}
}

func TestResetPostgresIntegrationSchemaNoFail_PostgresIntegrationRollsBackEarlierDropsWhenLaterRestrictFails(t *testing.T) {
	db := newPostgresIntegrationDB(t)

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration() error = %v; want nil", err)
	}

	fixture := seedMigrationFixture(t, db, "rollback-later-restrict")
	statements := postgresIntegrationResetStatements()
	targetStmt := "DROP TABLE IF EXISTS public.entries RESTRICT"
	targetIndex := slices.Index(statements, targetStmt)
	if targetIndex < 1 {
		t.Fatalf("reset order index for %q = %d; want > 0 so earlier drops execute before RESTRICT failure", targetStmt, targetIndex)
	}

	for _, earlierStmt := range []string{
		"DROP TABLE IF EXISTS public.entry_etymologies RESTRICT",
		"DROP TABLE IF EXISTS public.entry_forms RESTRICT",
		"DROP TABLE IF EXISTS public.sense_glosses_en RESTRICT",
	} {
		earlierIndex := slices.Index(statements, earlierStmt)
		if earlierIndex == -1 || earlierIndex >= targetIndex {
			t.Fatalf("reset order index for %q = %d, entries target = %d; want earlier table drop before later failure target", earlierStmt, earlierIndex, targetIndex)
		}
	}

	guardSchema, guardView := createPostgresIntegrationGuardView(
		t,
		db,
		"c07_reset_rollback",
		"entries_guard",
		"SELECT id, headword FROM public.entries",
	)

	err := resetPostgresIntegrationSchemaNoFail(db)
	if err == nil {
		t.Fatal("resetPostgresIntegrationSchemaNoFail() error = nil; want later cross-schema dependency rejection")
	}

	normalizedErr := strings.ToLower(err.Error())
	for _, fragment := range []string{"entries", "sqlstate 2bp01"} {
		if !strings.Contains(normalizedErr, fragment) {
			t.Fatalf("resetPostgresIntegrationSchemaNoFail() error = %v; want fragment %q", err, fragment)
		}
	}

	assertSchemaRelationExists(t, db, guardSchema, guardView)
	for _, tableName := range []string{"entry_etymologies", "entry_forms", "sense_glosses_en"} {
		assertSchemaRelationExists(t, db, "public", tableName)
	}
	assertEntryOwnedCascadeRows(t, db, fixture, 1)
	assertSenseOwnedCascadeRows(t, db, fixture, 1)
}

func TestRunMigration_PostgresIntegration(t *testing.T) {
	db := newPostgresIntegrationDB(t)

	if !t.Run("initial migration creates and verifies schema", func(t *testing.T) {
		runInitialMigrationContractSubtest(t, db)
	}) {
		return
	}

	if !t.Run("schema contract surfaces extra current-schema objects", func(t *testing.T) {
		runSchemaContractExtraObjectSubtest(t, db)
	}) {
		return
	}

	if !t.Run("sense delete cascades sense-level rows and releases only sense-owned import runs", func(t *testing.T) {
		runSenseDeleteCascadeSubtest(t, db)
	}) {
		return
	}

	if !t.Run("entry delete cascades entry-level and sense-level rows and clears all import run restrictions", func(t *testing.T) {
		runEntryDeleteCascadeSubtest(t, db)
	}) {
		return
	}

	sentinel := seedMigrationFixture(t, db, "sentinel")
	if !t.Run("repeated RunMigration calls remain idempotent", func(t *testing.T) {
		runRepeatedRunMigrationIdempotenceSubtest(t, db, sentinel)
	}) {
		return
	}

	t.Run("DropTables rebuilds a clean schema", func(t *testing.T) {
		runDropTablesRebuildSubtest(t, db)
	})
}

func runInitialMigrationContractSubtest(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration() error = %v; want nil", err)
	}

	assertMigrationTablesExist(t, db)
	assertExpectedIndexesExist(t, db)
	assertColumnTypes(t, db, []columnExpectation{
		{TableName: "import_runs", ColumnName: "source_dump_date", DataType: "date", UDTName: "date"},
		{TableName: "entries", ColumnName: "pos", DataType: "text", UDTName: "text"},
		{TableName: "pronunciation_ipas", ColumnName: "accent_code", DataType: "text", UDTName: "text"},
		{TableName: "pronunciation_audios", ColumnName: "accent_code", DataType: "text", UDTName: "text"},
		{TableName: "entry_forms", ColumnName: "source_relations", DataType: "ARRAY", UDTName: "_text"},
	})
}

func runSchemaContractExtraObjectSubtest(t *testing.T, db *gorm.DB) {
	t.Helper()

	sidecarSuffix := time.Now().UnixNano()
	sidecarTable := fmt.Sprintf("c08_sidecar_%d", sidecarSuffix)
	sidecarIndex := fmt.Sprintf("idx_c08_sidecar_note_%d", sidecarSuffix)
	quotedSidecarTable := quotePostgresIdentifier(sidecarTable)
	quotedSidecarIndex := quotePostgresIdentifier(sidecarIndex)
	t.Cleanup(func() {
		_ = db.Exec("DROP TABLE IF EXISTS " + quotedSidecarTable + " RESTRICT").Error
	})

	if err := db.Exec("CREATE TABLE " + quotedSidecarTable + " (id bigint PRIMARY KEY, note text NOT NULL)").Error; err != nil {
		t.Fatalf("CREATE TABLE %s error = %v; want nil", sidecarTable, err)
	}
	if err := db.Exec("CREATE INDEX " + quotedSidecarIndex + " ON " + quotedSidecarTable + " (note)").Error; err != nil {
		t.Fatalf("CREATE INDEX %s ON %s(note) error = %v; want nil", sidecarIndex, sidecarTable, err)
	}

	actualTables, err := loadCurrentSchemaTableNames(db)
	if err != nil {
		t.Fatalf("loadCurrentSchemaTableNames() error = %v; want nil", err)
	}
	missingTables, unexpectedTables := currentSchemaTableContractDiff(actualTables)
	if len(missingTables) > 0 || !slices.Equal(unexpectedTables, []string{sidecarTable}) {
		t.Fatalf("currentSchemaTableContractDiff() = (missing=%v, unexpected=%v); want (missing=[], unexpected=%v)", missingTables, unexpectedTables, []string{sidecarTable})
	}

	indexCatalog, err := loadCurrentSchemaIndexCatalog(db)
	if err != nil {
		t.Fatalf("loadCurrentSchemaIndexCatalog() error = %v; want nil", err)
	}
	filteredIndexCatalog, missingIndexes, unexpectedIndexes := currentSchemaIndexContractDiff(indexCatalog)
	wantUnexpectedIndexes := []string{postgresIntegrationIndexKey(sidecarTable, sidecarIndex)}
	if len(missingIndexes) > 0 || !slices.Equal(unexpectedIndexes, wantUnexpectedIndexes) {
		t.Fatalf(
			"currentSchemaIndexContractDiff() = (missing=%v, unexpected=%v); want (missing=[], unexpected=%v)",
			formatManagedIndexKeys(missingIndexes),
			formatManagedIndexKeys(unexpectedIndexes),
			formatManagedIndexKeys(wantUnexpectedIndexes),
		)
	}
	if _, ok := filteredIndexCatalog[wantUnexpectedIndexes[0]]; !ok {
		t.Fatalf("currentSchemaIndexContractDiff() removed unexpected secondary index %q; want it retained", wantUnexpectedIndexes[0])
	}
}

func runSenseDeleteCascadeSubtest(t *testing.T, db *gorm.DB) {
	t.Helper()

	fixture := seedMigrationFixture(t, db, "sense-delete")

	assertEntryFormSourceRelations(t, db, fixture.FormID, []string{"wiktionary", "plural"})
	assertModelRowCount(t, db, &model.Entry{}, 1, "id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.Sense{}, 1, "id = ?", fixture.SenseID)
	assertEntryOwnedCascadeRows(t, db, fixture, 1)
	assertSenseOwnedCascadeRows(t, db, fixture, 1)
	assertImportRunsExist(t, db, fixture, postgresIntegrationAllImportRunKeys)
	assertImportRunsDeleteRestricted(t, db, fixture, postgresIntegrationAllImportRunKeys)

	if err := db.Delete(&model.Sense{}, fixture.SenseID).Error; err != nil {
		t.Fatalf("Delete(sense) error = %v; want nil", err)
	}

	assertModelRowCount(t, db, &model.Entry{}, 1, "id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.Sense{}, 0, "id = ?", fixture.SenseID)
	assertEntryOwnedCascadeRows(t, db, fixture, 1)
	assertSenseOwnedCascadeRows(t, db, fixture, 0)
	assertImportRunsDeleteAllowed(t, db, fixture, postgresIntegrationSenseOwnedImportRunKeys)
	assertImportRunsDeleteRestricted(t, db, fixture, postgresIntegrationEntryOwnedImportRunKeys)
}

func runEntryDeleteCascadeSubtest(t *testing.T, db *gorm.DB) {
	t.Helper()

	fixture := seedMigrationFixture(t, db, "entry-delete")

	assertEntryFormSourceRelations(t, db, fixture.FormID, []string{"wiktionary", "plural"})
	assertModelRowCount(t, db, &model.Entry{}, 1, "id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.Sense{}, 1, "id = ?", fixture.SenseID)
	assertEntryOwnedCascadeRows(t, db, fixture, 1)
	assertSenseOwnedCascadeRows(t, db, fixture, 1)
	assertImportRunsDeleteRestricted(t, db, fixture, postgresIntegrationAllImportRunKeys)

	if err := db.Delete(&model.Entry{}, fixture.EntryID).Error; err != nil {
		t.Fatalf("Delete(entry) error = %v; want nil", err)
	}

	assertModelRowCount(t, db, &model.Entry{}, 0, "id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.Sense{}, 0, "id = ?", fixture.SenseID)
	assertEntryOwnedCascadeRows(t, db, fixture, 0)
	assertSenseOwnedCascadeRows(t, db, fixture, 0)
	assertImportRunsDeleteAllowed(t, db, fixture, postgresIntegrationAllImportRunKeys)
}

func runRepeatedRunMigrationIdempotenceSubtest(t *testing.T, db *gorm.DB, sentinel migrationFixture) {
	t.Helper()

	preRerunImportRunMaxID := loadTableMaxID(t, db, "import_runs")
	preRerunEntryMaxID := loadTableMaxID(t, db, "entries")

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("second RunMigration() error = %v; want nil", err)
	}

	assertMigrationTablesExist(t, db)
	assertExpectedIndexesExist(t, db)
	assertImportRunsExist(t, db, sentinel, postgresIntegrationAllImportRunKeys)
	assertModelRowCount(t, db, &model.Entry{}, 1, "id = ?", sentinel.EntryID)
	assertModelRowCount(t, db, &model.Sense{}, 1, "id = ?", sentinel.SenseID)
	assertEntryOwnedCascadeRows(t, db, sentinel, 1)
	assertSenseOwnedCascadeRows(t, db, sentinel, 1)

	rerunImportRunID := createImportRun(t, db, "post-rerun", "idempotence", time.Now().UTC())
	if rerunImportRunID <= preRerunImportRunMaxID {
		t.Fatalf("post-rerun import_runs.id = %d; want value greater than pre-rerun max %d", rerunImportRunID, preRerunImportRunMaxID)
	}

	rerunEntry := model.Entry{
		Headword:           "entry-post-rerun",
		NormalizedHeadword: "entrypostrerun",
		Pos:                model.POSNoun,
		EtymologyIndex:     0,
		SourceRunID:        rerunImportRunID,
	}
	if err := db.Create(&rerunEntry).Error; err != nil {
		t.Fatalf("Create(post-rerun entry) error = %v; want nil", err)
	}
	if rerunEntry.ID <= preRerunEntryMaxID {
		t.Fatalf("post-rerun entries.id = %d; want value greater than pre-rerun max %d", rerunEntry.ID, preRerunEntryMaxID)
	}

	assertModelRowCount(t, db, &model.ImportRun{}, 1, "id = ?", rerunImportRunID)
	assertModelRowCount(t, db, &model.Entry{}, 1, "id = ?", rerunEntry.ID)
}

func runDropTablesRebuildSubtest(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := RunMigration(db, MigrateOptions{DropTables: true}); err != nil {
		t.Fatalf("RunMigration(DropTables=true) error = %v; want nil", err)
	}

	assertMigrationTablesExist(t, db)
	assertExpectedIndexesExist(t, db)
	assertAllMigrationTablesEmpty(t, db)
}

func TestRunMigration_PostgresIntegration_UsesCurrentSchema(t *testing.T) {
	db := newPostgresIntegrationDB(t)

	tx := db.Begin()
	if err := tx.Error; err != nil {
		t.Fatalf("Begin() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	schemaName := fmt.Sprintf("migration_active_schema_%d", time.Now().UnixNano())
	quotedSchemaName := `"` + strings.ReplaceAll(schemaName, `"`, `""`) + `"`
	if err := tx.Exec("CREATE SCHEMA " + quotedSchemaName).Error; err != nil {
		t.Fatalf("Create schema %q error = %v; want nil", schemaName, err)
	}
	if err := tx.Exec("SELECT set_config('search_path', ?, true)", schemaName).Error; err != nil {
		t.Fatalf("set_config(search_path=%q) error = %v; want nil", schemaName, err)
	}

	target, err := inspectPostgresIntegrationTarget(tx)
	if err != nil {
		t.Fatalf("inspectPostgresIntegrationTarget() error = %v; want nil", err)
	}
	if target.CurrentSchema != schemaName {
		t.Fatalf("current_schema() = %q; want %q", target.CurrentSchema, schemaName)
	}
	if got := normalizePostgresSearchPath(target.SearchPath); got != schemaName {
		t.Fatalf("search_path = %q; want %q", got, schemaName)
	}

	if err := RunMigration(tx, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration(non-public schema) error = %v; want nil", err)
	}

	assertMigrationTablesExist(t, tx)
	assertExpectedIndexesExist(t, tx)

	fixture := seedMigrationFixture(t, tx, "nonpublic-schema")
	preRerunImportRunMaxID, err := loadCurrentSchemaMaxID(tx, "import_runs")
	if err != nil {
		t.Fatalf("loadCurrentSchemaMaxID(import_runs) error = %v; want nil", err)
	}
	preRerunEntryMaxID, err := loadCurrentSchemaMaxID(tx, "entries")
	if err != nil {
		t.Fatalf("loadCurrentSchemaMaxID(entries) error = %v; want nil", err)
	}

	if err := RunMigration(tx, MigrateOptions{}); err != nil {
		t.Fatalf("second RunMigration(non-public schema) error = %v; want nil", err)
	}

	assertImportRunsExist(t, tx, fixture, postgresIntegrationAllImportRunKeys)
	assertModelRowCount(t, tx, &model.Entry{}, 1, "id = ?", fixture.EntryID)

	rerunImportRunID := createImportRun(t, tx, "nonpublic-rerun", "idempotence", time.Now().UTC())
	if rerunImportRunID <= preRerunImportRunMaxID {
		t.Fatalf("post-rerun import_runs.id = %d; want value greater than pre-rerun max %d", rerunImportRunID, preRerunImportRunMaxID)
	}

	rerunEntry := model.Entry{
		Headword:           "entry-nonpublic-rerun",
		NormalizedHeadword: "entrynonpublicrerun",
		Pos:                model.POSNoun,
		EtymologyIndex:     0,
		SourceRunID:        rerunImportRunID,
	}
	if err := tx.Create(&rerunEntry).Error; err != nil {
		t.Fatalf("Create(non-public post-rerun entry) error = %v; want nil", err)
	}
	if rerunEntry.ID <= preRerunEntryMaxID {
		t.Fatalf("post-rerun entries.id = %d; want value greater than pre-rerun max %d", rerunEntry.ID, preRerunEntryMaxID)
	}
}

func TestRunMigration_PostgresIntegration_PgTrgmRemainsStableAcrossSchemas(t *testing.T) {
	db := newPostgresIntegrationDB(t)

	tx := db.Begin()
	if err := tx.Error; err != nil {
		t.Fatalf("Begin() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	firstSchema := fmt.Sprintf("migration_pgtrgm_first_%d", time.Now().UnixNano())
	secondSchema := firstSchema + "_second"
	for _, schemaName := range []string{firstSchema, secondSchema} {
		quotedSchemaName := `"` + strings.ReplaceAll(schemaName, `"`, `""`) + `"`
		if err := tx.Exec("CREATE SCHEMA " + quotedSchemaName).Error; err != nil {
			t.Fatalf("Create schema %q error = %v; want nil", schemaName, err)
		}
	}

	if err := tx.Exec("SELECT set_config('search_path', ?, true)", firstSchema).Error; err != nil {
		t.Fatalf("set_config(search_path=%q) error = %v; want nil", firstSchema, err)
	}
	if err := RunMigration(tx, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration(first schema=%q) error = %v; want nil", firstSchema, err)
	}
	assertMigrationTablesExist(t, tx)
	assertExpectedIndexesExist(t, tx)

	extensionSchema, err := loadExtensionSchema(tx, "pg_trgm")
	if err != nil {
		t.Fatalf("loadExtensionSchema(%q) error = %v; want nil", "pg_trgm", err)
	}
	assertCurrentSchemaIndexUsesOperatorClassSchema(t, tx, "entries", "idx_entries_normalized_headword_trgm", extensionSchema)

	if err := tx.Exec("SELECT set_config('search_path', ?, true)", secondSchema).Error; err != nil {
		t.Fatalf("set_config(search_path=%q) error = %v; want nil", secondSchema, err)
	}
	if err := RunMigration(tx, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration(second schema=%q) error = %v; want nil", secondSchema, err)
	}
	assertMigrationTablesExist(t, tx)
	assertExpectedIndexesExist(t, tx)
	assertExtensionSchema(t, tx, "pg_trgm", extensionSchema)
	assertCurrentSchemaIndexUsesOperatorClassSchema(t, tx, "entries", "idx_entries_normalized_headword_trgm", extensionSchema)
}

func TestRunMigration_PostgresIntegration_RepairsByDefaultIdentity(t *testing.T) {
	db := newPostgresIntegrationDB(t)

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration() error = %v; want nil", err)
	}

	firstRunID := createImportRun(t, db, "identity-repair", "before-one", time.Now().UTC())
	firstEntry := model.Entry{
		Headword:           "entry-identity-repair-one",
		NormalizedHeadword: "entryidentityrepairone",
		Pos:                model.POSNoun,
		EtymologyIndex:     0,
		SourceRunID:        firstRunID,
	}
	if err := db.Create(&firstEntry).Error; err != nil {
		t.Fatalf("Create(first repair entry) error = %v; want nil", err)
	}

	secondRunID := createImportRun(t, db, "identity-repair", "before-two", time.Now().UTC())
	secondEntry := model.Entry{
		Headword:           "entry-identity-repair-two",
		NormalizedHeadword: "entryidentityrepairtwo",
		Pos:                model.POSNoun,
		EtymologyIndex:     0,
		SourceRunID:        secondRunID,
	}
	if err := db.Create(&secondEntry).Error; err != nil {
		t.Fatalf("Create(second repair entry) error = %v; want nil", err)
	}

	if err := db.Exec(`ALTER TABLE public.entries ALTER COLUMN id SET GENERATED BY DEFAULT`).Error; err != nil {
		t.Fatalf("ALTER TABLE entries.id SET GENERATED BY DEFAULT error = %v; want nil", err)
	}
	if err := db.Exec(`SELECT setval(pg_get_serial_sequence('public.entries', 'id'), 1, true)`).Error; err != nil {
		t.Fatalf("setval(entries_id_seq) error = %v; want nil", err)
	}

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration() repair rerun error = %v; want nil", err)
	}

	state, err := loadIdentityColumnState(db, "entries")
	if err != nil {
		t.Fatalf("loadIdentityColumnState(entries) error = %v; want nil", err)
	}
	if issue := identityColumnStateIssue("entries", "public", state); issue != "" {
		t.Fatalf("identityColumnStateIssue(entries) = %q; want repaired GENERATED ALWAYS state", issue)
	}

	repairedRunID := createImportRun(t, db, "identity-repair", "after", time.Now().UTC())
	repairedEntry := model.Entry{
		Headword:           "entry-identity-repair-after",
		NormalizedHeadword: "entryidentityrepairafter",
		Pos:                model.POSNoun,
		EtymologyIndex:     0,
		SourceRunID:        repairedRunID,
	}
	if err := db.Create(&repairedEntry).Error; err != nil {
		t.Fatalf("Create(repaired entry) error = %v; want nil", err)
	}
	if repairedEntry.ID <= secondEntry.ID {
		t.Fatalf("repaired entries.id = %d; want value greater than pre-repair max %d", repairedEntry.ID, secondEntry.ID)
	}
}

func mustIndexDefinitionTarget(t *testing.T, expectations []sqlIndexDefinitionTarget, indexName string) sqlIndexDefinitionTarget {
	t.Helper()

	for _, target := range expectations {
		if target.IndexName == indexName {
			return target
		}
	}

	t.Fatalf("index definition target %q not found", indexName)
	return sqlIndexDefinitionTarget{}
}

func equalIndexTargetAndExpectation(left indexTarget, right indexExpectation) bool {
	return left.TableName == right.TableName && left.IndexName == right.IndexName
}

func equalIndexDefinitionTargets(left, right sqlIndexDefinitionTarget) bool {
	if left.TableName != right.TableName || left.IndexName != right.IndexName || left.Unique != right.Unique || left.Method != right.Method || left.Predicate != right.Predicate {
		return false
	}

	return slices.Equal(left.Columns, right.Columns)
}

func mustLoadGORMManagedIndexDefinitions(t *testing.T) []sqlIndexDefinitionTarget {
	t.Helper()

	definitions, err := loadGORMManagedIndexDefinitions()
	if err != nil {
		t.Fatalf("loadGORMManagedIndexDefinitions() error = %v; want nil", err)
	}

	return definitions
}

func buildCanonicalIndexDefinition(target sqlIndexDefinitionTarget, schemaName string) string {
	var builder strings.Builder
	builder.WriteString("CREATE ")
	if target.Unique {
		builder.WriteString("UNIQUE ")
	}
	builder.WriteString("INDEX ")
	builder.WriteString(target.IndexName)
	builder.WriteString(" ON ")
	if strings.TrimSpace(schemaName) != "" {
		builder.WriteString(schemaName)
		builder.WriteByte('.')
	}
	builder.WriteString(target.TableName)
	builder.WriteString(" USING ")
	builder.WriteString(target.Method)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(target.Columns, ", "))
	builder.WriteByte(')')
	if strings.TrimSpace(target.Predicate) != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(target.Predicate)
	}

	return builder.String()
}

func migrationTargetTableNames() []string {
	tables := make([]string, 0, len(migrationTargets))
	for _, target := range migrationTargets {
		tables = append(tables, target.TableName)
	}

	return tables
}

func deriveIdentityManagedTablesFromModelPrimaryKeyShape(t *testing.T) []string {
	t.Helper()

	cache := &sync.Map{}
	naming := gormschema.NamingStrategy{}
	tables := make([]string, 0, len(migrationTargets))
	for _, target := range migrationTargets {
		parsedSchema, err := gormschema.Parse(target.Model, cache, naming)
		if err != nil {
			t.Fatalf("schema.Parse(%T) error = %v; want nil", target.Model, err)
		}
		if parsedSchema.Table != target.TableName {
			t.Fatalf("schema.Parse(%T) table = %q; want %q", target.Model, parsedSchema.Table, target.TableName)
		}
		if len(parsedSchema.PrimaryFields) != 1 {
			t.Fatalf("%s primary field count = %d; want 1 to classify identity-managed tables", target.TableName, len(parsedSchema.PrimaryFields))
		}
		if parsedSchema.PrimaryFields[0].DBName == "id" {
			tables = append(tables, target.TableName)
		}
	}

	return tables
}

func sourceDerivedIndexExpectations(t *testing.T) []indexExpectation {
	t.Helper()

	definitions := append([]sqlIndexDefinitionTarget{}, mustLoadGORMManagedIndexDefinitions(t)...)
	definitions = append(definitions, deriveSQLManagedIndexDefinitionsFromEmbeddedSQL(t)...)

	return indexExpectationsFromDefinitionTargets(definitions)
}

func indexExpectationsFromDefinitionTargets(targets []sqlIndexDefinitionTarget) []indexExpectation {
	expectations := make([]indexExpectation, 0, len(targets))
	for _, target := range targets {
		expectations = append(expectations, indexExpectation{TableName: target.TableName, IndexName: target.IndexName})
	}

	return expectations
}

func deriveSQLManagedIndexDefinitionsFromEmbeddedSQL(t *testing.T) []sqlIndexDefinitionTarget {
	t.Helper()

	parsers := map[string]func(*testing.T, string) []sqlIndexDefinitionTarget{
		"sql/002_partial_indexes.sql":    parseCreateIndexStatementsFromEmbeddedSQLFile,
		"sql/003_expression_indexes.sql": parseCreateIndexStatementsFromEmbeddedSQLFile,
		"sql/004_gin_indexes.sql":        parseGINIndexDefinitionsFromEmbeddedSQLFile,
	}

	definitions := make([]sqlIndexDefinitionTarget, 0, len(sqlManagedIndexDefinitions))
	for _, fileName := range discoverEmbeddedSQLIndexDefinitionFiles(t) {
		parser, ok := parsers[fileName]
		if !ok {
			t.Fatalf("embedded SQL index-defining file %q is missing parser coverage", fileName)
		}
		definitions = append(definitions, parser(t, mustReadEmbeddedSQL(t, fileName))...)
	}

	return definitions
}

func mustListEmbeddedSQLFiles(t *testing.T) []string {
	t.Helper()

	files, err := listEmbeddedSQLFiles()
	if err != nil {
		t.Fatalf("listEmbeddedSQLFiles() error = %v; want nil", err)
	}

	return files
}

func discoverEmbeddedSQLIndexDefinitionFiles(t *testing.T) []string {
	t.Helper()

	discovered := make([]string, 0, len(embeddedSQLIndexDefinitionAllowList))
	for _, fileName := range mustListEmbeddedSQLFiles(t) {
		if !embeddedSQLDefinesIndex(mustReadEmbeddedSQL(t, fileName)) {
			continue
		}
		discovered = append(discovered, fileName)
	}

	if !slices.Equal(discovered, embeddedSQLIndexDefinitionAllowList) {
		t.Fatalf("embedded SQL index-defining files = %v; want %v", discovered, embeddedSQLIndexDefinitionAllowList)
	}

	return discovered
}

func embeddedSQLDefinesIndex(sqlText string) bool {
	return createIndexStatementPattern.MatchString(sqlText)
}

func parseCreateIndexStatementsFromEmbeddedSQLFile(t *testing.T, sqlText string) []sqlIndexDefinitionTarget {
	t.Helper()

	statements := splitEmbeddedSQLStatements(sqlText)
	definitions := make([]sqlIndexDefinitionTarget, 0, len(statements))
	for _, statement := range statements {
		definitions = append(definitions, parseCreateIndexStatementFromSQL(t, statement))
	}

	return definitions
}

func parseGINIndexDefinitionsFromEmbeddedSQLFile(t *testing.T, sqlText string) []sqlIndexDefinitionTarget {
	t.Helper()

	formatCalls := extractPostgresFormatCalls(t, sqlText)
	currentSchemaTables := make(map[string]string)
	definitions := make([]sqlIndexDefinitionTarget, 0, len(formatCalls))
	for _, formatCall := range formatCalls {
		if tableName, ok := postgresFormatCallCurrentSchemaTable(formatCall); ok {
			if formatCall.AssignmentTarget != "" {
				currentSchemaTables[formatCall.AssignmentTarget] = tableName
			}
			continue
		}
		if !createIndexStatementPattern.MatchString(formatCall.Template) {
			continue
		}

		renderedStatement := renderPostgresFormatCall(t, formatCall, currentSchemaTables)
		definitions = append(definitions, parseCreateIndexStatementFromSQL(t, renderedStatement))
	}

	if len(definitions) == 0 {
		t.Fatalf("embedded SQL = %q; want at least one CREATE INDEX template", sqlText)
	}

	return definitions
}

func extractPostgresFormatCalls(t *testing.T, sqlText string) []postgresFormatCall {
	t.Helper()

	formatCalls := make([]postgresFormatCall, 0)
	inStringLiteral := false
	for i := 0; i < len(sqlText); i++ {
		if inStringLiteral {
			if sqlText[i] != '\'' {
				continue
			}
			if i+1 < len(sqlText) && sqlText[i+1] == '\'' {
				i++
				continue
			}
			inStringLiteral = false
			continue
		}

		if sqlText[i] == '\'' {
			inStringLiteral = true
			continue
		}
		if i+len("format(") > len(sqlText) || !strings.EqualFold(sqlText[i:i+len("format(")], "format(") {
			continue
		}
		if i > 0 && isSQLIdentifierByte(sqlText[i-1]) {
			continue
		}

		callEnd, ok := findMatchingSQLParen(sqlText, i+len("format"))
		if !ok {
			t.Fatalf("format() call starting at byte %d has unmatched parentheses", i)
		}

		arguments := splitSQLTopLevelCommaList(sqlText[i+len("format(") : callEnd])
		if len(arguments) == 0 {
			t.Fatalf("format() call starting at byte %d has no arguments", i)
		}

		template, ok := extractSQLStringLiteral(arguments[0])
		if ok {
			formatCalls = append(formatCalls, postgresFormatCall{
				Template:         template,
				Arguments:        trimSQLArguments(arguments[1:]),
				AssignmentTarget: extractSQLAssignmentTarget(sqlText[:i]),
			})
		}

		i = callEnd
	}

	return formatCalls
}

func postgresFormatCallCurrentSchemaTable(formatCall postgresFormatCall) (string, bool) {
	if normalizeIndexDefinition(formatCall.Template) != "%i.%i" {
		return "", false
	}
	if len(formatCall.Arguments) != 2 || normalizeIndexDefinition(formatCall.Arguments[0]) != "current_schema()" {
		return "", false
	}

	tableName, ok := extractSQLStringLiteral(formatCall.Arguments[1])
	if !ok || strings.TrimSpace(tableName) == "" {
		return "", false
	}

	return tableName, true
}

func renderPostgresFormatCall(t *testing.T, formatCall postgresFormatCall, currentSchemaTables map[string]string) string {
	t.Helper()

	var builder strings.Builder
	argumentIndex := 0
	for i := 0; i < len(formatCall.Template); i++ {
		if formatCall.Template[i] != '%' {
			builder.WriteByte(formatCall.Template[i])
			continue
		}
		if i+1 >= len(formatCall.Template) {
			t.Fatalf("format template %q ends with dangling %%", formatCall.Template)
		}
		if formatCall.Template[i+1] == '%' {
			builder.WriteByte('%')
			i++
			continue
		}
		if argumentIndex >= len(formatCall.Arguments) {
			t.Fatalf("format template %q consumed %d arguments; want at least one more", formatCall.Template, argumentIndex)
		}

		builder.WriteString(renderPostgresFormatArgument(t, formatCall.Template[i+1], formatCall.Arguments[argumentIndex], currentSchemaTables))
		argumentIndex++
		i++
	}

	if argumentIndex != len(formatCall.Arguments) {
		t.Fatalf("format template %q consumed %d arguments; want %d", formatCall.Template, argumentIndex, len(formatCall.Arguments))
	}

	return builder.String()
}

func renderPostgresFormatArgument(t *testing.T, placeholder byte, expression string, currentSchemaTables map[string]string) string {
	t.Helper()

	switch placeholder {
	case 's', 'S':
		tableName, ok := resolveCurrentSchemaTableExpression(expression, currentSchemaTables)
		if !ok {
			t.Fatalf("format argument %q does not resolve to a current_schema-qualified table", expression)
		}
		return "public." + tableName
	case 'i', 'I':
		if normalizeIndexDefinition(expression) == "current_schema()" {
			return "public"
		}
		if identifier, ok := resolvePostgresIdentifierExpression(expression, currentSchemaTables); ok {
			return identifier
		}
		return "placeholder_ident"
	case 'l', 'L':
		literal, ok := extractSQLStringLiteral(expression)
		if !ok {
			return "'placeholder_literal'"
		}
		return "'" + strings.ReplaceAll(literal, "'", "''") + "'"
	default:
		return "placeholder_value"
	}
}

func resolveCurrentSchemaTableExpression(expression string, currentSchemaTables map[string]string) (string, bool) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return "", false
	}
	if tableName, ok := currentSchemaTables[expression]; ok {
		return tableName, true
	}

	formatCall, ok := parsePostgresFormatCallExpression(expression)
	if !ok {
		return "", false
	}

	return postgresFormatCallCurrentSchemaTable(formatCall)
}

func resolvePostgresIdentifierExpression(expression string, currentSchemaTables map[string]string) (string, bool) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return "", false
	}
	if tableName, ok := resolveCurrentSchemaTableExpression(expression, currentSchemaTables); ok {
		return tableName, true
	}

	identifier, ok := extractSQLStringLiteral(expression)
	if ok {
		return identifier, true
	}
	if sqlIdentifierPattern.MatchString(expression) {
		return expression, true
	}

	return "", false
}

func parsePostgresFormatCallExpression(expression string) (postgresFormatCall, bool) {
	expression = strings.TrimSpace(expression)
	if !strings.HasPrefix(strings.ToLower(expression), "format(") {
		return postgresFormatCall{}, false
	}

	callEnd, ok := findMatchingSQLParen(expression, len("format"))
	if !ok || strings.TrimSpace(expression[callEnd+1:]) != "" {
		return postgresFormatCall{}, false
	}

	arguments := splitSQLTopLevelCommaList(expression[len("format("):callEnd])
	if len(arguments) == 0 {
		return postgresFormatCall{}, false
	}

	template, ok := extractSQLStringLiteral(arguments[0])
	if !ok {
		return postgresFormatCall{}, false
	}

	return postgresFormatCall{Template: template, Arguments: trimSQLArguments(arguments[1:])}, true
}

func findMatchingSQLParen(value string, openParenIndex int) (int, bool) {
	if openParenIndex < 0 || openParenIndex >= len(value) || value[openParenIndex] != '(' {
		return 0, false
	}

	depth := 0
	inStringLiteral := false
	for i := openParenIndex; i < len(value); i++ {
		if inStringLiteral {
			if value[i] != '\'' {
				continue
			}
			if i+1 < len(value) && value[i+1] == '\'' {
				i++
				continue
			}
			inStringLiteral = false
			continue
		}

		switch value[i] {
		case '\'':
			inStringLiteral = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}

	return 0, false
}

func splitSQLTopLevelCommaList(value string) []string {
	parts := make([]string, 0, 4)
	start := 0
	depth := 0
	inStringLiteral := false
	for i := 0; i < len(value); i++ {
		if inStringLiteral {
			if value[i] != '\'' {
				continue
			}
			if i+1 < len(value) && value[i+1] == '\'' {
				i++
				continue
			}
			inStringLiteral = false
			continue
		}

		switch value[i] {
		case '\'':
			inStringLiteral = true
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(value[start:i]))
				start = i + 1
			}
		}
	}

	parts = append(parts, strings.TrimSpace(value[start:]))
	return parts
}

func trimSQLArguments(arguments []string) []string {
	trimmed := make([]string, len(arguments))
	for i, argument := range arguments {
		trimmed[i] = strings.TrimSpace(argument)
	}

	return trimmed
}

func extractSQLAssignmentTarget(prefix string) string {
	prefix = strings.TrimRight(prefix, " \t\r\n")
	boundary := strings.LastIndexAny(prefix, ";\n")
	if boundary >= 0 {
		prefix = prefix[boundary+1:]
	}

	match := sqlAssignmentTargetPattern.FindStringSubmatch(prefix)
	if len(match) != 2 {
		return ""
	}

	return match[1]
}

func extractSQLStringLiteral(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 2 || trimmed[0] != '\'' || trimmed[len(trimmed)-1] != '\'' {
		return "", false
	}

	return strings.ReplaceAll(trimmed[1:len(trimmed)-1], "''", "'"), true
}

func isSQLIdentifierByte(value byte) bool {
	return value == '_' || value >= '0' && value <= '9' || value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func splitEmbeddedSQLStatements(sqlText string) []string {
	statements := strings.Split(sqlText, ";")
	trimmedStatements := make([]string, 0, len(statements))
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		trimmedStatements = append(trimmedStatements, statement)
	}

	return trimmedStatements
}

func parseCreateIndexStatementFromSQL(t *testing.T, statement string) sqlIndexDefinitionTarget {
	t.Helper()

	normalized := normalizeIndexDefinition(statement)
	unique := false
	prefix := "create index if not exists "
	if strings.HasPrefix(normalized, "create unique index if not exists ") {
		unique = true
		prefix = "create unique index if not exists "
	}
	if !strings.HasPrefix(normalized, prefix) {
		t.Fatalf("embedded SQL statement = %q; want CREATE INDEX IF NOT EXISTS form", statement)
	}

	remainder := strings.TrimPrefix(normalized, prefix)
	indexName, remainder, ok := strings.Cut(remainder, " on ")
	if !ok {
		t.Fatalf("embedded SQL statement = %q; want ON clause", statement)
	}

	tableName, remainder, ok := strings.Cut(remainder, " ")
	if !ok {
		t.Fatalf("embedded SQL statement = %q; want table and definition body", statement)
	}

	method := "btree"
	if strings.HasPrefix(remainder, "using ") {
		remainder = strings.TrimPrefix(remainder, "using ")
		method, remainder, ok = strings.Cut(remainder, " ")
		if !ok {
			t.Fatalf("embedded SQL statement = %q; want method body", statement)
		}
	}

	columns, trailing, ok := splitIndexDefinition(remainder)
	if !ok {
		t.Fatalf("embedded SQL statement = %q; want index columns", statement)
	}
	for i, column := range columns {
		columns[i] = normalizeIndexExpression(column)
	}

	predicate := ""
	trailing = strings.TrimSpace(trailing)
	if trailing != "" {
		if !strings.HasPrefix(trailing, "where ") {
			t.Fatalf("embedded SQL statement = %q; want optional WHERE clause", statement)
		}
		predicate = normalizeIndexPredicate(strings.TrimPrefix(trailing, "where "))
	}

	return sqlIndexDefinitionTarget{
		TableName: tableName,
		IndexName: indexName,
		Unique:    unique,
		Method:    strings.ToLower(method),
		Columns:   columns,
		Predicate: predicate,
	}
}

func extractIdentityManagedTablesFromEmbeddedSQL(t *testing.T) []string {
	t.Helper()

	sqlText := mustReadEmbeddedSQL(t, "sql/005_identity_columns.sql")
	arrayMatch := regexp.MustCompile(`(?s)FOREACH\s+target_table\s+IN\s+ARRAY\s+ARRAY\[(.*?)\]\s+LOOP`).FindStringSubmatch(sqlText)
	if len(arrayMatch) != 2 {
		t.Fatalf("005_identity_columns.sql = %q; want FOREACH target_table IN ARRAY ARRAY[...] LOOP", sqlText)
	}

	tableMatches := regexp.MustCompile(`'([^']+)'`).FindAllStringSubmatch(arrayMatch[1], -1)
	if len(tableMatches) == 0 {
		t.Fatalf("005_identity_columns.sql = %q; want quoted identity-managed table names", sqlText)
	}

	tables := make([]string, 0, len(tableMatches))
	for _, match := range tableMatches {
		tables = append(tables, match[1])
	}

	return tables
}

func sortedIndexTargetKeys(targets []indexTarget) []string {
	keys := make([]string, 0, len(targets))
	for _, target := range targets {
		keys = append(keys, target.TableName+"\x00"+target.IndexName)
	}
	slices.Sort(keys)

	return keys
}

func sortedIndexExpectationKeys(expectations []indexExpectation) []string {
	keys := make([]string, 0, len(expectations))
	for _, expectation := range expectations {
		keys = append(keys, expectation.TableName+"\x00"+expectation.IndexName)
	}
	slices.Sort(keys)

	return keys
}

func sortedIndexDefinitionKeys(targets []sqlIndexDefinitionTarget) []string {
	keys := make([]string, 0, len(targets))
	for _, target := range targets {
		columns := make([]string, len(target.Columns))
		for i, column := range target.Columns {
			columns[i] = normalizeIndexExpression(column)
		}
		keys = append(keys, strings.Join([]string{
			target.TableName,
			target.IndexName,
			fmt.Sprintf("%t", target.Unique),
			strings.ToLower(strings.TrimSpace(target.Method)),
			strings.Join(columns, "\x01"),
			normalizeIndexPredicate(target.Predicate),
		}, "\x00"))
	}
	slices.Sort(keys)

	return keys
}

func newPostgresIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping PostgreSQL integration test in short mode")
	}

	dsn := strings.TrimSpace(os.Getenv(postgresIntegrationDSNEnv))
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL integration tests", postgresIntegrationDSNEnv)
	}

	resetConfirmation := strings.TrimSpace(os.Getenv(postgresIntegrationDestructiveResetEnv))
	if resetConfirmation == "" {
		t.Skipf(
			"set %s and %s=%q to run PostgreSQL integration tests",
			postgresIntegrationDSNEnv,
			postgresIntegrationDestructiveResetEnv,
			postgresIntegrationDestructiveResetConfirm,
		)
	}
	if err := validatePostgresIntegrationResetConfirmation(resetConfirmation); err != nil {
		t.Fatalf("unsafe PostgreSQL integration reset opt-in: %v", err)
	}
	if err := validatePostgresIntegrationDSN(dsn); err != nil {
		t.Fatalf("unsafe PostgreSQL integration target DSN: %v", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		t.Fatalf("ping postgres database: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	assertPostgresIntegrationTargetSafe(t, db)
	resetPostgresIntegrationSchema(t, db)
	t.Cleanup(func() {
		_ = resetPostgresIntegrationSchemaNoFail(db)
	})

	return db
}

func resetPostgresIntegrationSchema(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := resetPostgresIntegrationSchemaNoFail(db); err != nil {
		t.Fatalf("reset PostgreSQL integration schema: %v", err)
	}
}

func resetPostgresIntegrationSchemaNoFail(db *gorm.DB) error {
	if err := validatePostgresIntegrationResetConfirmation(os.Getenv(postgresIntegrationDestructiveResetEnv)); err != nil {
		return err
	}

	target, err := inspectPostgresIntegrationTarget(db)
	if err != nil {
		return fmt.Errorf("inspect PostgreSQL integration target: %w", err)
	}

	if err := validatePostgresIntegrationTarget(target); err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range postgresIntegrationResetStatements() {
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("exec %q: %w", stmt, err)
			}
		}

		return nil
	})
}

func assertPostgresIntegrationTargetSafe(t *testing.T, db *gorm.DB) {
	t.Helper()

	target, err := inspectPostgresIntegrationTarget(db)
	if err != nil {
		t.Fatalf("inspect PostgreSQL integration target: %v", err)
	}

	if err := validatePostgresIntegrationTarget(target); err != nil {
		t.Fatalf("unsafe PostgreSQL integration target: %v", err)
	}
}

func inspectPostgresIntegrationTarget(db *gorm.DB) (postgresIntegrationTarget, error) {
	var target postgresIntegrationTarget

	err := db.Raw(`
		SELECT
			current_database() AS database_name,
			current_schema() AS current_schema,
			current_setting('search_path') AS search_path
	`).Scan(&target).Error
	if err != nil {
		return postgresIntegrationTarget{}, err
	}

	return target, nil
}

func assertExtensionSchema(t *testing.T, db *gorm.DB, extensionName, wantSchema string) {
	t.Helper()

	schemaName, err := loadExtensionSchema(db, extensionName)
	if err != nil {
		t.Fatalf("loadExtensionSchema(%q) error = %v; want nil", extensionName, err)
	}
	if schemaName != wantSchema {
		t.Fatalf("extension %s schema = %q; want %q", extensionName, schemaName, wantSchema)
	}
}

func assertCurrentSchemaIndexUsesOperatorClassSchema(t *testing.T, db *gorm.DB, tableName, indexName, wantSchema string) {
	t.Helper()

	indexDefinition, err := loadIndexDefinition(db, tableName, indexName)
	if err != nil {
		t.Fatalf("loadIndexDefinition(%s, %s) error = %v; want nil", tableName, indexName, err)
	}

	normalizedDefinition := strings.ToLower(strings.ReplaceAll(indexDefinition, `"`, ""))
	wantFragment := strings.ToLower(wantSchema) + ".gin_trgm_ops"
	if !strings.Contains(normalizedDefinition, wantFragment) {
		t.Fatalf("index %s(%s) definition = %q; want operator class fragment %q", indexName, tableName, indexDefinition, wantFragment)
	}
}

func loadExtensionSchema(db *gorm.DB, extensionName string) (string, error) {
	var result struct {
		SchemaName string `gorm:"column:schema_name"`
	}

	err := db.Raw(`
		SELECT ns.nspname AS schema_name
		FROM pg_extension ext
		JOIN pg_namespace ns ON ns.oid = ext.extnamespace
		WHERE ext.extname = ?
	`, extensionName).Scan(&result).Error
	if err != nil {
		return "", err
	}

	result.SchemaName = strings.TrimSpace(result.SchemaName)
	if result.SchemaName == "" {
		return "", fmt.Errorf("extension %s not found", extensionName)
	}

	return result.SchemaName, nil
}

func validatePostgresIntegrationDSN(dsn string) error {
	return validatePostgresIntegrationDSNWithRemoteOverride(dsn, os.Getenv(postgresIntegrationRemoteOverrideEnv))
}

func validatePostgresIntegrationDSNWithRemoteOverride(dsn, remoteOverride string) error {
	trimmedDSN := strings.TrimSpace(dsn)
	remoteAllowed := hasPostgresIntegrationRemoteOverride(remoteOverride)
	if !remoteAllowed {
		usesService, err := postgresIntegrationUsesServiceIndirection(trimmedDSN)
		if err != nil {
			return fmt.Errorf("inspect %s for local-only target safety: %w", postgresIntegrationDSNEnv, err)
		}
		if usesService {
			return fmt.Errorf(
				"refusing PostgreSQL integration connection for %s: connection uses service/servicefile indirection via DSN or environment (%s/%s); set %s=%q only for disposable remote CI targets",
				postgresIntegrationDSNEnv,
				postgresIntegrationServiceEnv,
				postgresIntegrationServiceFileEnv,
				postgresIntegrationRemoteOverrideEnv,
				postgresIntegrationRemoteOverrideConfirm,
			)
		}
	}

	config, err := pgx.ParseConfig(trimmedDSN)
	if err != nil {
		return fmt.Errorf("parse %s: %w", postgresIntegrationDSNEnv, err)
	}

	if err := validatePostgresIntegrationDatabaseName(config.Database); err != nil {
		return fmt.Errorf("refusing PostgreSQL integration connection for %s: %w", postgresIntegrationDSNEnv, err)
	}

	if remoteAllowed {
		return nil
	}

	if err := validatePostgresIntegrationLocalHosts(trimmedDSN, config); err != nil {
		return fmt.Errorf(
			"refusing PostgreSQL integration connection for %s: %w; set %s=%q only for disposable remote CI targets",
			postgresIntegrationDSNEnv,
			err,
			postgresIntegrationRemoteOverrideEnv,
			postgresIntegrationRemoteOverrideConfirm,
		)
	}

	return nil
}

func hasPostgresIntegrationRemoteOverride(confirmation string) bool {
	return strings.TrimSpace(confirmation) == postgresIntegrationRemoteOverrideConfirm
}

func postgresIntegrationUsesServiceIndirection(dsn string) (bool, error) {
	usesService, err := postgresIntegrationDSNUsesService(dsn)
	if err != nil {
		return false, err
	}

	if usesService {
		return true, nil
	}

	return postgresIntegrationEnvironmentUsesService(), nil
}

func postgresIntegrationEnvironmentUsesService() bool {
	return strings.TrimSpace(os.Getenv(postgresIntegrationServiceEnv)) != ""
}

func validatePostgresIntegrationLocalHosts(dsn string, config *pgx.ConnConfig) error {
	usesExplicitHostSelection, err := postgresIntegrationUsesExplicitHostSelection(dsn)
	if err != nil {
		return fmt.Errorf("inspect explicit host selection: %w", err)
	}

	for _, host := range postgresIntegrationResolvedHosts(config) {
		if usesExplicitHostSelection {
			if isPostgresIntegrationLocalUnixSocketHost(host) {
				continue
			}

			return fmt.Errorf(
				"host %q is not eligible for the default local PostgreSQL safety path; explicit host selection only auto-allows Unix socket directories %s, and TCP targets including loopback are not implicitly trusted",
				host,
				strings.Join(postgresIntegrationLocalUnixSocketDirs, ", "),
			)
		}

		if !isImplicitLocalPostgresIntegrationHost(host) {
			return fmt.Errorf(
				"resolved default host %q is not a local PostgreSQL target; accepted defaults are empty, localhost, 127.0.0.1, ::1, or Unix socket directories %s",
				host,
				strings.Join(postgresIntegrationLocalUnixSocketDirs, ", "),
			)
		}
	}

	return nil
}

func postgresIntegrationResolvedHosts(config *pgx.ConnConfig) []string {
	hosts := make([]string, 0, 1+len(config.Fallbacks))
	appendUniqueHost := func(host string) {
		trimmedHost := strings.TrimSpace(host)
		for _, existing := range hosts {
			if existing == trimmedHost {
				return
			}
		}
		hosts = append(hosts, trimmedHost)
	}

	appendUniqueHost(config.Host)
	for _, fallback := range config.Fallbacks {
		if fallback == nil {
			continue
		}
		appendUniqueHost(fallback.Host)
	}

	return hosts
}

func postgresIntegrationUsesExplicitHostSelection(dsn string) (bool, error) {
	usesExplicitHostSelection, err := postgresIntegrationDSNUsesExplicitHostSelection(dsn)
	if err != nil {
		return false, err
	}

	if usesExplicitHostSelection {
		return true, nil
	}

	return strings.TrimSpace(os.Getenv(postgresIntegrationHostEnv)) != "", nil
}

func postgresIntegrationDSNUsesExplicitHostSelection(dsn string) (bool, error) {
	if strings.Contains(dsn, "://") {
		parsedURL, err := url.Parse(dsn)
		if err != nil {
			return false, err
		}

		if strings.TrimSpace(parsedURL.Host) != "" {
			return true, nil
		}

		return strings.TrimSpace(parsedURL.Query().Get("host")) != "", nil
	}

	hostValue, ok, err := postgresIntegrationKeywordValueValue(dsn, "host")
	if err != nil {
		return false, err
	}

	return ok && strings.TrimSpace(hostValue) != "", nil
}

func isImplicitLocalPostgresIntegrationHost(host string) bool {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		return true
	}

	switch strings.ToLower(strings.Trim(trimmedHost, "[]")) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		if !filepath.IsAbs(trimmedHost) {
			return false
		}

		_, ok := postgresIntegrationLocalUnixSocketDirSet[filepath.Clean(trimmedHost)]
		return ok
	}
}

func isPostgresIntegrationLocalUnixSocketHost(host string) bool {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" || !filepath.IsAbs(trimmedHost) {
		return false
	}

	_, ok := postgresIntegrationLocalUnixSocketDirSet[filepath.Clean(trimmedHost)]
	return ok
}

func postgresIntegrationDSNUsesService(dsn string) (bool, error) {
	if strings.Contains(dsn, "://") {
		parsedURL, err := url.Parse(dsn)
		if err != nil {
			return false, err
		}

		query := parsedURL.Query()
		return query.Has("service") || query.Has("servicefile"), nil
	}

	keys, err := postgresIntegrationKeywordValueKeys(dsn)
	if err != nil {
		return false, err
	}

	_, hasService := keys["service"]
	_, hasServiceFile := keys["servicefile"]
	return hasService || hasServiceFile, nil
}

func postgresIntegrationKeywordValueValue(dsn, targetKey string) (string, bool, error) {
	remaining := strings.TrimSpace(dsn)
	targetKey = strings.ToLower(strings.TrimSpace(targetKey))
	for remaining != "" {
		eqIndex := strings.IndexRune(remaining, '=')
		if eqIndex < 0 {
			return "", false, errors.New("invalid keyword/value")
		}

		key := strings.TrimSpace(remaining[:eqIndex])
		if key == "" {
			return "", false, errors.New("invalid keyword/value")
		}

		value, next, err := readPostgresIntegrationKeywordValue(strings.TrimLeft(remaining[eqIndex+1:], " \t\n\r\v\f"))
		if err != nil {
			return "", false, err
		}

		if strings.ToLower(key) == targetKey {
			return value, true, nil
		}

		remaining = strings.TrimLeft(next, " \t\n\r\v\f")
	}

	return "", false, nil
}

func postgresIntegrationKeywordValueKeys(dsn string) (map[string]struct{}, error) {
	keys := make(map[string]struct{})
	remaining := strings.TrimSpace(dsn)
	for remaining != "" {
		eqIndex := strings.IndexRune(remaining, '=')
		if eqIndex < 0 {
			return nil, errors.New("invalid keyword/value")
		}

		key := strings.TrimSpace(remaining[:eqIndex])
		if key == "" {
			return nil, errors.New("invalid keyword/value")
		}

		next, err := skipPostgresIntegrationKeywordValue(strings.TrimLeft(remaining[eqIndex+1:], " \t\n\r\v\f"))
		if err != nil {
			return nil, err
		}

		keys[strings.ToLower(key)] = struct{}{}
		remaining = strings.TrimLeft(next, " \t\n\r\v\f")
	}

	return keys, nil
}

func skipPostgresIntegrationKeywordValue(remaining string) (string, error) {
	_, next, err := readPostgresIntegrationKeywordValue(remaining)
	return next, err
}

func readPostgresIntegrationKeywordValue(remaining string) (string, string, error) {
	if remaining == "" {
		return "", "", nil
	}

	if remaining[0] != '\'' {
		escaped := false
		var builder strings.Builder
		for i := 0; i < len(remaining); i++ {
			switch {
			case escaped:
				builder.WriteByte(remaining[i])
				escaped = false
			case remaining[i] == '\\':
				escaped = true
			case isPostgresIntegrationKeywordValueSpace(remaining[i]):
				return builder.String(), remaining[i:], nil
			default:
				builder.WriteByte(remaining[i])
			}
		}
		if escaped {
			return "", "", errors.New("invalid backslash")
		}

		return builder.String(), "", nil
	}

	escaped := false
	var builder strings.Builder
	for i := 1; i < len(remaining); i++ {
		switch {
		case escaped:
			builder.WriteByte(remaining[i])
			escaped = false
		case remaining[i] == '\\':
			escaped = true
		case remaining[i] == '\'':
			return builder.String(), remaining[i+1:], nil
		default:
			builder.WriteByte(remaining[i])
		}
	}

	return "", "", errors.New("unterminated quoted string in connection info string")
}

func isPostgresIntegrationKeywordValueSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func validatePostgresIntegrationResetConfirmation(confirmation string) error {
	if strings.TrimSpace(confirmation) != postgresIntegrationDestructiveResetConfirm {
		return fmt.Errorf(
			"refusing PostgreSQL integration cleanup without %s=%q",
			postgresIntegrationDestructiveResetEnv,
			postgresIntegrationDestructiveResetConfirm,
		)
	}

	return nil
}

func validatePostgresIntegrationDatabaseName(name string) error {
	databaseName := strings.TrimSpace(name)
	normalized := strings.ToLower(databaseName)
	if _, ok := postgresIntegrationDisposableDatabaseNameSet[normalized]; ok {
		return nil
	}

	return fmt.Errorf(
		"database %q is not an accepted disposable test database; accepted names: %s",
		databaseName,
		strings.Join(postgresIntegrationDisposableDatabaseNames, ", "),
	)
}

func validatePostgresIntegrationTarget(target postgresIntegrationTarget) error {
	if err := validatePostgresIntegrationDatabaseName(target.DatabaseName); err != nil {
		return fmt.Errorf("refusing PostgreSQL integration cleanup: %w", err)
	}

	currentSchema := strings.TrimSpace(target.CurrentSchema)
	if currentSchema != "public" {
		return fmt.Errorf("refusing PostgreSQL integration cleanup because current_schema()=%q; want %q", currentSchema, "public")
	}

	searchPath := normalizePostgresSearchPath(target.SearchPath)
	if searchPath != "public" {
		return fmt.Errorf("refusing PostgreSQL integration cleanup because search_path=%q; want %q", searchPath, "public")
	}

	return nil
}

func normalizePostgresSearchPath(searchPath string) string {
	parts := strings.Split(searchPath, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(strings.Trim(part, `"`))
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}

	return strings.Join(normalized, ",")
}

func postgresIntegrationExpectedResetStatements() []string {
	statements := make([]string, 0, len(postgresIntegrationExpectedTables))
	for i := len(postgresIntegrationExpectedTables) - 1; i >= 0; i-- {
		statements = append(statements, fmt.Sprintf("DROP TABLE IF EXISTS public.%s RESTRICT", postgresIntegrationExpectedTables[i]))
	}

	return statements
}

func postgresIntegrationResetStatements() []string {
	statements := make([]string, 0, len(migrationTargets))
	for i := len(migrationTargets) - 1; i >= 0; i-- {
		statements = append(statements, fmt.Sprintf("DROP TABLE IF EXISTS public.%s RESTRICT", migrationTargets[i].TableName))
	}

	return statements
}

func assertMigrationTablesExist(t *testing.T, db *gorm.DB) {
	t.Helper()

	if len(postgresIntegrationExpectedTables) != 15 {
		t.Fatalf("postgresIntegrationExpectedTables = %d; want 15-table contract", len(postgresIntegrationExpectedTables))
	}

	actualTables, err := loadCurrentSchemaTableNames(db)
	if err != nil {
		t.Fatalf("loadCurrentSchemaTableNames() error = %v; want nil", err)
	}

	missingTables, unexpectedTables := currentSchemaTableContractDiff(actualTables)
	if len(missingTables) > 0 || len(unexpectedTables) > 0 {
		sortedActualTables := append([]string(nil), actualTables...)
		expectedTables := append([]string(nil), postgresIntegrationExpectedTables...)
		slices.Sort(sortedActualTables)
		slices.Sort(expectedTables)
		t.Fatalf("current-schema tables mismatch: missing = %v, unexpected = %v, actual tables = %v, want = %v", missingTables, unexpectedTables, sortedActualTables, expectedTables)
	}
}

func assertExpectedIndexesExist(t *testing.T, db *gorm.DB) {
	t.Helper()

	indexCatalog, err := loadCurrentSchemaIndexCatalog(db)
	if err != nil {
		t.Fatalf("loadCurrentSchemaIndexCatalog() error = %v; want nil", err)
	}

	filteredIndexCatalog, missingIndexes, unexpectedIndexes := currentSchemaIndexContractDiff(indexCatalog)
	actualIndexKeys := sortedIndexCatalogKeys(filteredIndexCatalog)
	expectedIndexKeys := sortedIndexExpectationKeys(postgresIntegrationExpectedIndexes)
	if len(missingIndexes) > 0 || len(unexpectedIndexes) > 0 {
		t.Fatalf(
			"current-schema non-implicit indexes mismatch: missing = %v, unexpected = %v, actual indexes = %v, want = %v",
			formatManagedIndexKeys(missingIndexes),
			formatManagedIndexKeys(unexpectedIndexes),
			formatManagedIndexKeys(actualIndexKeys),
			formatManagedIndexKeys(expectedIndexKeys),
		)
	}

	assertExpectedIndexDefinitions(t, filteredIndexCatalog, postgresIntegrationExpectedSQLManagedIndexDefinitions)
	assertExpectedIndexDefinitions(t, filteredIndexCatalog, postgresIntegrationExpectedGORMIndexDefinitions)
}

func assertExpectedIndexDefinitions(t *testing.T, indexCatalog map[string]postgresIntegrationIndexCatalog, expectations []sqlIndexDefinitionTarget) {
	t.Helper()

	for _, target := range expectations {
		index, ok := indexCatalog[postgresIntegrationIndexKey(target.TableName, target.IndexName)]
		if !ok {
			t.Fatalf("index %s(%s) not found in catalog", target.IndexName, target.TableName)
		}

		if !postgresIntegrationIndexMatchesExpectation(index, target) {
			t.Fatalf(
				"index %s(%s) catalog = {unique:%t method:%q columns:%v predicate:%q}; want {unique:%t method:%q columns:%v predicate:%q}",
				target.IndexName,
				target.TableName,
				index.IsUnique,
				index.Method,
				[]string(index.Columns),
				index.Predicate,
				target.Unique,
				target.Method,
				target.Columns,
				target.Predicate,
			)
		}
	}
}

func managedObjectContractDiff(expected, actual []string) ([]string, []string) {
	expectedSet := make(map[string]struct{}, len(expected))
	for _, value := range expected {
		expectedSet[value] = struct{}{}
	}

	actualSet := make(map[string]struct{}, len(actual))
	for _, value := range actual {
		actualSet[value] = struct{}{}
	}

	missing := make([]string, 0)
	for _, value := range expected {
		if _, ok := actualSet[value]; !ok {
			missing = append(missing, value)
		}
	}

	unexpected := make([]string, 0)
	for _, value := range actual {
		if _, ok := expectedSet[value]; !ok {
			unexpected = append(unexpected, value)
		}
	}

	slices.Sort(missing)
	slices.Sort(unexpected)
	return missing, unexpected
}

func currentSchemaTableContractDiff(actualTables []string) ([]string, []string) {
	return managedObjectContractDiff(postgresIntegrationExpectedTables, actualTables)
}

func currentSchemaIndexContractDiff(indexCatalog map[string]postgresIntegrationIndexCatalog) (map[string]postgresIntegrationIndexCatalog, []string, []string) {
	filteredIndexCatalog := filterAcceptedImplicitIndexes(indexCatalog)
	actualIndexKeys := sortedIndexCatalogKeys(filteredIndexCatalog)
	expectedIndexKeys := sortedIndexExpectationKeys(postgresIntegrationExpectedIndexes)
	missing, unexpected := managedObjectContractDiff(expectedIndexKeys, actualIndexKeys)

	return filteredIndexCatalog, missing, unexpected
}

func filterAcceptedImplicitIndexes(indexCatalog map[string]postgresIntegrationIndexCatalog) map[string]postgresIntegrationIndexCatalog {
	filtered := make(map[string]postgresIntegrationIndexCatalog, len(indexCatalog))
	for key, index := range indexCatalog {
		if isAcceptedImplicitIndex(index) {
			continue
		}
		filtered[key] = index
	}

	return filtered
}

func isAcceptedImplicitIndex(index postgresIntegrationIndexCatalog) bool {
	return index.IndexName == index.TableName+"_pkey" && index.IsUnique && strings.EqualFold(index.Method, "btree")
}

func sortedIndexCatalogKeys(indexCatalog map[string]postgresIntegrationIndexCatalog) []string {
	keys := make([]string, 0, len(indexCatalog))
	for key := range indexCatalog {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	return keys
}

func formatManagedIndexKeys(keys []string) []string {
	formatted := make([]string, 0, len(keys))
	for _, key := range keys {
		tableName, indexName, ok := strings.Cut(key, "\x00")
		if !ok {
			formatted = append(formatted, key)
			continue
		}
		formatted = append(formatted, fmt.Sprintf("%s(%s)", indexName, tableName))
	}
	slices.Sort(formatted)

	return formatted
}

func loadCurrentSchemaIndexCatalog(db *gorm.DB) (map[string]postgresIntegrationIndexCatalog, error) {
	var rows []postgresIntegrationIndexCatalog

	err := db.Raw(`
		SELECT
			tab.relname AS table_name,
			idx.relname AS index_name,
			ind.indisunique AS is_unique,
			am.amname AS method,
			COALESCE(pg_get_expr(ind.indpred, ind.indrelid, true), '') AS predicate,
			pg_get_indexdef(ind.indexrelid) AS definition,
			ARRAY(
				SELECT pg_get_indexdef(ind.indexrelid, ord.n, true)
				FROM generate_series(1, ind.indnkeyatts) AS ord(n)
				ORDER BY ord.n
			) AS columns
		FROM pg_index ind
		JOIN pg_class idx ON idx.oid = ind.indexrelid
		JOIN pg_class tab ON tab.oid = ind.indrelid
		JOIN pg_namespace ns ON ns.oid = tab.relnamespace
		JOIN pg_am am ON am.oid = idx.relam
		WHERE ns.nspname = current_schema()
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	indexCatalog := make(map[string]postgresIntegrationIndexCatalog, len(rows))
	for _, row := range rows {
		indexCatalog[postgresIntegrationIndexKey(row.TableName, row.IndexName)] = row
	}

	return indexCatalog, nil
}

func postgresIntegrationIndexKey(tableName, indexName string) string {
	return tableName + "\x00" + indexName
}

func postgresIntegrationIndexMatchesExpectation(index postgresIntegrationIndexCatalog, target sqlIndexDefinitionTarget) bool {
	if index.IsUnique != target.Unique {
		return false
	}
	if !strings.EqualFold(index.Method, target.Method) {
		return false
	}
	if len(index.Columns) != len(target.Columns) {
		return false
	}

	normalizedDefinition := normalizePostgresIntegrationIndexDefinition(index.Definition)

	for i := range index.Columns {
		actualColumn := normalizePostgresIntegrationIndexExpression(index.Columns[i])
		expectedColumn := normalizePostgresIntegrationIndexExpression(target.Columns[i])
		if actualColumn == expectedColumn {
			continue
		}
		if !postgresIntegrationIndexDefinitionCarriesExpectedColumn(normalizedDefinition, expectedColumn) {
			return false
		}
	}

	if strings.TrimSpace(target.Predicate) == "" {
		return strings.TrimSpace(index.Predicate) == ""
	}

	return normalizePostgresIntegrationIndexPredicate(index.Predicate) == normalizePostgresIntegrationIndexPredicate(target.Predicate)
}

func postgresIntegrationIndexDefinitionCarriesExpectedColumn(normalizedDefinition, expectedColumn string) bool {
	for _, fragment := range []string{"_ops", " desc", " asc", " nulls first", " nulls last"} {
		if strings.Contains(expectedColumn, fragment) {
			return strings.Contains(normalizedDefinition, expectedColumn)
		}
	}

	return false
}

func normalizePostgresIntegrationIndexExpression(value string) string {
	normalized := normalizeIndexDefinition(value)
	normalized = stripPostgresIntegrationWrappingParens(normalized)

	if strings.HasPrefix(normalized, "coalesce(") && strings.HasSuffix(normalized, ")") {
		args := splitPostgresIntegrationTopLevelCSV(normalized[len("coalesce(") : len(normalized)-1])
		for i, arg := range args {
			args[i] = normalizePostgresIntegrationIndexExpression(arg)
		}
		return "coalesce(" + strings.Join(args, ", ") + ")"
	}

	return normalized
}

func normalizePostgresIntegrationIndexDefinition(value string) string {
	return normalizeIndexDefinition(value)
}

func normalizePostgresIntegrationIndexPredicate(value string) string {
	normalized := normalizePostgresIntegrationIndexExpression(value)
	if strings.HasSuffix(normalized, " = true") {
		return strings.TrimSpace(strings.TrimSuffix(normalized, " = true"))
	}

	return normalized
}

func splitPostgresIntegrationTopLevelCSV(value string) []string {
	parts := make([]string, 0, 4)
	start := 0
	depth := 0

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(value[start:i]))
				start = i + 1
			}
		}
	}

	parts = append(parts, strings.TrimSpace(value[start:]))
	return parts
}

func stripPostgresIntegrationWrappingParens(value string) string {
	trimmed := strings.TrimSpace(value)
	for len(trimmed) >= 2 && trimmed[0] == '(' && trimmed[len(trimmed)-1] == ')' {
		depth := 0
		balanced := true
		for i := 0; i < len(trimmed); i++ {
			switch trimmed[i] {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 && i != len(trimmed)-1 {
					balanced = false
				}
			}
			if depth < 0 {
				balanced = false
				break
			}
		}

		if !balanced || depth != 0 {
			break
		}

		trimmed = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}

	return trimmed
}

func assertColumnTypes(t *testing.T, db *gorm.DB, expectations []columnExpectation) {
	t.Helper()

	for _, expectation := range expectations {
		metadata, err := loadColumnMetadata(db, expectation.TableName, expectation.ColumnName)
		if err != nil {
			t.Fatalf("loadColumnMetadata(%s, %s) error = %v; want nil", expectation.TableName, expectation.ColumnName, err)
		}

		if !strings.EqualFold(metadata.DataType, expectation.DataType) {
			t.Fatalf("column %s.%s data_type = %q; want %q", expectation.TableName, expectation.ColumnName, metadata.DataType, expectation.DataType)
		}
		if !strings.EqualFold(metadata.UDTName, expectation.UDTName) {
			t.Fatalf("column %s.%s udt_name = %q; want %q", expectation.TableName, expectation.ColumnName, metadata.UDTName, expectation.UDTName)
		}
	}
}

func loadColumnMetadata(db *gorm.DB, tableName, columnName string) (columnMetadata, error) {
	var metadata columnMetadata

	err := db.Raw(`
		SELECT data_type, udt_name
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = ?
		  AND column_name = ?
	`, tableName, columnName).Scan(&metadata).Error
	if err != nil {
		return columnMetadata{}, err
	}
	if metadata.DataType == "" {
		return columnMetadata{}, fmt.Errorf("column %s.%s not found", tableName, columnName)
	}

	return metadata, nil
}

func seedMigrationFixture(t *testing.T, db *gorm.DB, suffix string) migrationFixture {
	t.Helper()

	startedAt := time.Now().UTC()
	importRunIDs := seedMigrationFixtureImportRuns(t, db, suffix, startedAt)
	entry := seedMigrationFixtureEntry(t, db, suffix, importRunIDs[importRunKeyEntry])
	sense := seedMigrationFixtureSense(t, db, entry.ID)

	seedMigrationFixtureGlosses(t, db, suffix, sense.ID, importRunIDs[importRunKeySenseGlossZH])
	seedMigrationFixtureSenseMetadata(t, db, suffix, sense.ID)
	seedMigrationFixturePronunciations(t, db, suffix, entry.ID)
	form := seedMigrationFixtureFormAndRelations(t, db, suffix, entry.ID, sense.ID)
	seedMigrationFixtureSummaryAndSignals(t, db, suffix, entry.ID, sense.ID, importRunIDs)

	return migrationFixture{
		EntryID:      entry.ID,
		SenseID:      sense.ID,
		FormID:       form.ID,
		ImportRunIDs: importRunIDs,
	}
}

func seedMigrationFixtureImportRuns(t *testing.T, db *gorm.DB, suffix string, startedAt time.Time) map[string]int64 {
	t.Helper()

	importRunIDs := make(map[string]int64, len(postgresIntegrationAllImportRunKeys))
	for _, key := range postgresIntegrationAllImportRunKeys {
		importRunIDs[key] = createImportRun(t, db, suffix, key, startedAt)
	}

	return importRunIDs
}

func seedMigrationFixtureEntry(t *testing.T, db *gorm.DB, suffix string, sourceRunID int64) model.Entry {
	t.Helper()

	entry := model.Entry{
		Headword:           "entry-" + suffix,
		NormalizedHeadword: "entry" + suffix,
		Pos:                model.POSNoun,
		EtymologyIndex:     0,
		SourceRunID:        sourceRunID,
	}
	createFixtureRecord(t, db, "entry", &entry)

	return entry
}

func seedMigrationFixtureSense(t *testing.T, db *gorm.DB, entryID int64) model.Sense {
	t.Helper()

	sense := model.Sense{
		EntryID:    entryID,
		SenseOrder: 1,
	}
	createFixtureRecord(t, db, "sense", &sense)

	return sense
}

func seedMigrationFixtureGlosses(t *testing.T, db *gorm.DB, suffix string, senseID, sourceRunID int64) {
	t.Helper()

	senseGlossEN := model.SenseGlossEN{
		SenseID:    senseID,
		GlossOrder: 1,
		TextEN:     "definition-" + suffix,
	}
	createFixtureRecord(t, db, "sense_gloss_en", &senseGlossEN)

	senseGlossZH := model.SenseGlossZH{
		SenseID:     senseID,
		Source:      "wiktionary",
		SourceRunID: sourceRunID,
		GlossOrder:  1,
		TextZHHans:  "释义-" + suffix,
		IsPrimary:   true,
	}
	createFixtureRecord(t, db, "sense_gloss_zh", &senseGlossZH)
}

func seedMigrationFixtureSenseMetadata(t *testing.T, db *gorm.DB, suffix string, senseID int64) {
	t.Helper()

	senseLabel := model.SenseLabel{
		SenseID:    senseID,
		LabelType:  model.LabelTypeGrammar,
		LabelCode:  model.GrammarLabelTransitive,
		LabelOrder: 1,
	}
	createFixtureRecord(t, db, "sense_label", &senseLabel)

	senseExample := model.SenseExample{
		SenseID:      senseID,
		Source:       "wiktionary",
		ExampleOrder: 1,
		SentenceEN:   "example sentence " + suffix,
	}
	createFixtureRecord(t, db, "sense_example", &senseExample)
}

func seedMigrationFixturePronunciations(t *testing.T, db *gorm.DB, suffix string, entryID int64) {
	t.Helper()

	pronunciationIPA := model.PronunciationIPA{
		EntryID:      entryID,
		AccentCode:   model.AccentBritish,
		IPA:          "ipa-" + suffix,
		IsPrimary:    true,
		DisplayOrder: 1,
	}
	createFixtureRecord(t, db, "pronunciation_ipa", &pronunciationIPA)

	pronunciationAudio := model.PronunciationAudio{
		EntryID:       entryID,
		AccentCode:    model.AccentBritish,
		AudioFilename: "audio-" + suffix + ".mp3",
		IsPrimary:     true,
		DisplayOrder:  1,
	}
	createFixtureRecord(t, db, "pronunciation_audio", &pronunciationAudio)
}

func seedMigrationFixtureFormAndRelations(t *testing.T, db *gorm.DB, suffix string, entryID, senseID int64) model.EntryForm {
	t.Helper()

	formType := "plural"
	form := model.EntryForm{
		EntryID:         entryID,
		FormText:        "forms-" + suffix,
		NormalizedForm:  "forms" + suffix,
		RelationKind:    model.RelationKindForm,
		FormType:        &formType,
		SourceRelations: pq.StringArray{"wiktionary", "plural"},
	}
	createFixtureRecord(t, db, "entry_form", &form)

	entryRelation := model.LexicalRelation{
		EntryID:              entryID,
		RelationType:         model.RelationTypeDerived,
		TargetText:           "derived-" + suffix,
		TargetTextNormalized: "derived" + suffix,
		DisplayOrder:         1,
	}
	createFixtureRecord(t, db, "entry lexical_relation", &entryRelation)

	senseRelation := model.LexicalRelation{
		EntryID:              entryID,
		SenseID:              int64Ptr(senseID),
		RelationType:         model.RelationTypeSynonym,
		TargetText:           "synonym-" + suffix,
		TargetTextNormalized: "synonym" + suffix,
		DisplayOrder:         1,
	}
	createFixtureRecord(t, db, "sense lexical_relation", &senseRelation)

	return form
}

func seedMigrationFixtureSummaryAndSignals(t *testing.T, db *gorm.DB, suffix string, entryID, senseID int64, importRunIDs map[string]int64) {
	t.Helper()

	entrySummary := model.EntrySummaryZH{
		EntryID:     entryID,
		Source:      "summary-source",
		SourceRunID: importRunIDs[importRunKeyEntrySummary],
		SummaryText: "summary-" + suffix,
	}
	createFixtureRecord(t, db, "entry_summary_zh", &entrySummary)

	entryLearningSignal := model.EntryLearningSignal{
		EntryID:        entryID,
		CEFRLevel:      model.CEFRLevelA1,
		CEFRSource:     model.CEFRSourceOxford,
		CEFRRunID:      int64Ptr(importRunIDs[importRunKeyEntryLearningCEFR]),
		OxfordLevel:    model.OxfordLevel3000,
		OxfordRunID:    int64Ptr(importRunIDs[importRunKeyEntryLearningOxford]),
		CETLevel:       model.CETLevel4,
		CETRunID:       int64Ptr(importRunIDs[importRunKeyEntryLearningCET]),
		FrequencyRank:  10,
		FrequencyCount: 100,
		FrequencyRunID: int64Ptr(importRunIDs[importRunKeyEntryLearningFrequency]),
		CollinsStars:   model.CollinsThreeStars,
		CollinsRunID:   int64Ptr(importRunIDs[importRunKeyEntryLearningCollins]),
	}
	createFixtureRecord(t, db, "entry_learning_signal", &entryLearningSignal)

	senseLearningSignal := model.SenseLearningSignal{
		SenseID:     senseID,
		CEFRLevel:   model.CEFRLevelA1,
		CEFRSource:  model.CEFRSourceOxford,
		CEFRRunID:   int64Ptr(importRunIDs[importRunKeySenseLearningCEFR]),
		OxfordLevel: model.OxfordLevel3000,
		OxfordRunID: int64Ptr(importRunIDs[importRunKeySenseLearningOxford]),
	}
	createFixtureRecord(t, db, "sense_learning_signal", &senseLearningSignal)

	entryEtymology := model.EntryEtymology{
		EntryID:          entryID,
		Source:           "wiktionary",
		SourceRunID:      importRunIDs[importRunKeyEntryEtymology],
		EtymologyTextRaw: "etymology-" + suffix,
	}
	createFixtureRecord(t, db, "entry_etymology", &entryEtymology)
}

func createFixtureRecord[T any](t *testing.T, db *gorm.DB, label string, value *T) {
	t.Helper()

	if err := db.Create(value).Error; err != nil {
		t.Fatalf("Create(%s) error = %v; want nil", label, err)
	}
}

func createImportRun(t *testing.T, db *gorm.DB, suffix, key string, startedAt time.Time) int64 {
	t.Helper()

	run := model.ImportRun{
		SourceName:      "integration-" + suffix + "-" + strings.ReplaceAll(key, "_", "-"),
		SourcePath:      "fixtures/" + suffix + "/" + key,
		PipelineVersion: "test",
		Status:          model.ImportRunStatusRunning,
		StartedAt:       startedAt,
	}
	if err := db.Create(&run).Error; err != nil {
		t.Fatalf("Create(import_run=%s) error = %v; want nil", key, err)
	}

	return run.ID
}

func int64Ptr(value int64) *int64 {
	return &value
}

func importRunID(t *testing.T, fixture migrationFixture, key string) int64 {
	t.Helper()

	runID, ok := fixture.ImportRunIDs[key]
	if !ok {
		t.Fatalf("migration fixture import run %q not found", key)
	}

	return runID
}

func assertImportRunsExist(t *testing.T, db *gorm.DB, fixture migrationFixture, keys []string) {
	t.Helper()

	for _, key := range keys {
		assertModelRowCount(t, db, &model.ImportRun{}, 1, "id = ?", importRunID(t, fixture, key))
	}
}

func assertImportRunsDeleteRestricted(t *testing.T, db *gorm.DB, fixture migrationFixture, keys []string) {
	t.Helper()

	for _, key := range keys {
		assertImportRunDeleteRestricted(t, db, importRunID(t, fixture, key), key)
	}
}

func assertImportRunDeleteRestricted(t *testing.T, db *gorm.DB, runID int64, key string) {
	t.Helper()

	err := db.Delete(&model.ImportRun{}, runID).Error
	if err == nil {
		t.Fatalf("Delete(import_run=%s:%d) error = nil; want RESTRICT violation", key, runID)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "foreign key") {
		t.Fatalf("Delete(import_run=%s:%d) error = %v; want foreign key restriction", key, runID, err)
	}

	assertModelRowCount(t, db, &model.ImportRun{}, 1, "id = ?", runID)
}

func assertImportRunsDeleteAllowed(t *testing.T, db *gorm.DB, fixture migrationFixture, keys []string) {
	t.Helper()

	for _, key := range keys {
		assertImportRunDeleteAllowed(t, db, importRunID(t, fixture, key), key)
	}
}

func assertImportRunDeleteAllowed(t *testing.T, db *gorm.DB, runID int64, key string) {
	t.Helper()

	if err := db.Delete(&model.ImportRun{}, runID).Error; err != nil {
		t.Fatalf("Delete(import_run=%s:%d) error = %v; want nil", key, runID, err)
	}

	assertModelRowCount(t, db, &model.ImportRun{}, 0, "id = ?", runID)
}

func assertEntryOwnedCascadeRows(t *testing.T, db *gorm.DB, fixture migrationFixture, want int64) {
	t.Helper()

	assertModelRowCount(t, db, &model.PronunciationIPA{}, want, "entry_id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.PronunciationAudio{}, want, "entry_id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.EntryForm{}, want, "entry_id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.LexicalRelation{}, want, "entry_id = ? AND sense_id IS NULL", fixture.EntryID)
	assertModelRowCount(t, db, &model.EntrySummaryZH{}, want, "entry_id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.EntryLearningSignal{}, want, "entry_id = ?", fixture.EntryID)
	assertModelRowCount(t, db, &model.EntryEtymology{}, want, "entry_id = ?", fixture.EntryID)
}

func assertSenseOwnedCascadeRows(t *testing.T, db *gorm.DB, fixture migrationFixture, want int64) {
	t.Helper()

	assertModelRowCount(t, db, &model.SenseGlossEN{}, want, "sense_id = ?", fixture.SenseID)
	assertModelRowCount(t, db, &model.SenseGlossZH{}, want, "sense_id = ?", fixture.SenseID)
	assertModelRowCount(t, db, &model.SenseLabel{}, want, "sense_id = ?", fixture.SenseID)
	assertModelRowCount(t, db, &model.SenseExample{}, want, "sense_id = ?", fixture.SenseID)
	assertModelRowCount(t, db, &model.LexicalRelation{}, want, "sense_id = ?", fixture.SenseID)
	assertModelRowCount(t, db, &model.SenseLearningSignal{}, want, "sense_id = ?", fixture.SenseID)
}

func loadCurrentSchemaTableNames(db *gorm.DB) ([]string, error) {
	var rows []struct {
		TableName string `gorm:"column:table_name"`
	}

	err := db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema()
		  AND table_type = 'BASE TABLE'
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	tableNames := make([]string, 0, len(rows))
	for _, row := range rows {
		tableNames = append(tableNames, row.TableName)
	}

	return tableNames, nil
}

func assertSchemaRelationExists(t *testing.T, db *gorm.DB, schemaName, relationName string) {
	t.Helper()

	var result struct {
		Exists bool `gorm:"column:exists"`
	}

	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM pg_class cls
			JOIN pg_namespace ns ON ns.oid = cls.relnamespace
			WHERE ns.nspname = ?
			  AND cls.relname = ?
		) AS exists
	`, schemaName, relationName).Scan(&result).Error
	if err != nil {
		t.Fatalf("check relation %s.%s existence: %v", schemaName, relationName, err)
	}
	if !result.Exists {
		t.Fatalf("relation %s.%s missing; want it preserved", schemaName, relationName)
	}
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func createPostgresIntegrationGuardView(t *testing.T, db *gorm.DB, schemaPrefix, viewName, selectSQL string) (string, string) {
	t.Helper()

	guardSchema := fmt.Sprintf("%s_%d", schemaPrefix, time.Now().UnixNano())
	quotedGuardSchema := quotePostgresIdentifier(guardSchema)
	quotedGuardView := quotePostgresIdentifier(viewName)
	t.Cleanup(func() {
		_ = db.Exec("DROP SCHEMA IF EXISTS " + quotedGuardSchema + " CASCADE").Error
	})

	if err := db.Exec("CREATE SCHEMA " + quotedGuardSchema).Error; err != nil {
		t.Fatalf("CREATE SCHEMA %s error = %v; want nil", guardSchema, err)
	}
	if err := db.Exec("CREATE VIEW " + quotedGuardSchema + "." + quotedGuardView + " AS " + selectSQL).Error; err != nil {
		t.Fatalf("CREATE VIEW %s.%s error = %v; want nil", guardSchema, viewName, err)
	}

	return guardSchema, viewName
}

func assertEntryFormSourceRelations(t *testing.T, db *gorm.DB, formID int64, want []string) {
	t.Helper()

	var form model.EntryForm
	if err := db.First(&form, formID).Error; err != nil {
		t.Fatalf("First(entry_form=%d) error = %v; want nil", formID, err)
	}

	if !slices.Equal([]string(form.SourceRelations), want) {
		t.Fatalf("entry_form source_relations = %v; want %v", []string(form.SourceRelations), want)
	}
}

func assertModelRowCount(t *testing.T, db *gorm.DB, modelValue any, want int64, query string, args ...any) {
	t.Helper()

	var count int64
	if err := db.Model(modelValue).Where(query, args...).Count(&count).Error; err != nil {
		t.Fatalf("Count(%T) error = %v; want nil", modelValue, err)
	}

	if count != want {
		t.Fatalf("Count(%T) = %d; want %d", modelValue, count, want)
	}
}

func loadTableMaxID(t *testing.T, db *gorm.DB, tableName string) int64 {
	t.Helper()

	maxID, err := loadCurrentSchemaMaxID(db, tableName)
	if err != nil {
		t.Fatalf("load max id for %s: %v", tableName, err)
	}

	return maxID
}

func assertAllMigrationTablesEmpty(t *testing.T, db *gorm.DB) {
	t.Helper()

	currentSchema, err := loadCurrentSchema(db)
	if err != nil {
		t.Fatalf("load current schema: %v", err)
	}
	quotedCurrentSchema := quotePostgresIdentifier(currentSchema)

	for _, tableName := range postgresIntegrationExpectedTables {
		var result struct {
			RowCount int64 `gorm:"column:row_count"`
		}

		qualifiedTableName := quotedCurrentSchema + "." + quotePostgresIdentifier(tableName)
		if err := db.Raw("SELECT COUNT(*) AS row_count FROM " + qualifiedTableName).Scan(&result).Error; err != nil {
			t.Fatalf("count rows for %s: %v", tableName, err)
		}
		if result.RowCount != 0 {
			t.Fatalf("row count for %s = %d; want 0 after DropTables rebuild", tableName, result.RowCount)
		}
	}
}
