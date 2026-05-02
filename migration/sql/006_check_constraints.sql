DO $$
DECLARE
	active_schema text;
	qualified_table text;
	target_constraint text;
BEGIN
	active_schema := current_schema();
	IF active_schema IS NULL OR active_schema = '' THEN
		RAISE EXCEPTION 'current_schema() returned empty while repairing check constraints';
	END IF;

	qualified_table := format('%I.%I', active_schema, 'sense_labels');

	FOR target_constraint IN
		SELECT con.conname
		FROM pg_constraint con
		JOIN pg_class cls ON cls.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
		WHERE nsp.nspname = active_schema
		  AND cls.relname = 'sense_labels'
		  AND con.contype = 'c'
		  AND pg_get_constraintdef(con.oid) LIKE '%label_type%'
	LOOP
		EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', qualified_table, target_constraint);
	END LOOP;

	EXECUTE format(
		'ALTER TABLE %s ADD CONSTRAINT chk_sense_labels_label_type CHECK (label_type IN (''grammar'',''register'',''region'',''temporal'',''domain'',''attitude'',''variety''))',
		qualified_table
	);

	qualified_table := format('%I.%I', active_schema, 'lexical_relations');

	FOR target_constraint IN
		SELECT con.conname
		FROM pg_constraint con
		JOIN pg_class cls ON cls.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
		WHERE nsp.nspname = active_schema
		  AND cls.relname = 'lexical_relations'
		  AND con.contype = 'c'
		  AND pg_get_constraintdef(con.oid) LIKE '%relation_type%'
	LOOP
		EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', qualified_table, target_constraint);
	END LOOP;

	EXECUTE format(
		'ALTER TABLE %s ADD CONSTRAINT chk_lexical_relations_relation_type CHECK (relation_type IN (''synonym'',''antonym'',''derived'',''related'',''hypernym'',''hyponym'',''coordinate_term'',''meronym'',''holonym'',''troponym'',''instance''))',
		qualified_table
	);
END $$;
