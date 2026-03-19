package migration

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/simp-lee/isdict-commons/model"
	"gorm.io/gorm"
)

var postgresTypeCastPattern = regexp.MustCompile(`::[a-z0-9_\.\[\]]+`)

// Migrator handles database schema migrations
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// MigrateOptions migration configuration options
type MigrateOptions struct {
	DropTables     bool // Drop existing tables before migration
	SkipExtensions bool // Skip PostgreSQL extension creation
	SkipIndexes    bool // Skip index creation (for incremental migration)
	Verbose        bool // Enable verbose logging
}

// Migrate performs full database migration
func (m *Migrator) Migrate(opts *MigrateOptions) error {
	if opts == nil {
		opts = &MigrateOptions{}
	}

	m.logMigrationStart(opts)

	if err := m.runSchemaMigration(opts); err != nil {
		return err
	}

	skippedIndexes, err := m.createIndexesForMigration(opts)
	if err != nil {
		return err
	}

	m.updateStatisticsWithWarning()

	if err := m.verifyMigrationWithLogging(opts, skippedIndexes); err != nil {
		return err
	}

	m.logMigrationComplete(opts)
	return nil
}

func (m *Migrator) logMigrationStart(opts *MigrateOptions) {
	if opts.Verbose {
		log.Println("========== Starting Database Migration ==========")
	}
}

func (m *Migrator) runSchemaMigration(opts *MigrateOptions) error {
	if opts.DropTables {
		if err := m.DropAllTables(); err != nil {
			return fmt.Errorf("drop tables: %w", err)
		}
	}

	if !opts.SkipExtensions {
		if err := m.EnableExtensions(); err != nil {
			return fmt.Errorf("enable extensions: %w", err)
		}
	}

	if err := m.AutoMigrateTables(); err != nil {
		return fmt.Errorf("auto migrate tables: %w", err)
	}

	if err := m.CreateUniqueConstraints(); err != nil {
		return fmt.Errorf("create unique constraints: %w", err)
	}

	return nil
}

func (m *Migrator) createIndexesForMigration(opts *MigrateOptions) ([]string, error) {
	var skippedIndexes []string
	if !opts.SkipIndexes {
		var err error
		skippedIndexes, err = m.CreateIndexes()
		if err != nil {
			return nil, fmt.Errorf("create indexes: %w", err)
		}
	} else if opts.Verbose {
		log.Println("Step 5/7: Skipping performance index creation (SkipIndexes option enabled)")
	}

	return skippedIndexes, nil
}

func (m *Migrator) updateStatisticsWithWarning() {
	if err := m.UpdateStatistics(); err != nil {
		log.Printf("⚠️  Warning: Failed to update statistics: %v", err)
	}
}

func (m *Migrator) verifyMigrationWithLogging(opts *MigrateOptions, skippedIndexes []string) error {
	if opts.Verbose {
		log.Println("Step 7/7: Verifying migration integrity...")
	}

	status, err := m.VerifyMigration(opts, skippedIndexes)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to verify migration: %v", err)
		return fmt.Errorf("verify migration: %w", err)
	}

	if opts.Verbose {
		log.Println(status.Summary())
	}

	return m.handleMigrationVerificationStatus(status)
}

func (m *Migrator) handleMigrationVerificationStatus(status *MigrationStatus) error {
	if !status.IsComplete() {
		return fmt.Errorf("migration verification failed: missing components")
	}

	if status.HasIssues() {
		log.Printf("⚠️  Warning: Migration completed with integrity issues:")
		for _, issue := range status.Issues {
			log.Printf("   - %s", issue)
		}

		if requiredIssues := status.RequiredIndexIssues(); len(requiredIssues) > 0 {
			return fmt.Errorf("migration verification failed: required index integrity issues: %s", strings.Join(requiredIssues, "; "))
		}
	}

	return nil
}

func (m *Migrator) logMigrationComplete(opts *MigrateOptions) {
	if opts.Verbose {
		log.Println("========== Migration Complete ==========")
	}
}

// DropAllTables drops all tables in dependency order
func (m *Migrator) DropAllTables() error {
	log.Println("Step 1/7: Dropping existing tables...")

	// Drop in reverse dependency order (child -> parent)
	tables := []interface{}{
		&model.Example{},
		&model.Sense{},
		&model.Pronunciation{},
		&model.WordVariant{},
		&model.Word{},
	}

	if err := m.db.Migrator().DropTable(tables...); err != nil {
		return err
	}

	log.Println("✓ All tables dropped successfully")
	return nil
}

// EnableExtensions enables required PostgreSQL extensions
func (m *Migrator) EnableExtensions() error {
	log.Println("Step 2/7: Enabling PostgreSQL extensions...")

	// Enable pg_trgm for trigram fuzzy search
	if err := m.db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm`).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to enable pg_trgm extension: %v", err)
		log.Printf("   Fuzzy search may be slower without this extension")
		log.Printf("   Manual fix: Run as superuser: CREATE EXTENSION pg_trgm;")
		return nil // Non-fatal, continue
	}

	log.Println("✓ pg_trgm extension enabled (fuzzy search support)")
	return nil
}

// AutoMigrateTables creates tables with basic constraints
func (m *Migrator) AutoMigrateTables() error {
	log.Println("Step 3/7: Creating tables and basic indexes...")

	// Auto-migrate all tables
	if err := m.db.AutoMigrate(
		&model.Word{},
		&model.Pronunciation{},
		&model.Sense{},
		&model.Example{},
		&model.WordVariant{},
	); err != nil {
		return err
	}

	log.Println("✓ Tables created successfully")
	return nil
}

// CreateUniqueConstraints creates custom unique constraints
func (m *Migrator) CreateUniqueConstraints() error {
	log.Println("Step 4/7: Creating custom unique constraints...")

	// Constraint 1: Prevent duplicate word variants with proper NULL handling
	// This uses COALESCE to treat NULL form_type as 0 for uniqueness
	if err := m.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_word_variant_unique
		ON word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))
	`).Error; err != nil {
		return fmt.Errorf("create word variant unique index: %w", err)
	}
	log.Println("✓ Word variant unique constraint created (with NULL handling)")

	// Constraint 2: One primary pronunciation per word per accent
	if err := m.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_pronunciation_primary_unique
		ON pronunciations(word_id, accent)
		WHERE is_primary
	`).Error; err != nil {
		return fmt.Errorf("create pronunciation primary unique index: %w", err)
	}
	log.Println("✓ Pronunciation primary constraint created")

	return nil
}

// CreateIndexes creates all performance indexes
func (m *Migrator) CreateIndexes() ([]string, error) {
	log.Println("Step 5/7: Creating performance indexes...")

	skipped := make([]string, 0)

	indexes := []struct {
		name  string
		sql   string
		isTrg bool // is trigram index
	}{
		// Words table trigram indexes (GORM cannot create these)
		{
			name:  "idx_words_headword_trgm",
			sql:   `CREATE INDEX IF NOT EXISTS idx_words_headword_trgm ON words USING gin(headword_normalized gin_trgm_ops)`,
			isTrg: true,
		},
		{
			name:  "idx_words_phrase_lower_trgm",
			sql:   `CREATE INDEX IF NOT EXISTS idx_words_phrase_lower_trgm ON words USING gin((lower(headword)) gin_trgm_ops) WHERE headword LIKE '% %'`,
			isTrg: true,
		},

		// Word variants trigram indexes (GORM cannot create these)
		{
			name:  "idx_word_variants_headword_trgm",
			sql:   `CREATE INDEX IF NOT EXISTS idx_word_variants_headword_trgm ON word_variants USING gin(headword_normalized gin_trgm_ops)`,
			isTrg: true,
		},
		{
			name:  "idx_word_variants_phrase_lower_trgm",
			sql:   `CREATE INDEX IF NOT EXISTS idx_word_variants_phrase_lower_trgm ON word_variants USING gin((lower(variant_text)) gin_trgm_ops) WHERE variant_text LIKE '% %'`,
			isTrg: true,
		},
	}

	// Create indexes
	for _, idx := range indexes {
		startTime := time.Now()
		err := m.db.Exec(idx.sql).Error

		if err != nil {
			// Handle trigram index creation failure gracefully
			if idx.isTrg && m.isTrigramIndexError(err) {
				log.Printf("⚠️  Warning: Cannot create %s - pg_trgm extension not available", idx.name)
				log.Printf("   Error: %v", err)
				skipped = append(skipped, idx.name)
				continue
			}
			return skipped, fmt.Errorf("create index %s: %w", idx.name, err)
		}

		elapsed := time.Since(startTime)
		log.Printf("✓ %s created (%.2fs)", idx.name, elapsed.Seconds())
	}

	if len(skipped) > 0 {
		log.Printf("⚠️  Skipped %d index(es) due to missing pg_trgm support: %v", len(skipped), skipped)
	} else {
		log.Println("✓ All indexes created successfully")
	}

	return skipped, nil
}

// isTrigramIndexError checks if an error is related to missing pg_trgm extension
func (m *Migrator) isTrigramIndexError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// Common error patterns when pg_trgm extension is missing:
	// - "operator class "gin_trgm_ops" does not exist"
	// - "extension "pg_trgm" does not exist"
	// - "data type gin_trgm_ops does not exist"
	return strings.Contains(errMsg, "pg_trgm") ||
		strings.Contains(errMsg, "gin_trgm_ops") ||
		strings.Contains(errMsg, "trgm")
}

func requiredIndexesList() []string {
	return []string{
		"idx_pronunciation_primary_unique",
		"idx_word_variant_unique",
	}
}

func optionalPerformanceIndexesList() []string {
	return []string{
		"idx_words_headword_trgm",
		"idx_words_phrase_lower_trgm",
		"idx_word_variants_headword_trgm",
		"idx_word_variants_phrase_lower_trgm",
	}
}

// expectedIndexesList returns all custom indexes managed by the migrator
func expectedIndexesList() []string {
	indexes := make([]string, 0, len(requiredIndexesList())+len(optionalPerformanceIndexesList()))
	indexes = append(indexes, requiredIndexesList()...)
	indexes = append(indexes, optionalPerformanceIndexesList()...)
	return indexes
}

// UpdateStatistics updates table statistics for query optimization
func (m *Migrator) UpdateStatistics() error {
	log.Println("Step 6/7: Updating table statistics...")

	if err := m.db.Exec(`ANALYZE words, word_variants, pronunciations, senses, examples`).Error; err != nil {
		return err
	}

	log.Println("✓ Table statistics updated")
	return nil
}

type indexResult struct {
	IndexName string
	IndexDef  string
}

type indexInfo struct {
	TableName string
	IndexName string
	IndexDef  string
	Columns   string
}

const loadExistingIndexMapQuery = `
		SELECT
			i.indexname AS index_name,
			pg_get_indexdef(c.oid) AS index_def
		FROM pg_indexes i
		JOIN pg_namespace n ON n.nspname = i.schemaname
		JOIN pg_class c ON c.relname = i.indexname AND c.relnamespace = n.oid
		WHERE i.schemaname = 'public'
	`

// VerifyMigration verifies that all expected tables and indexes exist
func (m *Migrator) VerifyMigration(opts *MigrateOptions, skippedIndexes []string) (*MigrationStatus, error) {
	status := &MigrationStatus{
		Issues: []string{},
	}

	skipSet := buildSkippedIndexSet(opts, skippedIndexes)
	m.collectTableStatus(status)

	indexMap, err := m.loadExistingIndexMap()
	if err != nil {
		return nil, err
	}

	collectExpectedIndexStatus(status, skipSet, indexMap)
	m.checkDuplicateIndexes(status)
	if err := m.collectExtensionStatus(status); err != nil {
		return nil, err
	}

	return status, nil
}

func buildSkippedIndexSet(opts *MigrateOptions, skippedIndexes []string) map[string]struct{} {
	skipSet := make(map[string]struct{})
	if opts != nil && opts.SkipIndexes {
		for _, name := range optionalPerformanceIndexesList() {
			skipSet[name] = struct{}{}
		}
	}
	for _, name := range skippedIndexes {
		skipSet[name] = struct{}{}
	}
	return skipSet
}

func (m *Migrator) collectTableStatus(status *MigrationStatus) {
	tables := []string{"words", "word_variants", "pronunciations", "senses", "examples"}
	for _, table := range tables {
		if m.db.Migrator().HasTable(table) {
			status.Tables = append(status.Tables, table)
			continue
		}
		status.MissingTables = append(status.MissingTables, table)
	}
}

func (m *Migrator) loadExistingIndexMap() (map[string]string, error) {
	var existingIndexes []indexResult
	if err := m.db.Raw(loadExistingIndexMapQuery).Scan(&existingIndexes).Error; err != nil {
		return nil, fmt.Errorf("query indexes: %w", err)
	}

	indexMap := make(map[string]string)
	for _, idx := range existingIndexes {
		indexMap[idx.IndexName] = idx.IndexDef
	}

	return indexMap, nil
}

func collectExpectedIndexStatus(status *MigrationStatus, skipSet map[string]struct{}, indexMap map[string]string) {
	for _, name := range expectedIndexesList() {
		if _, shouldSkip := skipSet[name]; shouldSkip {
			status.SkippedIndexes = append(status.SkippedIndexes, name)
			continue
		}

		if def, exists := indexMap[name]; exists {
			status.Indexes = append(status.Indexes, name)
			status.Issues = append(status.Issues, expectedIndexDefinitionIssues(name, def)...)
		} else {
			status.MissingIndexes = append(status.MissingIndexes, name)
		}
	}
}

func expectedIndexDefinitionIssues(name, def string) []string {
	switch name {
	case "idx_word_variant_unique":
		compact := normalizeIndexDefinition(def)
		if !strings.Contains(compact, "coalesce(") {
			return []string{"idx_word_variant_unique does not use COALESCE expression (NULL handling incorrect)"}
		}

		if !isExpectedWordVariantUniqueIndexDefinition(compact) {
			return []string{"idx_word_variant_unique is not the expected unique index on word_variants(word_id, variant_text, kind, COALESCE(form_type, 0))"}
		}
	case "idx_pronunciation_primary_unique":
		compact := normalizeIndexDefinition(def)
		if !strings.Contains(compact, "create unique index") ||
			(!strings.Contains(compact, " on pronunciations") && !strings.Contains(compact, " on public.pronunciations")) ||
			!strings.Contains(compact, "(word_id, accent)") ||
			(!strings.Contains(compact, " where is_primary") && !strings.Contains(compact, " where (is_primary)")) {
			return []string{"idx_pronunciation_primary_unique is not the expected partial unique index on pronunciations(word_id, accent) WHERE is_primary"}
		}
	}

	return nil
}

func isExpectedWordVariantUniqueIndexDefinition(def string) bool {
	if !strings.Contains(def, "create unique index") {
		return false
	}

	if !strings.Contains(def, " on word_variants") && !strings.Contains(def, " on public.word_variants") {
		return false
	}

	columns, ok := extractIndexColumns(def)
	if !ok || len(columns) != 4 {
		return false
	}

	if normalizeIdentifierExpression(columns[0]) != "word_id" ||
		normalizeIdentifierExpression(columns[1]) != "variant_text" ||
		normalizeIdentifierExpression(columns[2]) != "kind" {
		return false
	}

	return isExpectedWordVariantCoalesceExpression(columns[3])
}

func extractIndexColumns(def string) ([]string, bool) {
	start := strings.Index(def, "(")
	if start == -1 {
		return nil, false
	}

	depth := 0
	for i := start; i < len(def); i++ {
		switch def[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				if strings.TrimSpace(def[i+1:]) != "" {
					return nil, false
				}
				return splitTopLevelCSV(def[start+1 : i]), true
			}
		}
	}

	return nil, false
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

func normalizeIdentifierExpression(value string) string {
	return stripWrappingParens(strings.TrimSpace(value))
}

func isExpectedWordVariantCoalesceExpression(value string) bool {
	normalized := stripWrappingParens(strings.TrimSpace(value))
	normalized = postgresTypeCastPattern.ReplaceAllString(normalized, "")
	normalized = stripWrappingParens(normalized)

	if !strings.HasPrefix(normalized, "coalesce(") || !strings.HasSuffix(normalized, ")") {
		return false
	}

	args := splitTopLevelCSV(normalized[len("coalesce(") : len(normalized)-1])
	if len(args) != 2 {
		return false
	}

	return stripWrappingParens(args[0]) == "form_type" && stripWrappingParens(args[1]) == "0"
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

func normalizeIndexDefinition(def string) string {
	normalized := strings.ToLower(def)
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")
	normalized = strings.ReplaceAll(normalized, "\"", "")
	return strings.Join(strings.Fields(normalized), " ")
}

func (m *Migrator) checkDuplicateIndexes(status *MigrationStatus) {
	var allIndexes []indexInfo
	if err := m.db.Raw(`
		SELECT 
			t.relname AS table_name,
			i.relname AS index_name,
			pg_get_indexdef(i.oid) AS index_def,
			string_agg(a.attname, ', ' ORDER BY array_position(ix.indkey, a.attnum)) AS columns
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relnamespace = 'public'::regnamespace
			AND t.relname IN ('words', 'word_variants')
			AND NOT ix.indisprimary
			AND i.relname NOT LIKE 'pg_%'
		GROUP BY t.relname, i.relname, i.oid, ix.indkey
		ORDER BY t.relname, i.relname
	`).Scan(&allIndexes).Error; err != nil {
		log.Printf("⚠️  Warning: Could not check for duplicate indexes: %v", err)
		return
	}

	type indexGroup struct {
		tableName string
		columns   string
	}

	indexGroups := make(map[indexGroup][]string)
	for _, idx := range allIndexes {
		if strings.Contains(idx.IndexDef, "gin_trgm_ops") ||
			strings.Contains(idx.IndexDef, "COALESCE") ||
			strings.Contains(idx.IndexDef, "lower(") {
			continue
		}

		key := indexGroup{
			tableName: idx.TableName,
			columns:   idx.Columns,
		}
		indexGroups[key] = append(indexGroups[key], idx.IndexName)
	}

	for key, indexNames := range indexGroups {
		if len(indexNames) <= 1 {
			continue
		}
		status.Issues = append(status.Issues,
			fmt.Sprintf("Duplicate B-tree indexes on %s(%s): %v",
				key.tableName, key.columns, indexNames))
	}
}

func (m *Migrator) collectExtensionStatus(status *MigrationStatus) error {
	type extResult struct {
		ExtName string
	}
	var extensions []extResult
	if err := m.db.Raw(`SELECT extname AS ext_name FROM pg_extension WHERE extname = 'pg_trgm'`).Scan(&extensions).Error; err != nil {
		return fmt.Errorf("query extensions: %w", err)
	}

	if len(extensions) > 0 {
		status.Extensions = append(status.Extensions, "pg_trgm")
	}

	return nil
}

// MigrationStatus represents the current migration state
type MigrationStatus struct {
	Tables         []string
	MissingTables  []string
	Indexes        []string
	MissingIndexes []string
	Extensions     []string
	Issues         []string // Integrity issues found during verification
	SkippedIndexes []string // Indexes intentionally skipped in this run
}

// IsComplete returns true if all expected components exist
func (s *MigrationStatus) IsComplete() bool {
	return len(s.MissingTables) == 0 && len(s.MissingIndexes) == 0
}

// HasIssues returns true if any integrity issues were found
func (s *MigrationStatus) HasIssues() bool {
	return len(s.Issues) > 0
}

// RequiredIndexIssues returns integrity issues for required indexes that must fail migration.
func (s *MigrationStatus) RequiredIndexIssues() []string {
	requiredNames := requiredIndexesList()
	issues := make([]string, 0)

	for _, issue := range s.Issues {
		for _, name := range requiredNames {
			if strings.Contains(issue, name) {
				issues = append(issues, issue)
				break
			}
		}
	}

	return issues
}

// Summary returns a human-readable summary
func (s *MigrationStatus) Summary() string {
	var sb strings.Builder

	sb.WriteString("📊 Migration Status:\n")
	fmt.Fprintf(&sb, "  • Tables: %d/5 created, %d missing\n", len(s.Tables), len(s.MissingTables))
	fmt.Fprintf(&sb, "  • Custom indexes: %d/6 created, %d missing\n", len(s.Indexes), len(s.MissingIndexes))
	fmt.Fprintf(&sb, "  • Extensions: %d/1 enabled\n", len(s.Extensions))
	sb.WriteString("  • Note: Basic B-tree indexes are auto-created by GORM\n")

	if len(s.MissingTables) > 0 {
		sb.WriteString("\n⚠️  Missing tables:\n")
		for _, t := range s.MissingTables {
			fmt.Fprintf(&sb, "  - %s\n", t)
		}
	}

	if len(s.MissingIndexes) > 0 {
		sb.WriteString("\n⚠️  Missing custom indexes:\n")
		for _, i := range s.MissingIndexes {
			fmt.Fprintf(&sb, "  - %s\n", i)
		}
	}

	if len(s.SkippedIndexes) > 0 {
		sb.WriteString("\nℹ️  Skipped indexes (not verified in this run):\n")
		for _, i := range s.SkippedIndexes {
			fmt.Fprintf(&sb, "  - %s\n", i)
		}
	}

	if len(s.Issues) > 0 {
		sb.WriteString("\n⚠️  Integrity issues:\n")
		for _, issue := range s.Issues {
			fmt.Fprintf(&sb, "  - %s\n", issue)
		}
	}

	if s.IsComplete() && !s.HasIssues() {
		sb.WriteString("\n✅ Migration is complete!")
	} else if s.IsComplete() {
		sb.WriteString("\n⚠️  Migration complete but has integrity issues")
	}

	return sb.String()
}
