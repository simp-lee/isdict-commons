CREATE UNIQUE INDEX IF NOT EXISTS idx_entry_forms_entry_id_relation_kind_form_text_form_type
ON entry_forms (entry_id, relation_kind, form_text, COALESCE(form_type, ''));

CREATE UNIQUE INDEX IF NOT EXISTS idx_lexical_relations_entry_id_sense_id_rel_type_target_norm
ON lexical_relations (entry_id, COALESCE(sense_id, 0), relation_type, target_text_normalized);