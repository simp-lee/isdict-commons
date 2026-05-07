package migration

import (
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	postgresRealDataDSNEnv            = "ISDICT_REALDATA_POSTGRES_DSN"
	postgresRealDataAllowRefreshEnv   = "ISDICT_REALDATA_POSTGRES_ALLOW_REFRESH"
	postgresRealDataAllowRefreshValue = "refresh-derived-read-models"
)

var postgresRealDataCanonicalTables = []string{
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
	"entry_cefr_source_signals",
	"sense_learning_signals",
	"sense_cefr_source_signals",
	"entry_etymologies",
}

func TestRunMigration_PostgresRealData(t *testing.T) {
	db := newPostgresRealDataDB(t)

	sourceCountsBefore := loadPostgresRealDataTableCounts(t, db, postgresRealDataCanonicalTables)
	if sourceCountsBefore["entries"] == 0 {
		t.Fatal("real data entries count = 0; want existing imported data")
	}

	if err := RunMigration(db, MigrateOptions{}); err != nil {
		t.Fatalf("RunMigration(real data, DropTables=false) error = %v; want nil", err)
	}

	assertPostgresRealDataSourceCountsStable(t, db, sourceCountsBefore)
	assertPostgresRealDataMigrationTablesExist(t, db)
	assertPostgresRealDataExpectedIndexesExist(t, db)
	assertCurrentSchemaIndexAbsent(t, db, "entries", "idx_entries_normalized_headword_trgm")
	assertCurrentSchemaIndexAbsent(t, db, "entry_forms", "idx_entry_forms_normalized_form_trgm")
	assertPostgresRealDataReadModelCounts(t, db)
	assertPostgresRealDataSearchTermsMatchSource(t, db)
	assertPostgresRealDataFeaturedCandidatesMatchSource(t, db)
}

func newPostgresRealDataDB(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping PostgreSQL real-data test in short mode")
	}

	dsn := strings.TrimSpace(os.Getenv(postgresRealDataDSNEnv))
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL real-data tests", postgresRealDataDSNEnv)
	}

	if got := strings.TrimSpace(os.Getenv(postgresRealDataAllowRefreshEnv)); got != postgresRealDataAllowRefreshValue {
		t.Skipf(
			"set %s=%q to allow refreshing derived read models in the target database",
			postgresRealDataAllowRefreshEnv,
			postgresRealDataAllowRefreshValue,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL real-data database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		t.Fatalf("ping PostgreSQL real-data database: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func loadPostgresRealDataTableCounts(t *testing.T, db *gorm.DB, tableNames []string) map[string]int64 {
	t.Helper()

	counts := make(map[string]int64, len(tableNames))
	for _, tableName := range tableNames {
		counts[tableName] = loadTableRowCount(t, db, tableName)
	}

	return counts
}

func assertPostgresRealDataSourceCountsStable(t *testing.T, db *gorm.DB, before map[string]int64) {
	t.Helper()

	after := loadPostgresRealDataTableCounts(t, db, postgresRealDataCanonicalTables)
	for _, tableName := range postgresRealDataCanonicalTables {
		if got := after[tableName]; got != before[tableName] {
			t.Fatalf("%s count after real-data migration = %d; want unchanged source count %d", tableName, got, before[tableName])
		}
	}
}

func assertPostgresRealDataMigrationTablesExist(t *testing.T, db *gorm.DB) {
	t.Helper()

	missingTables := make([]string, 0)
	for _, tableName := range postgresIntegrationExpectedTables {
		if db.Migrator().HasTable(tableName) {
			continue
		}
		missingTables = append(missingTables, tableName)
	}
	if len(missingTables) > 0 {
		t.Fatalf("real-data database missing commons tables: %v", missingTables)
	}
}

func assertPostgresRealDataExpectedIndexesExist(t *testing.T, db *gorm.DB) {
	t.Helper()

	indexCatalog, err := loadCurrentSchemaIndexCatalog(db)
	if err != nil {
		t.Fatalf("loadCurrentSchemaIndexCatalog() error = %v; want nil", err)
	}

	missingIndexes := make([]string, 0)
	for _, target := range postgresIntegrationExpectedIndexes {
		key := postgresIntegrationIndexKey(target.TableName, target.IndexName)
		if _, ok := indexCatalog[key]; ok {
			continue
		}
		missingIndexes = append(missingIndexes, key)
	}
	if len(missingIndexes) > 0 {
		t.Fatalf("real-data database missing commons indexes: %v", missingIndexes)
	}

	assertExpectedIndexDefinitions(t, indexCatalog, postgresIntegrationExpectedSQLManagedIndexDefinitions)
	assertExpectedIndexDefinitions(t, indexCatalog, postgresIntegrationExpectedGORMIndexDefinitions)
}

func assertPostgresRealDataReadModelCounts(t *testing.T, db *gorm.DB) {
	t.Helper()

	var result struct {
		EntryCount                int64 `gorm:"column:entry_count"`
		EntryFormCount            int64 `gorm:"column:entry_form_count"`
		SearchTermCount           int64 `gorm:"column:search_term_count"`
		HeadwordSearchTermCount   int64 `gorm:"column:headword_search_term_count"`
		FormAliasSearchTermCount  int64 `gorm:"column:form_alias_search_term_count"`
		FeaturedSourceCount       int64 `gorm:"column:featured_source_count"`
		FeaturedCandidateCount    int64 `gorm:"column:featured_candidate_count"`
		InvalidSearchTermKindRows int64 `gorm:"column:invalid_search_term_kind_rows"`
	}
	if err := db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM entries) AS entry_count,
			(SELECT COUNT(*) FROM entry_forms) AS entry_form_count,
			(SELECT COUNT(*) FROM entry_search_terms) AS search_term_count,
			(SELECT COUNT(*) FROM entry_search_terms WHERE term_kind = 'headword') AS headword_search_term_count,
			(SELECT COUNT(*) FROM entry_search_terms WHERE term_kind IN ('form', 'alias')) AS form_alias_search_term_count,
			(
				SELECT COUNT(*)
				FROM entries e
				JOIN entry_learning_signals ls ON ls.entry_id = e.id
				WHERE ls.frequency_rank > 0 OR ls.cefr_level > 0
			) AS featured_source_count,
			(SELECT COUNT(*) FROM featured_candidates) AS featured_candidate_count,
			(SELECT COUNT(*) FROM entry_search_terms WHERE term_kind NOT IN ('headword', 'form', 'alias')) AS invalid_search_term_kind_rows
	`).Scan(&result).Error; err != nil {
		t.Fatalf("load real-data read model counts: %v", err)
	}

	wantSearchTermCount := result.EntryCount + result.EntryFormCount
	if result.SearchTermCount != wantSearchTermCount {
		t.Fatalf("entry_search_terms count = %d; want entries + entry_forms = %d", result.SearchTermCount, wantSearchTermCount)
	}
	if result.HeadwordSearchTermCount != result.EntryCount {
		t.Fatalf("entry_search_terms headword count = %d; want entries count %d", result.HeadwordSearchTermCount, result.EntryCount)
	}
	if result.FormAliasSearchTermCount != result.EntryFormCount {
		t.Fatalf("entry_search_terms form/alias count = %d; want entry_forms count %d", result.FormAliasSearchTermCount, result.EntryFormCount)
	}
	if result.FeaturedCandidateCount != result.FeaturedSourceCount {
		t.Fatalf("featured_candidates count = %d; want source join count %d", result.FeaturedCandidateCount, result.FeaturedSourceCount)
	}
	if result.InvalidSearchTermKindRows != 0 {
		t.Fatalf("entry_search_terms invalid term_kind rows = %d; want 0", result.InvalidSearchTermKindRows)
	}
}

func assertPostgresRealDataSearchTermsMatchSource(t *testing.T, db *gorm.DB) {
	t.Helper()

	var result struct {
		MismatchCount int64 `gorm:"column:mismatch_count"`
	}
	if err := db.Raw(`
		WITH expected AS (
			SELECT
				e.id AS entry_id,
				e.headword,
				e.headword AS term_text,
				e.normalized_headword AS normalized_term,
				'headword'::text AS term_kind,
				1::smallint AS term_rank,
				e.pos,
				e.is_multiword,
				COALESCE(ls.frequency_rank, 0) AS frequency_rank,
				COALESCE(ls.frequency_count, 0) AS frequency_count,
				COALESCE(ls.cefr_level, 0) AS cefr_level,
				COALESCE(ls.oxford_level, 0) AS oxford_level,
				COALESCE(ls.cet_level, 0) AS cet_level,
				COALESCE(ls.collins_stars, 0) AS collins_stars,
				COALESCE(ls.school_level, 0) AS school_level
			FROM entries e
			LEFT JOIN entry_learning_signals ls ON ls.entry_id = e.id
			UNION ALL
			SELECT
				e.id AS entry_id,
				e.headword,
				f.form_text AS term_text,
				f.normalized_form AS normalized_term,
				f.relation_kind AS term_kind,
				2::smallint AS term_rank,
				e.pos,
				POSITION(' ' IN f.form_text) > 0 OR POSITION('-' IN f.form_text) > 0 AS is_multiword,
				COALESCE(ls.frequency_rank, 0) AS frequency_rank,
				COALESCE(ls.frequency_count, 0) AS frequency_count,
				COALESCE(ls.cefr_level, 0) AS cefr_level,
				COALESCE(ls.oxford_level, 0) AS oxford_level,
				COALESCE(ls.cet_level, 0) AS cet_level,
				COALESCE(ls.collins_stars, 0) AS collins_stars,
				COALESCE(ls.school_level, 0) AS school_level
			FROM entry_forms f
			JOIN entries e ON e.id = f.entry_id
			LEFT JOIN entry_learning_signals ls ON ls.entry_id = e.id
		),
		actual AS (
			SELECT
				entry_id,
				headword,
				term_text,
				normalized_term,
				term_kind,
				term_rank,
				pos,
				is_multiword,
				frequency_rank,
				frequency_count,
				cefr_level,
				oxford_level,
				cet_level,
				collins_stars,
				school_level
			FROM entry_search_terms
		),
		mismatches AS (
			(SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
			UNION ALL
			(SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
		)
		SELECT COUNT(*) AS mismatch_count FROM mismatches
	`).Scan(&result).Error; err != nil {
		t.Fatalf("compare entry_search_terms with real source data: %v", err)
	}
	if result.MismatchCount != 0 {
		t.Fatalf("entry_search_terms real-data mismatch count = %d; want 0", result.MismatchCount)
	}
}

func assertPostgresRealDataFeaturedCandidatesMatchSource(t *testing.T, db *gorm.DB) {
	t.Helper()

	var result struct {
		MismatchCount int64 `gorm:"column:mismatch_count"`
	}
	if err := db.Raw(`
		WITH expected AS (
			SELECT
				e.id AS entry_id,
				e.headword,
				e.normalized_headword,
				e.is_multiword,
				e.pos,
				ls.frequency_rank,
				ls.cefr_level,
				ls.oxford_level,
				ls.cet_level,
				ls.collins_stars,
				CASE WHEN ls.frequency_rank > 0 THEN ls.frequency_rank ELSE 999999 END AS quality_rank
			FROM entries e
			JOIN entry_learning_signals ls ON ls.entry_id = e.id
			WHERE ls.frequency_rank > 0 OR ls.cefr_level > 0
		),
		actual AS (
			SELECT
				entry_id,
				headword,
				normalized_headword,
				is_multiword,
				pos,
				frequency_rank,
				cefr_level,
				oxford_level,
				cet_level,
				collins_stars,
				quality_rank
			FROM featured_candidates
		),
		mismatches AS (
			(SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
			UNION ALL
			(SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
		)
		SELECT COUNT(*) AS mismatch_count FROM mismatches
	`).Scan(&result).Error; err != nil {
		t.Fatalf("compare featured_candidates with real source data: %v", err)
	}
	if result.MismatchCount != 0 {
		t.Fatalf("featured_candidates real-data mismatch count = %d; want 0", result.MismatchCount)
	}
}
