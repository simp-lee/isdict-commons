CREATE UNIQUE INDEX IF NOT EXISTS idx_entry_forms_entry_id_relation_kind_form_text_form_type
ON entry_forms (entry_id, relation_kind, form_text, COALESCE(form_type, ''));
