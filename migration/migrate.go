package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/simp-lee/isdict-commons/model"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

//go:embed sql/*.sql
var embeddedSQLFiles embed.FS

const embeddedSQLGlobPattern = "sql/*.sql"

type migrationTarget struct {
	TableName string
	Model     any
}

type indexTarget struct {
	TableName string
	Model     any
	IndexName string
}

type sqlIndexDefinitionTarget struct {
	TableName string
	IndexName string
	Unique    bool
	Method    string
	Columns   []string
	Predicate string
}

type identityColumnState struct {
	IsIdentity         string        `gorm:"column:is_identity"`
	IdentityGeneration string        `gorm:"column:identity_generation"`
	SequenceSchema     string        `gorm:"column:sequence_schema"`
	SequenceName       string        `gorm:"column:sequence_name"`
	SequenceLastValue  sql.NullInt64 `gorm:"column:sequence_last_value"`
}

var migrationTargets = []migrationTarget{
	{TableName: "import_runs", Model: &model.ImportRun{}},
	{TableName: "entries", Model: &model.Entry{}},
	{TableName: "senses", Model: &model.Sense{}},
	{TableName: "sense_glosses_en", Model: &model.SenseGlossEN{}},
	{TableName: "sense_glosses_zh", Model: &model.SenseGlossZH{}},
	{TableName: "sense_labels", Model: &model.SenseLabel{}},
	{TableName: "sense_examples", Model: &model.SenseExample{}},
	{TableName: "pronunciation_ipas", Model: &model.PronunciationIPA{}},
	{TableName: "pronunciation_audios", Model: &model.PronunciationAudio{}},
	{TableName: "entry_forms", Model: &model.EntryForm{}},
	{TableName: "lexical_relations", Model: &model.LexicalRelation{}},
	{TableName: "entry_summaries_zh", Model: &model.EntrySummaryZH{}},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}},
	{TableName: "sense_learning_signals", Model: &model.SenseLearningSignal{}},
	{TableName: "entry_etymologies", Model: &model.EntryEtymology{}},
}

var identityManagedTables = []string{
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
}

var dropTargets = []any{
	&model.EntryEtymology{},
	&model.SenseLearningSignal{},
	&model.EntryLearningSignal{},
	&model.EntrySummaryZH{},
	&model.LexicalRelation{},
	&model.EntryForm{},
	&model.PronunciationAudio{},
	&model.PronunciationIPA{},
	&model.SenseExample{},
	&model.SenseLabel{},
	&model.SenseGlossZH{},
	&model.SenseGlossEN{},
	&model.Sense{},
	&model.Entry{},
	&model.ImportRun{},
}

var expectedIndexes = []indexTarget{
	{TableName: "import_runs", Model: &model.ImportRun{}, IndexName: "idx_import_runs_source_name_started_at"},
	{TableName: "import_runs", Model: &model.ImportRun{}, IndexName: "idx_import_runs_status"},
	{TableName: "entries", Model: &model.Entry{}, IndexName: "idx_entries_headword_pos_etymology_index"},
	{TableName: "entries", Model: &model.Entry{}, IndexName: "idx_entries_headword"},
	{TableName: "entries", Model: &model.Entry{}, IndexName: "idx_entries_normalized_headword"},
	{TableName: "entries", Model: &model.Entry{}, IndexName: "idx_entries_pos"},
	{TableName: "entries", Model: &model.Entry{}, IndexName: "idx_entries_source_run_id"},
	{TableName: "entries", Model: &model.Entry{}, IndexName: "idx_entries_normalized_headword_trgm"},
	{TableName: "senses", Model: &model.Sense{}, IndexName: "idx_senses_entry_id_sense_order"},
	{TableName: "sense_glosses_en", Model: &model.SenseGlossEN{}, IndexName: "idx_sense_glosses_en_sense_id_gloss_order"},
	{TableName: "sense_glosses_zh", Model: &model.SenseGlossZH{}, IndexName: "idx_sense_glosses_zh_sense_id_source_gloss_order"},
	{TableName: "sense_glosses_zh", Model: &model.SenseGlossZH{}, IndexName: "idx_sense_glosses_zh_sense_id_gloss_order"},
	{TableName: "sense_glosses_zh", Model: &model.SenseGlossZH{}, IndexName: "idx_sense_glosses_zh_source_run_id"},
	{TableName: "sense_glosses_zh", Model: &model.SenseGlossZH{}, IndexName: "idx_sense_glosses_zh_sense_id_source_primary"},
	{TableName: "sense_labels", Model: &model.SenseLabel{}, IndexName: "idx_sense_labels_sense_id_label_type_label_code"},
	{TableName: "sense_labels", Model: &model.SenseLabel{}, IndexName: "idx_sense_labels_sense_id_label_type_label_order"},
	{TableName: "sense_labels", Model: &model.SenseLabel{}, IndexName: "idx_sense_labels_label_type_label_code"},
	{TableName: "sense_examples", Model: &model.SenseExample{}, IndexName: "idx_sense_examples_sense_id_source_example_order"},
	{TableName: "sense_examples", Model: &model.SenseExample{}, IndexName: "idx_sense_examples_sense_id_example_order"},
	{TableName: "pronunciation_ipas", Model: &model.PronunciationIPA{}, IndexName: "idx_pronunciation_ipas_entry_id_accent_code_ipa"},
	{TableName: "pronunciation_ipas", Model: &model.PronunciationIPA{}, IndexName: "idx_pronunciation_ipas_entry_id_accent_code_display_order"},
	{TableName: "pronunciation_ipas", Model: &model.PronunciationIPA{}, IndexName: "idx_pronunciation_ipas_entry_id_accent_code_primary"},
	{TableName: "pronunciation_audios", Model: &model.PronunciationAudio{}, IndexName: "idx_pronunciation_audios_entry_id_accent_code_audio_filename"},
	{TableName: "pronunciation_audios", Model: &model.PronunciationAudio{}, IndexName: "idx_pronunciation_audios_entry_id_accent_code_display_order"},
	{TableName: "pronunciation_audios", Model: &model.PronunciationAudio{}, IndexName: "idx_pronunciation_audios_audio_filename"},
	{TableName: "pronunciation_audios", Model: &model.PronunciationAudio{}, IndexName: "idx_pronunciation_audios_entry_id_accent_code_primary"},
	{TableName: "entry_forms", Model: &model.EntryForm{}, IndexName: "idx_entry_forms_entry_id_relation_kind"},
	{TableName: "entry_forms", Model: &model.EntryForm{}, IndexName: "idx_entry_forms_normalized_form"},
	{TableName: "entry_forms", Model: &model.EntryForm{}, IndexName: "idx_entry_forms_entry_id_relation_kind_form_text_form_type"},
	{TableName: "lexical_relations", Model: &model.LexicalRelation{}, IndexName: "idx_lexical_relations_entry_id_relation_type"},
	{TableName: "lexical_relations", Model: &model.LexicalRelation{}, IndexName: "idx_lexical_relations_sense_id_relation_type"},
	{TableName: "lexical_relations", Model: &model.LexicalRelation{}, IndexName: "idx_lexical_relations_entry_id_sense_id_rel_type_target_norm"},
	{TableName: "entry_summaries_zh", Model: &model.EntrySummaryZH{}, IndexName: "idx_entry_summaries_zh_entry_id_source"},
	{TableName: "entry_summaries_zh", Model: &model.EntrySummaryZH{}, IndexName: "idx_entry_summaries_zh_entry_id"},
	{TableName: "entry_summaries_zh", Model: &model.EntrySummaryZH{}, IndexName: "idx_entry_summaries_zh_source_updated_at"},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}, IndexName: "idx_entry_learning_signals_cefr_level"},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}, IndexName: "idx_entry_learning_signals_oxford_level"},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}, IndexName: "idx_entry_learning_signals_cet_level"},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}, IndexName: "idx_entry_learning_signals_school_level"},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}, IndexName: "idx_entry_learning_signals_frequency_rank"},
	{TableName: "entry_learning_signals", Model: &model.EntryLearningSignal{}, IndexName: "idx_entry_learning_signals_collins_stars"},
	{TableName: "sense_learning_signals", Model: &model.SenseLearningSignal{}, IndexName: "idx_sense_learning_signals_cefr_level"},
	{TableName: "sense_learning_signals", Model: &model.SenseLearningSignal{}, IndexName: "idx_sense_learning_signals_oxford_level"},
	{TableName: "entry_etymologies", Model: &model.EntryEtymology{}, IndexName: "idx_entry_etymologies_source_updated_at"},
}

var postgresTypeCastPattern = regexp.MustCompile(`::[a-z0-9_\.\[\]]+`)

var schemaQualifierPattern = regexp.MustCompile(`(?:^|[\s,(])(?:[a-z0-9_\-$]+\.)+`)

var sqlManagedIndexDefinitions = []sqlIndexDefinitionTarget{
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

const analyzeTablesSQL = `ANALYZE import_runs, entries, senses, sense_glosses_en, sense_glosses_zh, sense_labels, sense_examples, pronunciation_ipas, pronunciation_audios, entry_forms, lexical_relations, entry_summaries_zh, entry_learning_signals, sense_learning_signals, entry_etymologies`

// MigrateOptions controls schema reset and verbose progress logging for RunMigration.
// ANALYZE is intentionally best-effort and always attempted; failures are warned about,
// but they do not change the API surface or fail the migration.
type MigrateOptions struct {
	DropTables bool
	Verbose    bool
}

// RunMigration applies the full schema migration for the current 15-table model set.
// ANALYZE is best-effort: the migration always attempts it and logs a warning if it fails.
func RunMigration(db *gorm.DB, opts MigrateOptions) error {
	if db == nil {
		return fmt.Errorf("nil database handle")
	}

	logVerbose(opts, "starting schema migration")

	if opts.DropTables {
		if err := dropTables(db, opts); err != nil {
			return fmt.Errorf("drop tables: %w", err)
		}
	}

	if err := autoMigrateTables(db, opts); err != nil {
		return fmt.Errorf("auto migrate tables: %w", err)
	}

	if err := executeEmbeddedSQL(db, opts); err != nil {
		return fmt.Errorf("execute embedded SQL: %w", err)
	}

	if err := analyzeTables(db, opts); err != nil {
		logAnalyzeWarning(err)
	}

	if err := verifyMigration(db, opts); err != nil {
		return fmt.Errorf("verify migration: %w", err)
	}

	logVerbose(opts, "schema migration completed")
	return nil
}

func dropTables(db *gorm.DB, opts MigrateOptions) error {
	logVerbose(opts, "dropping existing tables")
	if err := db.Migrator().DropTable(dropTargets...); err != nil {
		return err
	}

	return nil
}

func autoMigrateTables(db *gorm.DB, opts MigrateOptions) error {
	logVerbose(opts, "running AutoMigrate for all models")
	if err := db.AutoMigrate(modelsForAutoMigrate()...); err != nil {
		return err
	}

	return nil
}

func listEmbeddedSQLFiles() ([]string, error) {
	files, err := fs.Glob(embeddedSQLFiles, embeddedSQLGlobPattern)
	if err != nil {
		return nil, fmt.Errorf("list embedded SQL files: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no embedded SQL files found")
	}

	sort.Strings(files)
	return files, nil
}

func executeEmbeddedSQL(db *gorm.DB, opts MigrateOptions) error {
	files, err := listEmbeddedSQLFiles()
	if err != nil {
		return err
	}

	for _, fileName := range files {
		sqlBytes, err := embeddedSQLFiles.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("read %s: %w", fileName, err)
		}

		sqlText := strings.TrimSpace(string(sqlBytes))
		if sqlText == "" {
			continue
		}

		logVerbose(opts, "executing %s", fileName)
		if err := db.Exec(sqlText).Error; err != nil {
			return fmt.Errorf("exec %s: %w", fileName, err)
		}
	}

	return nil
}

func analyzeTables(db *gorm.DB, opts MigrateOptions) error {
	logVerbose(opts, "updating table statistics")
	return db.Exec(analyzeTablesSQL).Error
}

func logAnalyzeWarning(err error) {
	if err == nil {
		return
	}

	log.Printf("migration: analyze tables warning (best-effort, migration continues): %v", err)
}

func verifyMigration(db *gorm.DB, opts MigrateOptions) error {
	logVerbose(opts, "verifying tables, extension, indexes, and identity sequences")

	if err := verifyTables(db); err != nil {
		return err
	}

	if err := verifyExtension(db, "pg_trgm"); err != nil {
		return err
	}

	if err := verifyIndexes(db); err != nil {
		return err
	}

	if err := verifyIdentityColumns(db); err != nil {
		return err
	}

	return nil
}

func verifyTables(db *gorm.DB) error {
	missingTables := make([]string, 0)
	for _, target := range migrationTargets {
		if db.Migrator().HasTable(target.TableName) {
			continue
		}
		missingTables = append(missingTables, target.TableName)
	}

	if len(missingTables) > 0 {
		return fmt.Errorf("missing tables: %s", strings.Join(missingTables, ", "))
	}

	return nil
}

func verifyExtension(db *gorm.DB, extensionName string) error {
	var result struct {
		Count int64 `gorm:"column:count"`
	}

	if err := db.Raw(`SELECT COUNT(*) AS count FROM pg_extension WHERE extname = ?`, extensionName).Scan(&result).Error; err != nil {
		return fmt.Errorf("check extension %s: %w", extensionName, err)
	}

	if result.Count == 0 {
		return fmt.Errorf("missing extension: %s", extensionName)
	}

	return nil
}

func verifyIndexes(db *gorm.DB) error {
	missingIndexes := make([]string, 0)
	missingIndexNames := make(map[string]struct{})
	for _, target := range expectedIndexes {
		if db.Migrator().HasIndex(target.Model, target.IndexName) {
			continue
		}
		missingIndexNames[target.IndexName] = struct{}{}
		missingIndexes = append(missingIndexes, fmt.Sprintf("%s(%s)", target.IndexName, target.TableName))
	}

	invalidDefinitions, err := verifySQLManagedIndexDefinitions(db, missingIndexNames)
	if err != nil {
		return err
	}
	gormManagedInvalidDefinitions, err := verifyGORMManagedIndexDefinitions(db, missingIndexNames)
	if err != nil {
		return err
	}
	invalidDefinitions = append(invalidDefinitions, gormManagedInvalidDefinitions...)

	issues := make([]string, 0, 2)
	if len(missingIndexes) > 0 {
		issues = append(issues, fmt.Sprintf("missing indexes: %s", strings.Join(missingIndexes, ", ")))
	}
	if len(invalidDefinitions) > 0 {
		issues = append(issues, fmt.Sprintf("invalid index definitions: %s", strings.Join(invalidDefinitions, "; ")))
	}

	if len(issues) > 0 {
		return fmt.Errorf("%s", strings.Join(issues, "; "))
	}

	return nil
}

func verifyIdentityColumns(db *gorm.DB) error {
	currentSchema, err := loadCurrentSchema(db)
	if err != nil {
		return err
	}

	issues := make([]string, 0)
	for _, tableName := range identityManagedTables {
		state, err := loadIdentityColumnState(db, tableName)
		if err != nil {
			return err
		}

		if issue := identityColumnStateIssue(tableName, currentSchema, state); issue != "" {
			issues = append(issues, issue)
			continue
		}

		maxID, err := loadCurrentSchemaMaxID(db, tableName)
		if err != nil {
			return err
		}
		if issue := identitySequencePositionIssue(tableName, state, maxID); issue != "" {
			issues = append(issues, issue)
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("invalid identity columns: %s", strings.Join(issues, "; "))
	}

	return nil
}

func identityColumnStateIssue(tableName, currentSchema string, state identityColumnState) string {
	if !strings.EqualFold(strings.TrimSpace(state.IsIdentity), "YES") {
		return fmt.Sprintf("%s.id is not an identity column in schema %s", tableName, currentSchema)
	}

	if !strings.EqualFold(strings.TrimSpace(state.IdentityGeneration), "ALWAYS") {
		return fmt.Sprintf("%s.id identity generation is %q in schema %s; want ALWAYS", tableName, strings.TrimSpace(state.IdentityGeneration), currentSchema)
	}

	if strings.TrimSpace(state.SequenceSchema) == "" || strings.TrimSpace(state.SequenceName) == "" {
		return fmt.Sprintf("%s.id identity sequence missing in schema %s", tableName, currentSchema)
	}

	if state.SequenceSchema != currentSchema {
		return fmt.Sprintf("%s.id identity sequence is in schema %s; want %s", tableName, state.SequenceSchema, currentSchema)
	}

	return ""
}

func identitySequencePositionIssue(tableName string, state identityColumnState, maxID int64) string {
	if !state.SequenceLastValue.Valid {
		if maxID == 0 {
			return ""
		}

		return fmt.Sprintf("%s.id identity sequence %s.%s has no last_value while max(id)=%d", tableName, state.SequenceSchema, state.SequenceName, maxID)
	}

	if state.SequenceLastValue.Int64 < maxID {
		return fmt.Sprintf("%s.id identity sequence %s.%s last_value=%d is behind max(id)=%d", tableName, state.SequenceSchema, state.SequenceName, state.SequenceLastValue.Int64, maxID)
	}

	return ""
}

func loadCurrentSchema(db *gorm.DB) (string, error) {
	var result struct {
		SchemaName string `gorm:"column:schema_name"`
	}

	if err := db.Raw(`SELECT current_schema() AS schema_name`).Scan(&result).Error; err != nil {
		return "", fmt.Errorf("load current schema: %w", err)
	}

	result.SchemaName = strings.TrimSpace(result.SchemaName)
	if result.SchemaName == "" {
		return "", fmt.Errorf("load current schema: current_schema() returned empty")
	}

	return result.SchemaName, nil
}

func loadIdentityColumnState(db *gorm.DB, tableName string) (identityColumnState, error) {
	var state identityColumnState

	if err := db.Raw(`
		SELECT
			c.is_identity AS is_identity,
			COALESCE(c.identity_generation, '') AS identity_generation,
			COALESCE(seq_ns.nspname, '') AS sequence_schema,
			COALESCE(seq_cls.relname, '') AS sequence_name,
			pgseq.last_value AS sequence_last_value
		FROM information_schema.columns c
		LEFT JOIN pg_class seq_cls
			ON seq_cls.oid = to_regclass(pg_get_serial_sequence(format('%I.%I', c.table_schema, c.table_name), c.column_name))
		LEFT JOIN pg_namespace seq_ns
			ON seq_ns.oid = seq_cls.relnamespace
		LEFT JOIN pg_sequences pgseq
			ON pgseq.schemaname = seq_ns.nspname
		   AND pgseq.sequencename = seq_cls.relname
		WHERE c.table_schema = current_schema()
		  AND c.table_name = ?
		  AND c.column_name = 'id'
	`, tableName).Scan(&state).Error; err != nil {
		return identityColumnState{}, fmt.Errorf("load identity state for %s.id: %w", tableName, err)
	}

	if strings.TrimSpace(state.IsIdentity) == "" {
		return identityColumnState{}, fmt.Errorf("load identity state for %s.id: column not found in current schema", tableName)
	}

	return state, nil
}

func loadCurrentSchemaMaxID(db *gorm.DB, tableName string) (int64, error) {
	var result struct {
		MaxID int64 `gorm:"column:max_id"`
	}

	if err := db.Table(tableName).Select("COALESCE(MAX(id), 0) AS max_id").Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("load max id for %s: %w", tableName, err)
	}

	return result.MaxID, nil
}

func verifySQLManagedIndexDefinitions(db *gorm.DB, missingIndexNames map[string]struct{}) ([]string, error) {
	return verifyManagedIndexDefinitions(db, missingIndexNames, sqlManagedIndexDefinitions, "SQL-managed")
}

func verifyGORMManagedIndexDefinitions(db *gorm.DB, missingIndexNames map[string]struct{}) ([]string, error) {
	targets, err := loadGORMManagedIndexDefinitions()
	if err != nil {
		return nil, err
	}

	return verifyManagedIndexDefinitions(db, missingIndexNames, targets, "GORM-managed")
}

func loadGORMManagedIndexDefinitions() ([]sqlIndexDefinitionTarget, error) {
	sqlManagedIndexNames := sqlManagedIndexNameSet()
	cache := &sync.Map{}
	namer := gormschema.NamingStrategy{}
	definitionsByName := make(map[string]sqlIndexDefinitionTarget, len(expectedIndexes))

	for _, target := range migrationTargets {
		if err := appendGORMManagedIndexDefinitions(cache, namer, definitionsByName, sqlManagedIndexNames, target); err != nil {
			return nil, err
		}
	}

	return finalizeGORMManagedIndexDefinitions(sqlManagedIndexNames, definitionsByName)
}

func sqlManagedIndexNameSet() map[string]struct{} {
	indexNames := make(map[string]struct{}, len(sqlManagedIndexDefinitions))
	for _, target := range sqlManagedIndexDefinitions {
		indexNames[target.IndexName] = struct{}{}
	}

	return indexNames
}

func appendGORMManagedIndexDefinitions(cache *sync.Map, namer gormschema.NamingStrategy, definitionsByName map[string]sqlIndexDefinitionTarget, sqlManagedIndexNames map[string]struct{}, target migrationTarget) error {
	parsedSchema, err := gormschema.Parse(target.Model, cache, namer)
	if err != nil {
		return fmt.Errorf("parse GORM schema for %s: %w", target.TableName, err)
	}
	if parsedSchema.Table != target.TableName {
		return fmt.Errorf("GORM schema table for %T = %s; want %s", target.Model, parsedSchema.Table, target.TableName)
	}

	for _, index := range parsedSchema.ParseIndexes() {
		if _, skip := sqlManagedIndexNames[index.Name]; skip {
			continue
		}
		if _, exists := definitionsByName[index.Name]; exists {
			return fmt.Errorf("duplicate GORM-managed index name %s", index.Name)
		}

		columns, err := gormIndexColumns(index.Fields)
		if err != nil {
			return fmt.Errorf("build GORM index columns for %s(%s): %w", index.Name, target.TableName, err)
		}

		method := strings.ToLower(strings.TrimSpace(index.Type))
		if method == "" {
			method = "btree"
		}

		definitionsByName[index.Name] = sqlIndexDefinitionTarget{
			TableName: target.TableName,
			IndexName: index.Name,
			Unique:    strings.EqualFold(index.Class, "UNIQUE"),
			Method:    method,
			Columns:   columns,
			Predicate: strings.TrimSpace(index.Where),
		}
	}

	return nil
}

func finalizeGORMManagedIndexDefinitions(sqlManagedIndexNames map[string]struct{}, definitionsByName map[string]sqlIndexDefinitionTarget) ([]sqlIndexDefinitionTarget, error) {
	definitions := make([]sqlIndexDefinitionTarget, 0, len(expectedIndexes))
	for _, target := range expectedIndexes {
		if _, skip := sqlManagedIndexNames[target.IndexName]; skip {
			continue
		}

		definition, ok := definitionsByName[target.IndexName]
		if !ok {
			return nil, fmt.Errorf("missing GORM-managed definition for %s(%s)", target.IndexName, target.TableName)
		}
		if definition.TableName != target.TableName {
			return nil, fmt.Errorf("GORM-managed definition for %s is on table %s; want %s", target.IndexName, definition.TableName, target.TableName)
		}

		definitions = append(definitions, definition)
		delete(definitionsByName, target.IndexName)
	}

	if len(definitionsByName) == 0 {
		return definitions, nil
	}

	extras := make([]string, 0, len(definitionsByName))
	for indexName, definition := range definitionsByName {
		extras = append(extras, fmt.Sprintf("%s(%s)", indexName, definition.TableName))
	}
	sort.Strings(extras)

	return nil, fmt.Errorf("unexpected GORM-managed indexes from model tags: %s", strings.Join(extras, ", "))
}

func gormIndexColumns(fields []gormschema.IndexOption) ([]string, error) {
	columns := make([]string, 0, len(fields))
	for _, field := range fields {
		column := strings.TrimSpace(field.Expression)
		if column == "" {
			if field.Field == nil || strings.TrimSpace(field.Field.DBName) == "" {
				return nil, fmt.Errorf("index field missing DB column name")
			}
			column = field.Field.DBName
		}

		if sortOrder := strings.ToLower(strings.TrimSpace(field.Sort)); sortOrder != "" {
			column += " " + sortOrder
		}

		columns = append(columns, column)
	}

	return columns, nil
}

func verifyManagedIndexDefinitions(db *gorm.DB, missingIndexNames map[string]struct{}, targets []sqlIndexDefinitionTarget, source string) ([]string, error) {
	invalidDefinitions := make([]string, 0)
	for _, target := range targets {
		if _, missing := missingIndexNames[target.IndexName]; missing {
			continue
		}

		indexDefinition, err := loadIndexDefinition(db, target.TableName, target.IndexName)
		if err != nil {
			return nil, err
		}

		if matchesIndexDefinition(indexDefinition, target) {
			continue
		}

		invalidDefinitions = append(invalidDefinitions, fmt.Sprintf("%s(%s) does not match expected %s definition", target.IndexName, target.TableName, source))
	}

	return invalidDefinitions, nil
}

func loadIndexDefinition(db *gorm.DB, tableName, indexName string) (string, error) {
	var result struct {
		IndexDefinition string `gorm:"column:indexdef"`
	}

	if err := db.Raw(`
		SELECT indexdef
		FROM pg_indexes
		WHERE schemaname = current_schema()
		  AND tablename = ?
		  AND indexname = ?
	`, tableName, indexName).Scan(&result).Error; err != nil {
		return "", fmt.Errorf("load index definition for %s(%s): %w", indexName, tableName, err)
	}

	return strings.TrimSpace(result.IndexDefinition), nil
}

func matchesIndexDefinition(indexDefinition string, target sqlIndexDefinitionTarget) bool {
	normalizedDefinition := normalizeIndexDefinition(indexDefinition)
	if normalizedDefinition == "" {
		return false
	}

	prefix := "create index "
	if target.Unique {
		prefix = "create unique index "
	}

	if !strings.HasPrefix(normalizedDefinition, prefix+target.IndexName+" on ") {
		return false
	}

	if !strings.Contains(normalizedDefinition, " on "+target.TableName+" ") {
		return false
	}

	if !strings.Contains(normalizedDefinition, " using "+target.Method+" ") {
		return false
	}

	columns, remainder, ok := splitIndexDefinition(normalizedDefinition)
	if !ok || len(columns) != len(target.Columns) {
		return false
	}

	for i, column := range columns {
		if normalizeIndexExpression(column) != normalizeIndexExpression(target.Columns[i]) {
			return false
		}
	}

	if target.Predicate == "" {
		return strings.TrimSpace(remainder) == ""
	}

	remainder = strings.TrimSpace(remainder)
	if !strings.HasPrefix(remainder, "where ") {
		return false
	}

	return normalizeIndexPredicate(strings.TrimPrefix(remainder, "where ")) == normalizeIndexPredicate(target.Predicate)
}

func splitIndexDefinition(indexDefinition string) ([]string, string, bool) {
	start := strings.Index(indexDefinition, "(")
	if start == -1 {
		return nil, "", false
	}

	depth := 0
	for i := start; i < len(indexDefinition); i++ {
		switch indexDefinition[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return splitTopLevelCSV(indexDefinition[start+1 : i]), strings.TrimSpace(indexDefinition[i+1:]), true
			}
		}
	}

	return nil, "", false
}

func splitTopLevelCSV(value string) []string {
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

func normalizeIndexDefinition(value string) string {
	normalized := strings.ToLower(value)
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")
	normalized = strings.ReplaceAll(normalized, `"`, "")
	normalized = postgresTypeCastPattern.ReplaceAllString(normalized, "")
	normalized = schemaQualifierPattern.ReplaceAllStringFunc(normalized, stripSchemaQualifierPrefix)
	return strings.Join(strings.Fields(normalized), " ")
}

func stripSchemaQualifierPrefix(value string) string {
	prefixStart := 0
	if len(value) > 0 {
		switch value[0] {
		case ' ', '(', ',':
			prefixStart = 1
		}
	}

	lastDot := strings.LastIndex(value, ".")
	if lastDot == -1 {
		return value
	}

	return value[:prefixStart] + value[lastDot+1:]
}

func normalizeIndexExpression(value string) string {
	normalized := stripWrappingParens(normalizeIndexDefinition(value))
	if strings.HasPrefix(normalized, "coalesce(") && strings.HasSuffix(normalized, ")") {
		args := splitTopLevelCSV(normalized[len("coalesce(") : len(normalized)-1])
		for i, arg := range args {
			args[i] = normalizeIndexExpression(arg)
		}
		return "coalesce(" + strings.Join(args, ", ") + ")"
	}

	return normalized
}

func normalizeIndexPredicate(value string) string {
	normalized := normalizeIndexExpression(value)
	if strings.HasSuffix(normalized, " = true") {
		return strings.TrimSpace(strings.TrimSuffix(normalized, " = true"))
	}

	return normalized
}

func stripWrappingParens(value string) string {
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

func modelsForAutoMigrate() []any {
	models := make([]any, 0, len(migrationTargets))
	for _, target := range migrationTargets {
		models = append(models, target.Model)
	}

	return models
}

func logVerbose(opts MigrateOptions, format string, args ...any) {
	if !opts.Verbose {
		return
	}

	log.Printf("migration: "+format, args...)
}
