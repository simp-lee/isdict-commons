CREATE UNIQUE INDEX IF NOT EXISTS idx_pronunciation_ipas_entry_id_accent_code_primary
ON pronunciation_ipas (entry_id, accent_code)
WHERE is_primary = true;

CREATE UNIQUE INDEX IF NOT EXISTS idx_pronunciation_audios_entry_id_accent_code_primary
ON pronunciation_audios (entry_id, accent_code)
WHERE is_primary = true;

CREATE UNIQUE INDEX IF NOT EXISTS idx_sense_glosses_zh_sense_id_source_primary
ON sense_glosses_zh (sense_id, source)
WHERE is_primary = true;

CREATE INDEX IF NOT EXISTS idx_entry_forms_reverse_form_text_entry_id
ON entry_forms (form_text, entry_id)
WHERE source_relations && ARRAY['form_of','alt_of']::text[];

CREATE INDEX IF NOT EXISTS idx_entry_search_terms_frequency_rank_active
ON entry_search_terms (frequency_rank, normalized_term)
WHERE frequency_rank > 0;

CREATE INDEX IF NOT EXISTS idx_entry_search_terms_cefr_level_active
ON entry_search_terms (cefr_level, normalized_term)
WHERE cefr_level > 0;

CREATE INDEX IF NOT EXISTS idx_entry_search_terms_oxford_level_active
ON entry_search_terms (oxford_level, normalized_term)
WHERE oxford_level > 0;

CREATE INDEX IF NOT EXISTS idx_entry_search_terms_cet_level_active
ON entry_search_terms (cet_level, normalized_term)
WHERE cet_level > 0;

CREATE INDEX IF NOT EXISTS idx_entry_search_terms_school_level_active
ON entry_search_terms (school_level, normalized_term)
WHERE school_level > 0;

CREATE INDEX IF NOT EXISTS idx_entry_search_terms_collins_stars_active
ON entry_search_terms (collins_stars, normalized_term)
WHERE collins_stars > 0;
