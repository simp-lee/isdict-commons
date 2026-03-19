package migration

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const postgresIntegrationDSNEnv = "ISDICT_TEST_POSTGRES_DSN"

var postgresIntegrationTestDBNamePattern = regexp.MustCompile(`(^|[-_])(test|tests|testing)([-_]|$)`)

type postgresIntegrationTarget struct {
	DatabaseName  string
	CurrentSchema string
	SearchPath    string
}

// AC-B8: real PostgreSQL migration/verification must be opt-in and succeed when only required indexes are verified.
func TestMigrate_PostgresIntegration_SkipIndexesAndExtensions(t *testing.T) {
	db := newPostgresIntegrationDB(t)
	migrator := NewMigrator(db)
	opts := &MigrateOptions{
		SkipExtensions: true,
		SkipIndexes:    true,
	}

	if err := migrator.Migrate(opts); err != nil {
		t.Fatalf("Migrate() error = %v; want nil", err)
	}

	status, err := migrator.VerifyMigration(opts, nil)
	if err != nil {
		t.Fatalf("VerifyMigration() error = %v; want nil", err)
	}

	if !status.IsComplete() {
		t.Fatalf("status.IsComplete() = false; status = %+v", status)
	}

	if len(status.Issues) != 0 {
		t.Fatalf("status.Issues = %v; want none", status.Issues)
	}

	if !sameStrings(status.Tables, []string{"words", "word_variants", "pronunciations", "senses", "examples"}) {
		t.Fatalf("status.Tables = %v; want all migrated tables", status.Tables)
	}

	if !sameStrings(status.Indexes, requiredIndexesList()) {
		t.Fatalf("status.Indexes = %v; want %v", status.Indexes, requiredIndexesList())
	}

	if len(status.MissingIndexes) != 0 {
		t.Fatalf("status.MissingIndexes = %v; want none", status.MissingIndexes)
	}

	if !sameStrings(status.SkippedIndexes, optionalPerformanceIndexesList()) {
		t.Fatalf("status.SkippedIndexes = %v; want %v", status.SkippedIndexes, optionalPerformanceIndexesList())
	}
}

// AC-B8: real PostgreSQL must enforce the partial unique index on pronunciations(word_id, accent) WHERE is_primary.
func TestMigrate_PostgresIntegration_EnforcesPronunciationPrimaryUniqueIndex(t *testing.T) {
	db := newMigratedPostgresIntegrationDB(t)
	wordID := insertPostgresIntegrationWord(t, db, "bug8-pronunciation-primary")

	firstPrimary := model.Pronunciation{
		WordID:    wordID,
		Accent:    1,
		IPA:       "w\u025c\u02d0d",
		IsPrimary: true,
	}
	if err := db.Create(&firstPrimary).Error; err != nil {
		t.Fatalf("Create(firstPrimary) error = %v; want nil", err)
	}

	nonPrimary := model.Pronunciation{
		WordID:    wordID,
		Accent:    1,
		IPA:       "w\u025d\u02d0d",
		IsPrimary: false,
	}
	if err := db.Create(&nonPrimary).Error; err != nil {
		t.Fatalf("Create(nonPrimary) error = %v; want nil", err)
	}

	duplicatePrimary := model.Pronunciation{
		WordID:    wordID,
		Accent:    1,
		IPA:       "w\u0259d",
		IsPrimary: true,
	}
	err := db.Create(&duplicatePrimary).Error
	if err == nil {
		t.Fatal("Create(duplicatePrimary) returned nil error; want unique index violation")
	}

	if !strings.Contains(err.Error(), "idx_pronunciation_primary_unique") {
		t.Fatalf("Create(duplicatePrimary) error = %v; want idx_pronunciation_primary_unique", err)
	}
}

// AC-B8: real PostgreSQL must enforce idx_word_variant_unique with COALESCE(form_type, 0) so duplicate NULL form_type rows collide.
func TestMigrate_PostgresIntegration_EnforcesWordVariantUniqueIndexNullHandling(t *testing.T) {
	db := newMigratedPostgresIntegrationDB(t)
	wordID := insertPostgresIntegrationWord(t, db, "bug8-word-variant-null")

	firstAlias := model.WordVariant{
		WordID:             wordID,
		VariantText:        "bug8-word-variant-null-alias",
		HeadwordNormalized: "bug8wordvariantnullalias",
		Kind:               model.VariantAlias,
	}
	if err := db.Create(&firstAlias).Error; err != nil {
		t.Fatalf("Create(firstAlias) error = %v; want nil", err)
	}

	duplicateAlias := model.WordVariant{
		WordID:             wordID,
		VariantText:        firstAlias.VariantText,
		HeadwordNormalized: firstAlias.HeadwordNormalized,
		Kind:               model.VariantAlias,
	}
	err := db.Create(&duplicateAlias).Error
	if err == nil {
		t.Fatal("Create(duplicateAlias) returned nil error; want COALESCE-backed unique index violation")
	}

	if !strings.Contains(err.Error(), "idx_word_variant_unique") {
		t.Fatalf("Create(duplicateAlias) error = %v; want idx_word_variant_unique", err)
	}
}

func TestVerifyMigration_SkipIndexesStillVerifiesRequiredUniqueIndexes(t *testing.T) {
	t.Parallel()

	status := &MigrationStatus{Issues: []string{}}
	skipSet := buildSkippedIndexSet(&MigrateOptions{SkipIndexes: true}, nil)

	collectExpectedIndexStatus(status, skipSet, map[string]string{})

	wantMissing := []string{
		"idx_pronunciation_primary_unique",
		"idx_word_variant_unique",
	}
	if !reflect.DeepEqual(status.MissingIndexes, wantMissing) {
		t.Fatalf("MissingIndexes = %v; want %v", status.MissingIndexes, wantMissing)
	}

	wantSkipped := optionalPerformanceIndexesList()
	if !sameStrings(status.SkippedIndexes, wantSkipped) {
		t.Fatalf("SkippedIndexes = %v; want %v", status.SkippedIndexes, wantSkipped)
	}
}

func TestVerifyMigration_SkipIndexesStillChecksRequiredUniqueIndexIntegrity(t *testing.T) {
	t.Parallel()

	status := &MigrationStatus{Issues: []string{}}
	skipSet := buildSkippedIndexSet(&MigrateOptions{SkipIndexes: true}, nil)
	indexMap := map[string]string{
		"idx_pronunciation_primary_unique": "CREATE UNIQUE INDEX idx_pronunciation_primary_unique ON pronunciations(word_id, accent) WHERE is_primary",
		"idx_word_variant_unique":          "CREATE UNIQUE INDEX idx_word_variant_unique ON word_variants(word_id, variant_text, kind, form_type)",
	}

	collectExpectedIndexStatus(status, skipSet, indexMap)

	wantIndexes := requiredIndexesList()
	if !sameStrings(status.Indexes, wantIndexes) {
		t.Fatalf("Indexes = %v; want %v", status.Indexes, wantIndexes)
	}

	if len(status.MissingIndexes) != 0 {
		t.Fatalf("MissingIndexes = %v; want none", status.MissingIndexes)
	}

	if len(status.Issues) != 1 || !strings.Contains(status.Issues[0], "idx_word_variant_unique") {
		t.Fatalf("Issues = %v; want word variant unique index integrity issue", status.Issues)
	}

	wantSkipped := optionalPerformanceIndexesList()
	if !sameStrings(status.SkippedIndexes, wantSkipped) {
		t.Fatalf("SkippedIndexes = %v; want %v", status.SkippedIndexes, wantSkipped)
	}
}

// AC-B5: VerifyMigration 必须验证 idx_word_variant_unique 是 word_variants 上的目标唯一索引，且列顺序与 COALESCE(form_type, 0) 表达式完全匹配。
func TestVerifyMigration_ValidatesWordVariantUniqueIndexDefinition(t *testing.T) {
	t.Parallel()

	skipSet := buildSkippedIndexSet(&MigrateOptions{SkipIndexes: true}, nil)
	pronunciationPrimaryIndex := "CREATE UNIQUE INDEX idx_pronunciation_primary_unique ON pronunciations(word_id, accent) WHERE is_primary"

	tests := []struct {
		name           string
		wordVariantDef string
		wantIssue      bool
	}{
		{
			name:           "same_name_non_unique_with_coalesce_reports_issue",
			wordVariantDef: "CREATE INDEX idx_word_variant_unique ON word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))",
			wantIssue:      true,
		},
		{
			name:           "same_name_on_wrong_table_with_coalesce_reports_issue",
			wordVariantDef: "CREATE UNIQUE INDEX idx_word_variant_unique ON words(word_id, variant_text, kind, COALESCE(form_type, 0))",
			wantIssue:      true,
		},
		{
			name:           "same_name_with_reordered_columns_reports_issue",
			wordVariantDef: "CREATE UNIQUE INDEX idx_word_variant_unique ON word_variants(word_id, kind, variant_text, COALESCE(form_type, 0))",
			wantIssue:      true,
		},
		{
			name:           "same_name_with_different_coalesce_expression_reports_issue",
			wordVariantDef: "CREATE UNIQUE INDEX idx_word_variant_unique ON word_variants(word_id, variant_text, kind, COALESCE(form_type, 1))",
			wantIssue:      true,
		},
		{
			name:           "expected_definition_passes",
			wordVariantDef: "CREATE UNIQUE INDEX idx_word_variant_unique ON public.word_variants USING btree (word_id, variant_text, kind, COALESCE(form_type, 0))",
			wantIssue:      false,
		},
		{
			name:           "postgres_deparser_equivalent_definition_passes",
			wordVariantDef: "CREATE UNIQUE INDEX idx_word_variant_unique ON public.word_variants USING btree (word_id, variant_text, kind, (COALESCE((form_type)::smallint, (0)::smallint)))",
			wantIssue:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := &MigrationStatus{Issues: []string{}}
			indexMap := map[string]string{
				"idx_pronunciation_primary_unique": pronunciationPrimaryIndex,
				"idx_word_variant_unique":          tt.wordVariantDef,
			}

			collectExpectedIndexStatus(status, skipSet, indexMap)

			if !sameStrings(status.Indexes, requiredIndexesList()) {
				t.Fatalf("Indexes = %v; want %v", status.Indexes, requiredIndexesList())
			}

			if len(status.MissingIndexes) != 0 {
				t.Fatalf("MissingIndexes = %v; want none", status.MissingIndexes)
			}

			hasWordVariantIssue := hasIssueContaining(status.Issues, "idx_word_variant_unique")
			if hasWordVariantIssue != tt.wantIssue {
				t.Fatalf("has word variant unique index issue = %v, Issues = %v; want %v", hasWordVariantIssue, status.Issues, tt.wantIssue)
			}

			if tt.wantIssue && !hasIssueContaining(status.Issues, "word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))") {
				t.Fatalf("Issues = %v; want expected word variant unique index definition message", status.Issues)
			}

			if !tt.wantIssue && len(status.Issues) != 0 {
				t.Fatalf("Issues = %v; want none", status.Issues)
			}
		})
	}
}

// AC-B3: VerifyMigration 必须验证 idx_pronunciation_primary_unique 的目标定义是带 WHERE is_primary 的 partial unique index，而不是仅校验索引名存在。
func TestVerifyMigration_ValidatesPronunciationPrimaryIndexDefinition(t *testing.T) {
	t.Parallel()

	skipSet := buildSkippedIndexSet(&MigrateOptions{SkipIndexes: true}, nil)
	wordVariantUniqueIndex := "CREATE UNIQUE INDEX idx_word_variant_unique ON word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))"

	tests := []struct {
		name                  string
		pronunciationIndexDef string
		wantIssue             bool
	}{
		{
			name:                  "same_name_without_partial_predicate_reports_issue",
			pronunciationIndexDef: "CREATE UNIQUE INDEX idx_pronunciation_primary_unique ON pronunciations(word_id, accent)",
			wantIssue:             true,
		},
		{
			name:                  "partial_unique_with_is_primary_passes",
			pronunciationIndexDef: "CREATE UNIQUE INDEX idx_pronunciation_primary_unique ON pronunciations(word_id, accent) WHERE is_primary",
			wantIssue:             false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := &MigrationStatus{Issues: []string{}}
			indexMap := map[string]string{
				"idx_pronunciation_primary_unique": tt.pronunciationIndexDef,
				"idx_word_variant_unique":          wordVariantUniqueIndex,
			}

			collectExpectedIndexStatus(status, skipSet, indexMap)

			if !sameStrings(status.Indexes, requiredIndexesList()) {
				t.Fatalf("Indexes = %v; want %v", status.Indexes, requiredIndexesList())
			}

			if len(status.MissingIndexes) != 0 {
				t.Fatalf("MissingIndexes = %v; want none", status.MissingIndexes)
			}

			hasPronunciationIndexIssue := hasIssueContaining(status.Issues, "idx_pronunciation_primary_unique")
			if hasPronunciationIndexIssue != tt.wantIssue {
				t.Fatalf("has pronunciation index issue = %v, Issues = %v; want %v", hasPronunciationIndexIssue, status.Issues, tt.wantIssue)
			}

			if !tt.wantIssue && len(status.Issues) != 0 {
				t.Fatalf("Issues = %v; want none", status.Issues)
			}
		})
	}
}

func TestMigrationStatus_RequiredIndexIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		issues []string
		want   []string
	}{
		{
			name: "returns_required_index_integrity_issues_only",
			issues: []string{
				"idx_word_variant_unique is not the expected unique index on word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))",
				"Duplicate B-tree indexes on words(headword): [idx_words_headword idx_words_headword_2]",
				"idx_pronunciation_primary_unique is not the expected partial unique index on pronunciations(word_id, accent) WHERE is_primary",
			},
			want: []string{
				"idx_word_variant_unique is not the expected unique index on word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))",
				"idx_pronunciation_primary_unique is not the expected partial unique index on pronunciations(word_id, accent) WHERE is_primary",
			},
		},
		{
			name: "ignores_non_required_issues",
			issues: []string{
				"Duplicate B-tree indexes on words(headword): [idx_words_headword idx_words_headword_2]",
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := &MigrationStatus{Issues: tt.issues}

			if got := status.RequiredIndexIssues(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("RequiredIndexIssues() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyMigrationWithLogging_ReturnsErrorForRequiredIndexIssues(t *testing.T) {
	t.Parallel()

	migrator := &Migrator{}
	status := &MigrationStatus{
		Tables:     []string{"words", "word_variants", "pronunciations", "senses", "examples"},
		Indexes:    requiredIndexesList(),
		Extensions: []string{"pg_trgm"},
		Issues: []string{
			"idx_word_variant_unique is not the expected unique index on word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))",
		},
	}

	err := migrator.handleMigrationVerificationStatus(status)
	if err == nil {
		t.Fatal("handleMigrationVerificationStatus() returned nil error, want required index integrity failure")
	}

	if !strings.Contains(err.Error(), "required index integrity issues") {
		t.Fatalf("error = %v; want required index integrity issues context", err)
	}

	if !strings.Contains(err.Error(), "idx_word_variant_unique") {
		t.Fatalf("error = %v; want failing required index name", err)
	}
}

// AC-B4: when verification reports only non-required warnings, migration verification handling does not fail.
func TestVerifyMigrationWithLogging_DoesNotFailForNonRequiredWarnings(t *testing.T) {
	t.Parallel()

	migrator := &Migrator{}
	status := &MigrationStatus{
		Tables:     []string{"words", "word_variants", "pronunciations", "senses", "examples"},
		Indexes:    requiredIndexesList(),
		Extensions: []string{"pg_trgm"},
		Issues: []string{
			"Duplicate B-tree indexes on words(headword): [idx_words_headword idx_words_headword_2]",
		},
	}

	if err := migrator.handleMigrationVerificationStatus(status); err != nil {
		t.Fatalf("handleMigrationVerificationStatus() error = %v; want nil for non-required warning", err)
	}
}

func TestLoadExistingIndexMap_QueryIncludesSchemaQualifiedCatalogJoin(t *testing.T) {
	t.Parallel()

	if !strings.Contains(loadExistingIndexMapQuery, "JOIN pg_namespace n ON n.nspname = i.schemaname") {
		t.Fatalf("loadExistingIndexMapQuery = %q; want pg_namespace join on schema name", loadExistingIndexMapQuery)
	}

	if !strings.Contains(loadExistingIndexMapQuery, "JOIN pg_class c ON c.relname = i.indexname AND c.relnamespace = n.oid") {
		t.Fatalf("loadExistingIndexMapQuery = %q; want pg_class join qualified by relnamespace", loadExistingIndexMapQuery)
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

	if !strings.Contains(err.Error(), "test database") {
		t.Fatalf("validatePostgresIntegrationTarget() error = %v; want test database context", err)
	}

	if !strings.Contains(err.Error(), "isdict") {
		t.Fatalf("validatePostgresIntegrationTarget() error = %v; want database name context", err)
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

func TestResetPostgresIntegrationSchemaStatements_TargetPublicSchemaOnly(t *testing.T) {
	t.Parallel()

	statements := postgresIntegrationResetStatements()
	if len(statements) == 0 {
		t.Fatal("postgresIntegrationResetStatements() returned no statements")
	}

	for _, stmt := range statements {
		if !strings.Contains(stmt, "DROP TABLE IF EXISTS public.") {
			t.Fatalf("reset statement = %q; want public schema-qualified DROP TABLE", stmt)
		}
	}
}

func sameStrings(got, want []string) bool {
	gotCopy := append([]string(nil), got...)
	wantCopy := append([]string(nil), want...)
	sort.Strings(gotCopy)
	sort.Strings(wantCopy)
	return reflect.DeepEqual(gotCopy, wantCopy)
}

func hasIssueContaining(issues []string, substring string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, substring) {
			return true
		}
	}

	return false
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

func newMigratedPostgresIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := newPostgresIntegrationDB(t)
	migrator := NewMigrator(db)
	if err := migrator.Migrate(&MigrateOptions{
		SkipExtensions: true,
		SkipIndexes:    true,
	}); err != nil {
		t.Fatalf("Migrate() error = %v; want nil", err)
	}

	return db
}

func resetPostgresIntegrationSchema(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := resetPostgresIntegrationSchemaNoFail(db); err != nil {
		t.Fatalf("reset PostgreSQL integration schema: %v", err)
	}
}

func resetPostgresIntegrationSchemaNoFail(db *gorm.DB) error {
	target, err := inspectPostgresIntegrationTarget(db)
	if err != nil {
		return fmt.Errorf("inspect PostgreSQL integration target: %w", err)
	}

	if err := validatePostgresIntegrationTarget(target); err != nil {
		return err
	}

	for _, stmt := range postgresIntegrationResetStatements() {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
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

func validatePostgresIntegrationTarget(target postgresIntegrationTarget) error {
	databaseName := strings.TrimSpace(target.DatabaseName)
	if !isSafePostgresIntegrationDatabaseName(databaseName) {
		return fmt.Errorf("refusing PostgreSQL integration cleanup on non-test database %q; database name must clearly identify a test database", databaseName)
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

func isSafePostgresIntegrationDatabaseName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	return normalized != "" && postgresIntegrationTestDBNamePattern.MatchString(normalized)
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

func postgresIntegrationResetStatements() []string {
	return []string{
		`DROP TABLE IF EXISTS public.examples CASCADE`,
		`DROP TABLE IF EXISTS public.senses CASCADE`,
		`DROP TABLE IF EXISTS public.pronunciations CASCADE`,
		`DROP TABLE IF EXISTS public.word_variants CASCADE`,
		`DROP TABLE IF EXISTS public.words CASCADE`,
	}
}

func insertPostgresIntegrationWord(t *testing.T, db *gorm.DB, headword string) uint {
	t.Helper()

	word := model.Word{
		Headword:           headword,
		HeadwordNormalized: headword,
	}
	if err := db.Create(&word).Error; err != nil {
		t.Fatalf("Create(word) error = %v; want nil", err)
	}

	return word.ID
}
