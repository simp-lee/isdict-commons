package model

import (
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

type foreignKeyAutoIncrementExpectation struct {
	name      string
	model     any
	fieldName string
}

type gormTagExpectation struct {
	name         string
	model        any
	fieldName    string
	wantFragment string
}

type foreignKeyPrimaryKeyExpectation struct {
	name             string
	model            any
	fieldName        string
	relationshipName string
	relatedModel     any
	wantDBName       string
}

func TestIdentityReferencedForeignKeysExplicitlyDisableAutoIncrement(t *testing.T) {
	t.Parallel()

	tests := []foreignKeyAutoIncrementExpectation{
		{name: "entry_source_run_id", model: Entry{}, fieldName: "SourceRunID"},
		{name: "sense_entry_id", model: Sense{}, fieldName: "EntryID"},
		{name: "sense_gloss_en_sense_id", model: SenseGlossEN{}, fieldName: "SenseID"},
		{name: "sense_gloss_zh_sense_id", model: SenseGlossZH{}, fieldName: "SenseID"},
		{name: "sense_gloss_zh_source_run_id", model: SenseGlossZH{}, fieldName: "SourceRunID"},
		{name: "sense_label_sense_id", model: SenseLabel{}, fieldName: "SenseID"},
		{name: "sense_example_sense_id", model: SenseExample{}, fieldName: "SenseID"},
		{name: "entry_definition_entry_id", model: EntryDefinition{}, fieldName: "EntryID"},
		{name: "entry_definition_sense_id", model: EntryDefinition{}, fieldName: "SenseID"},
		{name: "entry_definition_source_run_id", model: EntryDefinition{}, fieldName: "SourceRunID"},
		{name: "entry_example_entry_id", model: EntryExample{}, fieldName: "EntryID"},
		{name: "entry_example_sense_id", model: EntryExample{}, fieldName: "SenseID"},
		{name: "entry_example_source_run_id", model: EntryExample{}, fieldName: "SourceRunID"},
		{name: "pronunciation_ipa_entry_id", model: PronunciationIPA{}, fieldName: "EntryID"},
		{name: "pronunciation_audio_entry_id", model: PronunciationAudio{}, fieldName: "EntryID"},
		{name: "entry_form_entry_id", model: EntryForm{}, fieldName: "EntryID"},
		{name: "headword_relation_edge_import_run_id", model: HeadwordRelationEdge{}, fieldName: "ImportRunID"},
		{name: "entry_search_term_entry_id", model: EntrySearchTerm{}, fieldName: "EntryID"},
		{name: "entry_learning_signal_cefr_run_id", model: EntryLearningSignal{}, fieldName: "CEFRRunID"},
		{name: "entry_learning_signal_oxford_run_id", model: EntryLearningSignal{}, fieldName: "OxfordRunID"},
		{name: "entry_learning_signal_cet_run_id", model: EntryLearningSignal{}, fieldName: "CETRunID"},
		{name: "entry_learning_signal_school_run_id", model: EntryLearningSignal{}, fieldName: "SchoolRunID"},
		{name: "entry_learning_signal_frequency_run_id", model: EntryLearningSignal{}, fieldName: "FrequencyRunID"},
		{name: "entry_learning_signal_collins_run_id", model: EntryLearningSignal{}, fieldName: "CollinsRunID"},
		{name: "entry_cefr_source_signal_cefr_run_id", model: EntryCEFRSourceSignal{}, fieldName: "CEFRRunID"},
		{name: "sense_learning_signal_cefr_run_id", model: SenseLearningSignal{}, fieldName: "CEFRRunID"},
		{name: "sense_learning_signal_oxford_run_id", model: SenseLearningSignal{}, fieldName: "OxfordRunID"},
		{name: "sense_cefr_source_signal_cefr_run_id", model: SenseCEFRSourceSignal{}, fieldName: "CEFRRunID"},
		{name: "entry_summary_entry_id", model: EntrySummaryZH{}, fieldName: "EntryID"},
		{name: "entry_summary_source_run_id", model: EntrySummaryZH{}, fieldName: "SourceRunID"},
		{name: "entry_etymology_source_run_id", model: EntryEtymology{}, fieldName: "SourceRunID"},
		{name: "featured_candidate_entry_id", model: FeaturedCandidate{}, fieldName: "EntryID"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field, ok := reflect.TypeOf(tt.model).FieldByName(tt.fieldName)
			if !ok {
				t.Fatalf("%T.%s field not found", tt.model, tt.fieldName)
			}

			tag := field.Tag.Get("gorm")
			if !strings.Contains(tag, "autoIncrement:false") {
				t.Fatalf("%T.%s gorm tag = %q; want autoIncrement:false", tt.model, tt.fieldName, tag)
			}

			parsedSchema := mustParseModelSchema(t, tt.model)
			parsedField := parsedSchema.LookUpField(tt.fieldName)
			if parsedField == nil {
				t.Fatalf("schema field %T.%s not found", tt.model, tt.fieldName)
			}

			if parsedField.AutoIncrement {
				t.Fatalf("schema field %T.%s AutoIncrement = true; want false", tt.model, tt.fieldName)
			}

			if strings.Contains(strings.ToLower(string(parsedField.DataType)), "identity") {
				t.Fatalf("schema field %T.%s DataType = %q; want no copied identity clause", tt.model, tt.fieldName, parsedField.DataType)
			}
		})
	}
}

func TestStep34GORMTagContracts(t *testing.T) {
	t.Parallel()

	tests := []gormTagExpectation{
		{name: "import_run_source_dump_date", model: ImportRun{}, fieldName: "SourceDumpDate", wantFragment: "type:date"},
		{name: "entry_pos", model: Entry{}, fieldName: "Pos", wantFragment: "type:text"},
		{name: "pronunciation_ipa_accent_code", model: PronunciationIPA{}, fieldName: "AccentCode", wantFragment: "type:text"},
		{name: "entry_form_source_relations", model: EntryForm{}, fieldName: "SourceRelations", wantFragment: "type:text[]"},
		{name: "entry_search_term_normalized_term", model: EntrySearchTerm{}, fieldName: "NormalizedTerm", wantFragment: "idx_entry_search_terms_normalized_term"},
		{name: "featured_candidate_quality_rank", model: FeaturedCandidate{}, fieldName: "QualityRank", wantFragment: "idx_featured_candidates_quality_rank"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field := mustStructField(t, tt.model, tt.fieldName)
			tag := field.Tag.Get("gorm")
			if !strings.Contains(tag, tt.wantFragment) {
				t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
			}
		})
	}
}

func TestHeadwordRelationEdgeGORMContracts(t *testing.T) {
	t.Parallel()

	if got := (HeadwordRelationEdge{}).TableName(); got != "headword_relation_edges" {
		t.Fatalf("HeadwordRelationEdge.TableName() = %q; want %q", got, "headword_relation_edges")
	}

	relationTypeTag := mustStructField(t, HeadwordRelationEdge{}, "RelationType").Tag.Get("gorm")
	for _, fragment := range []string{
		"index:idx_headword_relation_edges_source_headword_pos_type,priority:3",
		"uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:3",
	} {
		if !strings.Contains(relationTypeTag, fragment) {
			t.Fatalf("HeadwordRelationEdge.RelationType gorm tag = %q; want fragment %q", relationTypeTag, fragment)
		}
	}

	for _, tt := range []gormTagExpectation{
		{name: "source_headword_query_index", model: HeadwordRelationEdge{}, fieldName: "SourceHeadwordNormalized", wantFragment: "idx_headword_relation_edges_source_headword_pos_type,priority:1"},
		{name: "target_headword_query_index", model: HeadwordRelationEdge{}, fieldName: "TargetHeadwordNormalized", wantFragment: "idx_headword_relation_edges_target_headword_pos,priority:1"},
		{name: "source_relation_type_check", model: HeadwordRelationEdge{}, fieldName: "SourceRelationType", wantFragment: "check:source_relation_type IN ('members','antonym','derivation','pertainym','hypernym','mero_part','mero_member','mero_substance','similar','also','domain_topic','domain_region','exemplifies','attribute','entails','causes','event','agent','result','by_means_of','undergoer','instrument','uses','state','property','location','material','vehicle','participle','body_part','destination')"},
		{name: "source_pos_check", model: HeadwordRelationEdge{}, fieldName: "SourcePOSCode", wantFragment: "check:source_pos_code IN (1,2,3,4)"},
		{name: "target_pos_check", model: HeadwordRelationEdge{}, fieldName: "TargetPOSCode", wantFragment: "check:target_pos_code IN (1,2,3,4)"},
		{name: "source_synset_check", model: HeadwordRelationEdge{}, fieldName: "SourceSynsetID", wantFragment: "check:source_synset_id <> ''"},
		{name: "target_synset_check", model: HeadwordRelationEdge{}, fieldName: "TargetSynsetID", wantFragment: "check:target_synset_id <> ''"},
		{name: "import_run_fk", model: HeadwordRelationEdge{}, fieldName: "ImportRunID", wantFragment: "autoIncrement:false;not null;index:idx_headword_relation_edges_import_run_id"},
	} {
		field := mustStructField(t, tt.model, tt.fieldName)
		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, tt.wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
		}
	}

	const checkPrefix = "check:relation_type IN ("
	start := strings.Index(relationTypeTag, checkPrefix)
	if start < 0 {
		t.Fatalf("HeadwordRelationEdge.RelationType gorm tag = %q; want relation_type check", relationTypeTag)
	}

	checkValues := relationTypeTag[start+len(checkPrefix):]
	end := strings.Index(checkValues, ")")
	if end < 0 {
		t.Fatalf("HeadwordRelationEdge.RelationType gorm tag = %q; want closed relation_type check", relationTypeTag)
	}

	got := make(map[string]struct{})
	for _, rawValue := range strings.Split(checkValues[:end], ",") {
		value := strings.Trim(rawValue, "'")
		got[value] = struct{}{}
	}

	want := RelationTypeCodeToName()
	if len(got) != len(want) {
		t.Fatalf("HeadwordRelationEdge.RelationType check values = %v; want %d controlled relation types", got, len(want))
	}
	for code := range want {
		if _, ok := got[code]; !ok {
			t.Fatalf("HeadwordRelationEdge.RelationType check values = %v; missing controlled relation type %q", got, code)
		}
	}
}

func TestSenseSchemaContract(t *testing.T) {
	t.Parallel()

	for _, fieldName := range []string{"DefinitionEN", "DefinitionZH"} {
		if _, ok := reflect.TypeOf(Sense{}).FieldByName(fieldName); ok {
			t.Fatalf("Sense must not expose %s; english and chinese glosses belong in dedicated gloss tables", fieldName)
		}
	}

	parsedSchema := mustParseModelSchema(t, Sense{})
	gotColumns := append([]string(nil), parsedSchema.DBNames...)
	sort.Strings(gotColumns)

	wantColumns := []string{"entry_id", "id", "sense_order"}
	sort.Strings(wantColumns)

	if !reflect.DeepEqual(gotColumns, wantColumns) {
		t.Fatalf("Sense schema columns = %v; want exactly %v", gotColumns, wantColumns)
	}
}

func TestSenseGlossZHExposesSourceRunIDField(t *testing.T) {
	t.Parallel()

	field := mustStructField(t, SenseGlossZH{}, "SourceRunID")
	if field.Type.Kind() != reflect.Int64 {
		t.Fatalf("SenseGlossZH.SourceRunID kind = %s; want %s", field.Type.Kind(), reflect.Int64)
	}

	parsedSchema := mustParseModelSchema(t, SenseGlossZH{})
	parsedField := parsedSchema.LookUpField("SourceRunID")
	if parsedField == nil {
		t.Fatal("schema field SenseGlossZH.SourceRunID not found")
	}

	if parsedField.DBName != "source_run_id" {
		t.Fatalf("schema field SenseGlossZH.SourceRunID DBName = %q; want %q", parsedField.DBName, "source_run_id")
	}
}

func TestSenseGlossZHChineseDisplayTextUniqueIndexContract(t *testing.T) {
	t.Parallel()

	const indexName = "idx_sense_glosses_zh_sense_id_text_zh_hans"
	tests := []gormTagExpectation{
		{name: "sense_id", model: SenseGlossZH{}, fieldName: "SenseID", wantFragment: "uniqueIndex:" + indexName + ",priority:1"},
		{name: "text_zh_hans", model: SenseGlossZH{}, fieldName: "TextZHHans", wantFragment: "uniqueIndex:" + indexName + ",priority:2"},
	}

	for _, tt := range tests {
		field := mustStructField(t, tt.model, tt.fieldName)
		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, tt.wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
		}
	}

	parsedSchema := mustParseModelSchema(t, SenseGlossZH{})
	index := parsedSchema.LookIndex(indexName)
	if index == nil {
		t.Fatalf("SenseGlossZH schema index %q not found", indexName)
	}
	if !strings.EqualFold(index.Class, "UNIQUE") {
		t.Fatalf("SenseGlossZH schema index %q class = %q; want UNIQUE", indexName, index.Class)
	}

	gotColumns := make([]string, 0, len(index.Fields))
	for _, field := range index.Fields {
		if field.Field == nil {
			gotColumns = append(gotColumns, "")
			continue
		}
		gotColumns = append(gotColumns, field.Field.DBName)
	}
	wantColumns := []string{"sense_id", "text_zh_hans"}
	if !reflect.DeepEqual(gotColumns, wantColumns) {
		t.Fatalf("SenseGlossZH schema index %q columns = %v; want %v", indexName, gotColumns, wantColumns)
	}
}

func TestEntryDefinitionSchemaContract(t *testing.T) {
	t.Parallel()

	if got := (EntryDefinition{}).TableName(); got != "entry_definitions" {
		t.Fatalf("EntryDefinition.TableName() = %q; want %q", got, "entry_definitions")
	}

	parsedSchema := mustParseModelSchema(t, EntryDefinition{})
	for _, fieldName := range []string{
		"ID",
		"EntryID",
		"SenseID",
		"POS",
		"Source",
		"SourceRunID",
		"DefinitionOrder",
		"TextZHHans",
		"TextEN",
		"NormalizedZHHansKey",
		"NormalizedENKey",
		"UpdatedAt",
	} {
		if parsedSchema.LookUpField(fieldName) == nil {
			t.Fatalf("EntryDefinition schema field %s not found", fieldName)
		}
	}

	assertNullableInt64ForeignKeyField(t, EntryDefinition{}, "SenseID")
	requireBelongsToRelationship(t, EntryDefinition{}, parsedSchema, "Entry", Entry{})
	requireBelongsToRelationship(t, EntryDefinition{}, parsedSchema, "Sense", Sense{})
	requireBelongsToRelationship(t, EntryDefinition{}, parsedSchema, "SourceRun", ImportRun{})

	for _, tt := range []gormTagExpectation{
		{name: "entry_id_index", model: EntryDefinition{}, fieldName: "EntryID", wantFragment: "index:idx_entry_definitions_entry_id"},
		{name: "entry_id_order_index", model: EntryDefinition{}, fieldName: "EntryID", wantFragment: "index:idx_entry_definitions_entry_id_definition_order,priority:1"},
		{name: "entry_id_unique_zh", model: EntryDefinition{}, fieldName: "EntryID", wantFragment: "uniqueIndex:idx_entry_definitions_entry_id_pos_normalized_zh_hans_key,priority:1"},
		{name: "sense_id_nullable_index", model: EntryDefinition{}, fieldName: "SenseID", wantFragment: "index:idx_entry_definitions_sense_id"},
		{name: "pos_unique_zh", model: EntryDefinition{}, fieldName: "POS", wantFragment: "uniqueIndex:idx_entry_definitions_entry_id_pos_normalized_zh_hans_key,priority:2"},
		{name: "source_updated_index", model: EntryDefinition{}, fieldName: "Source", wantFragment: "index:idx_entry_definitions_source_updated_at,priority:1"},
		{name: "source_run_fk", model: EntryDefinition{}, fieldName: "SourceRunID", wantFragment: "autoIncrement:false;not null;index:idx_entry_definitions_source_run_id"},
		{name: "definition_order_check", model: EntryDefinition{}, fieldName: "DefinitionOrder", wantFragment: "check:definition_order >= 1"},
		{name: "definition_order_index", model: EntryDefinition{}, fieldName: "DefinitionOrder", wantFragment: "index:idx_entry_definitions_entry_id_definition_order,priority:2"},
		{name: "zh_text_check", model: EntryDefinition{}, fieldName: "TextZHHans", wantFragment: "check:text_zh_hans <> ''"},
		{name: "zh_key_check", model: EntryDefinition{}, fieldName: "NormalizedZHHansKey", wantFragment: "check:normalized_zh_hans_key <> ''"},
		{name: "zh_key_unique", model: EntryDefinition{}, fieldName: "NormalizedZHHansKey", wantFragment: "uniqueIndex:idx_entry_definitions_entry_id_pos_normalized_zh_hans_key,priority:3"},
		{name: "updated_desc_index", model: EntryDefinition{}, fieldName: "UpdatedAt", wantFragment: "index:idx_entry_definitions_source_updated_at,priority:2,sort:desc"},
	} {
		field := mustStructField(t, tt.model, tt.fieldName)
		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, tt.wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
		}
	}

	for _, tt := range []gormTagExpectation{
		{name: "entry_cascade", model: EntryDefinition{}, fieldName: "Entry", wantFragment: "constraint:OnDelete:CASCADE"},
		{name: "sense_set_null", model: EntryDefinition{}, fieldName: "Sense", wantFragment: "constraint:OnDelete:SET NULL"},
		{name: "source_run_restrict", model: EntryDefinition{}, fieldName: "SourceRun", wantFragment: "constraint:OnDelete:RESTRICT"},
	} {
		field := mustStructField(t, tt.model, tt.fieldName)
		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, tt.wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
		}
	}
}

func TestEntryExampleSchemaContract(t *testing.T) {
	t.Parallel()

	if got := (EntryExample{}).TableName(); got != "entry_examples" {
		t.Fatalf("EntryExample.TableName() = %q; want %q", got, "entry_examples")
	}

	parsedSchema := mustParseModelSchema(t, EntryExample{})
	for _, fieldName := range []string{
		"ID",
		"EntryID",
		"SenseID",
		"Source",
		"SourceRunID",
		"ExampleOrder",
		"SentenceEN",
		"SentenceZHHans",
		"NormalizedSentenceENKey",
		"NormalizedSentenceZHKey",
		"UpdatedAt",
	} {
		if parsedSchema.LookUpField(fieldName) == nil {
			t.Fatalf("EntryExample schema field %s not found", fieldName)
		}
	}

	assertNullableInt64ForeignKeyField(t, EntryExample{}, "SenseID")
	requireBelongsToRelationship(t, EntryExample{}, parsedSchema, "Entry", Entry{})
	requireBelongsToRelationship(t, EntryExample{}, parsedSchema, "Sense", Sense{})
	requireBelongsToRelationship(t, EntryExample{}, parsedSchema, "SourceRun", ImportRun{})

	for _, tt := range []gormTagExpectation{
		{name: "entry_id_index", model: EntryExample{}, fieldName: "EntryID", wantFragment: "index:idx_entry_examples_entry_id"},
		{name: "entry_id_order_index", model: EntryExample{}, fieldName: "EntryID", wantFragment: "index:idx_entry_examples_entry_id_example_order,priority:1"},
		{name: "entry_id_unique_en", model: EntryExample{}, fieldName: "EntryID", wantFragment: "uniqueIndex:idx_entry_examples_entry_id_normalized_sentence_en_key,priority:1"},
		{name: "sense_id_nullable_index", model: EntryExample{}, fieldName: "SenseID", wantFragment: "index:idx_entry_examples_sense_id"},
		{name: "source_updated_index", model: EntryExample{}, fieldName: "Source", wantFragment: "index:idx_entry_examples_source_updated_at,priority:1"},
		{name: "source_run_fk", model: EntryExample{}, fieldName: "SourceRunID", wantFragment: "autoIncrement:false;not null;index:idx_entry_examples_source_run_id"},
		{name: "example_order_check", model: EntryExample{}, fieldName: "ExampleOrder", wantFragment: "check:example_order >= 1"},
		{name: "example_order_index", model: EntryExample{}, fieldName: "ExampleOrder", wantFragment: "index:idx_entry_examples_entry_id_example_order,priority:2"},
		{name: "sentence_en_check", model: EntryExample{}, fieldName: "SentenceEN", wantFragment: "check:sentence_en <> ''"},
		{name: "sentence_en_key_check", model: EntryExample{}, fieldName: "NormalizedSentenceENKey", wantFragment: "check:normalized_sentence_en_key <> ''"},
		{name: "sentence_en_key_unique", model: EntryExample{}, fieldName: "NormalizedSentenceENKey", wantFragment: "uniqueIndex:idx_entry_examples_entry_id_normalized_sentence_en_key,priority:2"},
		{name: "updated_desc_index", model: EntryExample{}, fieldName: "UpdatedAt", wantFragment: "index:idx_entry_examples_source_updated_at,priority:2,sort:desc"},
	} {
		field := mustStructField(t, tt.model, tt.fieldName)
		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, tt.wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
		}
	}

	for _, tt := range []gormTagExpectation{
		{name: "entry_cascade", model: EntryExample{}, fieldName: "Entry", wantFragment: "constraint:OnDelete:CASCADE"},
		{name: "sense_set_null", model: EntryExample{}, fieldName: "Sense", wantFragment: "constraint:OnDelete:SET NULL"},
		{name: "source_run_restrict", model: EntryExample{}, fieldName: "SourceRun", wantFragment: "constraint:OnDelete:RESTRICT"},
	} {
		field := mustStructField(t, tt.model, tt.fieldName)
		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, tt.wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
		}
	}
}

func TestForeignKeysAsPrimaryKeysRemainOneToOneContracts(t *testing.T) {
	t.Parallel()

	tests := []foreignKeyPrimaryKeyExpectation{
		{
			name:             "entry_learning_signal_entry_id",
			model:            EntryLearningSignal{},
			fieldName:        "EntryID",
			relationshipName: "Entry",
			relatedModel:     Entry{},
			wantDBName:       "entry_id",
		},
		{
			name:             "sense_learning_signal_sense_id",
			model:            SenseLearningSignal{},
			fieldName:        "SenseID",
			relationshipName: "Sense",
			relatedModel:     Sense{},
			wantDBName:       "sense_id",
		},
		{
			name:             "entry_etymology_entry_id",
			model:            EntryEtymology{},
			fieldName:        "EntryID",
			relationshipName: "Entry",
			relatedModel:     Entry{},
			wantDBName:       "entry_id",
		},
		{
			name:             "featured_candidate_entry_id",
			model:            FeaturedCandidate{},
			fieldName:        "EntryID",
			relationshipName: "Entry",
			relatedModel:     Entry{},
			wantDBName:       "entry_id",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsedSchema := mustParseModelSchema(t, tt.model)
			assertOneToOnePrimaryKeyFieldContract(t, tt.model, parsedSchema, tt.fieldName)
			assertOneToOnePrimaryKeyTagContract(t, tt.model, tt.fieldName)
			relationship := requireBelongsToRelationship(t, tt.model, parsedSchema, tt.relationshipName, tt.relatedModel)
			assertOneToOneRelationshipReference(t, tt.model, tt.relationshipName, tt.fieldName, tt.wantDBName, relationship)
		})
	}
}

func TestCEFRSourceSignalTableNameContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		model     interface{ TableName() string }
		wantTable string
	}{
		{name: "entry_source_evidence", model: EntryCEFRSourceSignal{}, wantTable: "entry_cefr_source_signals"},
		{name: "sense_source_evidence", model: SenseCEFRSourceSignal{}, wantTable: "sense_cefr_source_signals"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.model.TableName(); got != tt.wantTable {
				t.Fatalf("%T.TableName() = %q; want %q", tt.model, got, tt.wantTable)
			}
		})
	}
}

func TestReadModelTableNameContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		model     interface{ TableName() string }
		wantTable string
	}{
		{name: "entry_search_terms", model: EntrySearchTerm{}, wantTable: "entry_search_terms"},
		{name: "featured_candidates", model: FeaturedCandidate{}, wantTable: "featured_candidates"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.model.TableName(); got != tt.wantTable {
				t.Fatalf("%T.TableName() = %q; want %q", tt.model, got, tt.wantTable)
			}
		})
	}
}

func TestEntrySearchTermSchemaContract(t *testing.T) {
	t.Parallel()

	parsedSchema := mustParseModelSchema(t, EntrySearchTerm{})
	for _, fieldName := range []string{
		"ID",
		"EntryID",
		"Headword",
		"TermText",
		"NormalizedTerm",
		"TermKind",
		"TermRank",
		"Pos",
		"IsMultiword",
		"FrequencyRank",
		"FrequencyCount",
		"CEFRLevel",
		"OxfordLevel",
		"CETLevel",
		"CollinsStars",
		"SchoolLevel",
	} {
		if parsedSchema.LookUpField(fieldName) == nil {
			t.Fatalf("EntrySearchTerm schema field %s not found", fieldName)
		}
	}
	if parsedSchema.LookUpField("SchoolRunID") != nil {
		t.Fatal("EntrySearchTerm must not expose SchoolRunID; read models do not expose internal provenance")
	}

	termKindTag := mustStructField(t, EntrySearchTerm{}, "TermKind").Tag.Get("gorm")
	if !strings.Contains(termKindTag, "check:term_kind IN ('headword','form','alias')") {
		t.Fatalf("EntrySearchTerm.TermKind gorm tag = %q; want controlled term_kind check", termKindTag)
	}
}

func TestFeaturedCandidateSchemaContract(t *testing.T) {
	t.Parallel()

	parsedSchema := mustParseModelSchema(t, FeaturedCandidate{})
	assertOneToOnePrimaryKeyFieldContract(t, FeaturedCandidate{}, parsedSchema, "EntryID")

	for _, fieldName := range []string{
		"Headword",
		"NormalizedHeadword",
		"IsMultiword",
		"Pos",
		"FrequencyRank",
		"CEFRLevel",
		"OxfordLevel",
		"CETLevel",
		"CollinsStars",
		"SchoolLevel",
		"QualityRank",
	} {
		if parsedSchema.LookUpField(fieldName) == nil {
			t.Fatalf("FeaturedCandidate schema field %s not found", fieldName)
		}
	}
	if parsedSchema.LookUpField("SchoolRunID") != nil {
		t.Fatal("FeaturedCandidate must not expose SchoolRunID; read models do not expose internal provenance")
	}
}

func TestCEFRSourceSignalCompositePrimaryKeyContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		model            any
		wantPrimaryNames []string
		wantPrimaryDBs   []string
		relationshipName string
		relatedModel     any
	}{
		{
			name:             "entry_source_evidence",
			model:            EntryCEFRSourceSignal{},
			wantPrimaryNames: []string{"EntryID", "CEFRSource"},
			wantPrimaryDBs:   []string{"entry_id", "cefr_source"},
			relationshipName: "Entry",
			relatedModel:     Entry{},
		},
		{
			name:             "sense_source_evidence",
			model:            SenseCEFRSourceSignal{},
			wantPrimaryNames: []string{"SenseID", "CEFRSource"},
			wantPrimaryDBs:   []string{"sense_id", "cefr_source"},
			relationshipName: "Sense",
			relatedModel:     Sense{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsedSchema := mustParseModelSchema(t, tt.model)
			if len(parsedSchema.PrimaryFields) != len(tt.wantPrimaryNames) {
				t.Fatalf("%T primary fields = %d; want %d", tt.model, len(parsedSchema.PrimaryFields), len(tt.wantPrimaryNames))
			}
			for i, wantName := range tt.wantPrimaryNames {
				field := parsedSchema.PrimaryFields[i]
				if field.Name != wantName || field.DBName != tt.wantPrimaryDBs[i] {
					t.Fatalf("%T primary field[%d] = %s/%s; want %s/%s", tt.model, i, field.Name, field.DBName, wantName, tt.wantPrimaryDBs[i])
				}
				if field.AutoIncrement {
					t.Fatalf("%T primary field %s AutoIncrement = true; want false", tt.model, wantName)
				}
			}

			relationship := requireBelongsToRelationship(t, tt.model, parsedSchema, tt.relationshipName, tt.relatedModel)
			if len(relationship.References) == 0 {
				t.Fatalf("schema relationship %T.%s has no references", tt.model, tt.relationshipName)
			}
		})
	}
}

func TestCEFRSourceSignalGORMChecksAreSourceEvidenceOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		model        any
		fieldName    string
		wantFragment string
	}{
		{name: "entry_source", model: EntryCEFRSourceSignal{}, fieldName: "CEFRSource", wantFragment: "check:cefr_source IN ('oxford','cefrj','octanove')"},
		{name: "sense_source", model: SenseCEFRSourceSignal{}, fieldName: "CEFRSource", wantFragment: "check:cefr_source IN ('oxford','cefrj','octanove')"},
		{name: "entry_level", model: EntryCEFRSourceSignal{}, fieldName: "CEFRLevel", wantFragment: "check:cefr_level >= 0 AND cefr_level <= 6"},
		{name: "sense_level", model: SenseCEFRSourceSignal{}, fieldName: "CEFRLevel", wantFragment: "check:cefr_level >= 0 AND cefr_level <= 6"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field := mustStructField(t, tt.model, tt.fieldName)
			tag := field.Tag.Get("gorm")
			if !strings.Contains(tag, tt.wantFragment) {
				t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
			}
		})
	}
}

func TestLearningSignalGORMChecksAllowUnsetOrRealCEFRSources(t *testing.T) {
	t.Parallel()

	tests := []gormTagExpectation{
		{name: "entry_source", model: EntryLearningSignal{}, fieldName: "CEFRSource", wantFragment: "check:cefr_source IN ('','oxford','cefrj','octanove')"},
		{name: "sense_source", model: SenseLearningSignal{}, fieldName: "CEFRSource", wantFragment: "check:cefr_source IN ('','oxford','cefrj','octanove')"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field := mustStructField(t, tt.model, tt.fieldName)
			tag := field.Tag.Get("gorm")
			if !strings.Contains(tag, tt.wantFragment) {
				t.Fatalf("%T.%s gorm tag = %q; want fragment %q", tt.model, tt.fieldName, tag, tt.wantFragment)
			}
		})
	}
}

func TestEntryLearningSignalSchoolRunIDContract(t *testing.T) {
	t.Parallel()

	field := mustStructField(t, EntryLearningSignal{}, "SchoolRunID")
	if field.Type.Kind() != reflect.Ptr || field.Type.Elem().Kind() != reflect.Int64 {
		t.Fatalf("EntryLearningSignal.SchoolRunID type = %s; want *int64", field.Type)
	}

	parsedSchema := mustParseModelSchema(t, EntryLearningSignal{})
	parsedField := parsedSchema.LookUpField("SchoolRunID")
	if parsedField == nil {
		t.Fatal("schema field EntryLearningSignal.SchoolRunID not found")
	}
	if parsedField.DBName != "school_run_id" {
		t.Fatalf("schema field EntryLearningSignal.SchoolRunID DBName = %q; want %q", parsedField.DBName, "school_run_id")
	}

	requireBelongsToRelationship(t, EntryLearningSignal{}, parsedSchema, "SchoolRun", ImportRun{})
	relationshipTag := mustStructField(t, EntryLearningSignal{}, "SchoolRun").Tag.Get("gorm")
	if !strings.Contains(relationshipTag, "constraint:OnDelete:RESTRICT") {
		t.Fatalf("EntryLearningSignal.SchoolRun gorm tag = %q; want RESTRICT foreign key", relationshipTag)
	}
}

func TestCEFRSourceSignalRunIDColumnContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model any
	}{
		{name: "entry_source_evidence", model: EntryCEFRSourceSignal{}},
		{name: "sense_source_evidence", model: SenseCEFRSourceSignal{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field := mustParseModelSchema(t, tt.model).LookUpField("CEFRRunID")
			if field == nil {
				t.Fatalf("schema field %T.CEFRRunID not found", tt.model)
			}
			if field.DBName != "cefr_run_id" {
				t.Fatalf("schema field %T.CEFRRunID DBName = %q; want %q", tt.model, field.DBName, "cefr_run_id")
			}
		})
	}
}

func assertOneToOnePrimaryKeyFieldContract(t *testing.T, model any, parsedSchema *schema.Schema, fieldName string) {
	t.Helper()

	pkField := parsedSchema.LookUpField(fieldName)
	if pkField == nil {
		t.Fatalf("schema field %T.%s not found", model, fieldName)
	}
	if !pkField.PrimaryKey {
		t.Fatalf("schema field %T.%s PrimaryKey = false; want true", model, fieldName)
	}
	if pkField.AutoIncrement {
		t.Fatalf("schema field %T.%s AutoIncrement = true; want false", model, fieldName)
	}
	if strings.Contains(strings.ToLower(string(pkField.DataType)), "identity") {
		t.Fatalf("schema field %T.%s DataType = %q; want no copied identity clause", model, fieldName, pkField.DataType)
	}
}

func assertOneToOnePrimaryKeyTagContract(t *testing.T, model any, fieldName string) {
	t.Helper()

	field := mustStructField(t, model, fieldName)
	tag := field.Tag.Get("gorm")
	for _, wantFragment := range []string{"primaryKey", "autoIncrement:false", "type:bigint"} {
		if !strings.Contains(tag, wantFragment) {
			t.Fatalf("%T.%s gorm tag = %q; want fragment %q", model, fieldName, tag, wantFragment)
		}
	}
}

func assertNullableInt64ForeignKeyField(t *testing.T, model any, fieldName string) {
	t.Helper()

	field := mustStructField(t, model, fieldName)
	if field.Type.Kind() != reflect.Ptr || field.Type.Elem().Kind() != reflect.Int64 {
		t.Fatalf("%T.%s type = %s; want *int64", model, fieldName, field.Type)
	}
	tag := field.Tag.Get("gorm")
	if strings.Contains(tag, "not null") {
		t.Fatalf("%T.%s gorm tag = %q; want nullable foreign key", model, fieldName, tag)
	}

	parsedField := mustParseModelSchema(t, model).LookUpField(fieldName)
	if parsedField == nil {
		t.Fatalf("schema field %T.%s not found", model, fieldName)
	}
	if parsedField.DBName == "" {
		t.Fatalf("schema field %T.%s DBName empty; want concrete database column", model, fieldName)
	}
}

func requireBelongsToRelationship(t *testing.T, model any, parsedSchema *schema.Schema, relationshipName string, relatedModel any) *schema.Relationship {
	t.Helper()

	relationship := parsedSchema.Relationships.Relations[relationshipName]
	if relationship == nil {
		t.Fatalf("schema relationship %T.%s not found", model, relationshipName)
	}
	if relationship.Type != schema.BelongsTo {
		t.Fatalf("schema relationship %T.%s Type = %q; want %q", model, relationshipName, relationship.Type, schema.BelongsTo)
	}

	relatedSchema := mustParseModelSchema(t, relatedModel)
	if relationship.FieldSchema != nil && relationship.FieldSchema.Table == relatedSchema.Table {
		return relationship
	}

	got := "<nil>"
	if relationship.FieldSchema != nil {
		got = relationship.FieldSchema.Table
	}
	t.Fatalf("schema relationship %T.%s target table = %q; want %q", model, relationshipName, got, relatedSchema.Table)
	return nil
}

func assertOneToOneRelationshipReference(t *testing.T, model any, relationshipName, fieldName, wantDBName string, relationship *schema.Relationship) {
	t.Helper()

	for _, reference := range relationship.References {
		if reference.ForeignKey == nil || reference.PrimaryKey == nil {
			continue
		}
		if reference.ForeignKey.Name != fieldName || reference.PrimaryKey.Name != "ID" {
			continue
		}

		if !reference.ForeignKey.PrimaryKey {
			t.Fatalf("schema relationship %T.%s foreign key %s PrimaryKey = false; want true", model, relationshipName, fieldName)
		}
		if reference.ForeignKey.DBName != wantDBName {
			t.Fatalf("schema relationship %T.%s foreign key DBName = %q; want %q", model, relationshipName, reference.ForeignKey.DBName, wantDBName)
		}
		if reference.PrimaryKey.DBName != "id" {
			t.Fatalf("schema relationship %T.%s primary key DBName = %q; want %q", model, relationshipName, reference.PrimaryKey.DBName, "id")
		}

		return
	}

	t.Fatalf("schema relationship %T.%s does not reference %s -> ID", model, relationshipName, fieldName)
}

func mustStructField(t *testing.T, model any, fieldName string) reflect.StructField {
	t.Helper()

	typeOfModel := reflect.TypeOf(model)
	if typeOfModel.Kind() == reflect.Pointer {
		typeOfModel = typeOfModel.Elem()
	}

	field, ok := typeOfModel.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%T.%s field not found", model, fieldName)
	}

	return field
}

func mustParseModelSchema(t *testing.T, model any) *schema.Schema {
	t.Helper()

	parsedSchema, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("schema.Parse(%T) error = %v", model, err)
	}

	return parsedSchema
}
