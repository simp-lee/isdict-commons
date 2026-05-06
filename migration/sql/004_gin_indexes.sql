DO $$
DECLARE
	extension_schema text;
	qualified_table text;
BEGIN
	SELECT ns.nspname
	INTO extension_schema
	FROM pg_extension ext
	JOIN pg_namespace ns ON ns.oid = ext.extnamespace
	WHERE ext.extname = 'pg_trgm';

	IF extension_schema IS NULL THEN
		RAISE EXCEPTION 'pg_trgm extension missing while creating idx_entries_normalized_headword_trgm and idx_entry_forms_normalized_form_trgm';
	END IF;

	qualified_table := format('%I.%I', current_schema(), 'entries');
	EXECUTE format(
		'CREATE INDEX IF NOT EXISTS idx_entries_normalized_headword_trgm ON %s USING gin (normalized_headword %I.gin_trgm_ops)',
		qualified_table,
		extension_schema
	);

	qualified_table := format('%I.%I', current_schema(), 'entry_forms');
	EXECUTE format(
		'CREATE INDEX IF NOT EXISTS idx_entry_forms_normalized_form_trgm ON %s USING gin (normalized_form %I.gin_trgm_ops)',
		qualified_table,
		extension_schema
	);
END $$;
