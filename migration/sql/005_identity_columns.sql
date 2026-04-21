DO $$
DECLARE
	active_schema text;
	qualified_table text;
	target_table text;
	max_id bigint;
	sequence_name text;
BEGIN
	active_schema := current_schema();
	IF active_schema IS NULL OR active_schema = '' THEN
		RAISE EXCEPTION 'current_schema() returned empty while repairing identity columns';
	END IF;

	FOREACH target_table IN ARRAY ARRAY[
		'import_runs',
		'entries',
		'senses',
		'sense_glosses_en',
		'sense_glosses_zh',
		'sense_labels',
		'sense_examples',
		'pronunciation_ipas',
		'pronunciation_audios',
		'entry_forms',
		'lexical_relations',
		'entry_summaries_zh'
	] LOOP
		qualified_table := format('%I.%I', active_schema, target_table);

		IF EXISTS (
			SELECT 1
			FROM information_schema.columns c
			WHERE c.table_schema = active_schema
			  AND c.table_name = target_table
			  AND c.column_name = 'id'
			  AND c.is_identity = 'NO'
		) THEN
			EXECUTE format('ALTER TABLE %s ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY', qualified_table);
		ELSIF EXISTS (
			SELECT 1
			FROM information_schema.columns c
			WHERE c.table_schema = active_schema
			  AND c.table_name = target_table
			  AND c.column_name = 'id'
			  AND c.is_identity = 'YES'
			  AND COALESCE(c.identity_generation, '') <> 'ALWAYS'
		) THEN
			EXECUTE format('ALTER TABLE %s ALTER COLUMN id SET GENERATED ALWAYS', qualified_table);
		END IF;

		SELECT pg_get_serial_sequence(qualified_table, 'id')
		INTO sequence_name;
		IF sequence_name IS NULL THEN
			RAISE EXCEPTION 'identity sequence missing for %.id in schema %', target_table, active_schema;
		END IF;

		EXECUTE format('SELECT COALESCE(MAX(id), 0) FROM %s', qualified_table) INTO max_id;
		PERFORM setval(sequence_name, CASE WHEN max_id > 0 THEN max_id ELSE 1 END, max_id > 0);
	END LOOP;
END $$;