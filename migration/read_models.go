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
	e.id,
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
	CASE WHEN ls.frequency_rank > 0 THEN ls.frequency_rank ELSE 999999 END
FROM entries e
JOIN entry_learning_signals ls ON ls.entry_id = e.id
WHERE ls.frequency_rank > 0 OR ls.cefr_level > 0 OR ls.school_level > 0
ORDER BY
	CASE WHEN ls.frequency_rank > 0 THEN ls.frequency_rank ELSE 999999 END,
	ls.cefr_level DESC,
	ls.collins_stars DESC,
	ls.school_level ASC,
	e.id`
