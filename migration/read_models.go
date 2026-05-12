package migration

import (
	"fmt"

	"gorm.io/gorm"
)

// RefreshReadModels rebuilds derived read models from the canonical source tables.
func RefreshReadModels(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("nil database handle")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`TRUNCATE TABLE entry_search_terms, featured_candidates RESTART IDENTITY`).Error; err != nil {
			return fmt.Errorf("truncate read models: %w", err)
		}
		if err := tx.Exec(entrySearchTermsHeadwordInsertSQL).Error; err != nil {
			return fmt.Errorf("insert entry_search_terms headwords: %w", err)
		}
		if err := tx.Exec(entrySearchTermsFormInsertSQL).Error; err != nil {
			return fmt.Errorf("insert entry_search_terms forms: %w", err)
		}
		if err := tx.Exec(featuredCandidatesInsertSQL).Error; err != nil {
			return fmt.Errorf("insert featured_candidates: %w", err)
		}
		if err := tx.Exec(featuredCandidatesUniqueHeadwordIndexSQL).Error; err != nil {
			return fmt.Errorf("ensure featured_candidates normalized headword uniqueness: %w", err)
		}

		return nil
	})
}

const entrySearchTermsHeadwordInsertSQL = `
INSERT INTO entry_search_terms (
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
)
SELECT
	e.id,
	e.headword,
	e.headword,
	e.normalized_headword,
	'headword',
	1,
	e.pos,
	e.is_multiword,
	COALESCE(ls.frequency_rank, 0),
	COALESCE(ls.frequency_count, 0),
	COALESCE(ls.cefr_level, 0),
	COALESCE(ls.oxford_level, 0),
	COALESCE(ls.cet_level, 0),
	COALESCE(ls.collins_stars, 0),
	COALESCE(ls.school_level, 0)
FROM entries e
LEFT JOIN entry_learning_signals ls ON ls.entry_id = e.id
ORDER BY e.id`

const entrySearchTermsFormInsertSQL = `
INSERT INTO entry_search_terms (
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
)
SELECT
	e.id,
	e.headword,
	f.form_text,
	f.normalized_form,
	f.relation_kind,
	2,
	e.pos,
	POSITION(' ' IN f.form_text) > 0 OR POSITION('-' IN f.form_text) > 0,
	COALESCE(ls.frequency_rank, 0),
	COALESCE(ls.frequency_count, 0),
	COALESCE(ls.cefr_level, 0),
	COALESCE(ls.oxford_level, 0),
	COALESCE(ls.cet_level, 0),
	COALESCE(ls.collins_stars, 0),
	COALESCE(ls.school_level, 0)
FROM entry_forms f
JOIN entries e ON e.id = f.entry_id
LEFT JOIN entry_learning_signals ls ON ls.entry_id = e.id
ORDER BY f.entry_id, f.id`

const featuredCandidatesInsertSQL = `
WITH eligible AS (
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
		ls.school_level,
		ROW_NUMBER() OVER (
			PARTITION BY e.normalized_headword
			ORDER BY
				CASE WHEN ls.frequency_rank > 0 THEN 0 ELSE 1 END,
				ls.frequency_rank,
				CASE WHEN ls.cefr_level > 0 THEN 0 ELSE 1 END,
				ls.cefr_level DESC,
				CASE WHEN ls.collins_stars > 0 THEN 0 ELSE 1 END,
				ls.collins_stars DESC,
				CASE WHEN ls.school_level > 0 THEN 0 ELSE 1 END,
				ls.school_level,
				CASE WHEN e.headword = LOWER(e.headword) THEN 0 ELSE 1 END,
				e.id
		) AS headword_rn
	FROM entries e
	JOIN entry_learning_signals ls ON ls.entry_id = e.id
	WHERE ls.frequency_rank > 0 OR ls.cefr_level > 0 OR ls.school_level > 0
),
selected AS (
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
		school_level,
		(ROW_NUMBER() OVER (
			ORDER BY
				CASE WHEN frequency_rank > 0 THEN 0 ELSE 1 END,
				frequency_rank,
				CASE WHEN cefr_level > 0 THEN 0 ELSE 1 END,
				cefr_level DESC,
				CASE WHEN collins_stars > 0 THEN 0 ELSE 1 END,
				collins_stars DESC,
				CASE WHEN school_level > 0 THEN 0 ELSE 1 END,
				school_level,
				CASE WHEN headword = LOWER(headword) THEN 0 ELSE 1 END,
				entry_id
		))::integer AS quality_rank
	FROM eligible
	WHERE headword_rn = 1
)
INSERT INTO featured_candidates (
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
	school_level,
	quality_rank
)
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
	school_level,
	quality_rank
FROM selected
ORDER BY quality_rank`

const featuredCandidatesUniqueHeadwordIndexSQL = `
DO $$
DECLARE
	active_schema text;
	qualified_index text;
	qualified_table text;
BEGIN
	active_schema := current_schema();
	IF active_schema IS NULL OR active_schema = '' THEN
		RAISE EXCEPTION 'current_schema() returned empty while creating featured_candidates unique headword index';
	END IF;

	qualified_index := format('%I.%I', active_schema, 'idx_featured_candidates_normalized_headword');
	qualified_table := format('%I.%I', active_schema, 'featured_candidates');

	EXECUTE format('DROP INDEX IF EXISTS %s', qualified_index);
	EXECUTE format(
		'CREATE UNIQUE INDEX %I ON %s (normalized_headword)',
		'idx_featured_candidates_normalized_headword',
		qualified_table
	);
END $$`
