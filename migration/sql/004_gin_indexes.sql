DO $$
DECLARE
	active_schema text;
	extension_schema text;
	qualified_table text;
BEGIN
	active_schema := current_schema();
	IF active_schema IS NULL OR active_schema = '' THEN
		RAISE EXCEPTION 'current_schema() returned empty while creating entry_search_terms trigram indexes';
	END IF;

	SELECT ns.nspname
	INTO extension_schema
	FROM pg_extension ext
	JOIN pg_namespace ns ON ns.oid = ext.extnamespace
	WHERE ext.extname = 'pg_trgm';

	IF extension_schema IS NULL THEN
		RAISE EXCEPTION 'pg_trgm extension missing while creating entry_search_terms trigram indexes';
	END IF;

	EXECUTE format('DROP INDEX IF EXISTS %I.%I', active_schema, 'idx_entries_normalized_headword_trgm');
	EXECUTE format('DROP INDEX IF EXISTS %I.%I', active_schema, 'idx_entry_forms_normalized_form_trgm');

	qualified_table := format('%I.%I', current_schema(), 'entry_search_terms');
	EXECUTE format(
		'CREATE INDEX IF NOT EXISTS idx_entry_search_terms_normalized_term_trgm ON %s USING gin (normalized_term %I.gin_trgm_ops)',
		qualified_table,
		extension_schema
	);

	EXECUTE format(
		'CREATE INDEX IF NOT EXISTS idx_entry_search_terms_multiword_normalized_term_trgm ON %s USING gin (normalized_term %I.gin_trgm_ops) WHERE is_multiword = true',
		qualified_table,
		extension_schema
	);
END $$;
