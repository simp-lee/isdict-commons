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
