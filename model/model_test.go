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
		{name: "pronunciation_ipa_entry_id", model: PronunciationIPA{}, fieldName: "EntryID"},
		{name: "pronunciation_audio_entry_id", model: PronunciationAudio{}, fieldName: "EntryID"},
		{name: "entry_form_entry_id", model: EntryForm{}, fieldName: "EntryID"},
		{name: "lexical_relation_entry_id", model: LexicalRelation{}, fieldName: "EntryID"},
		{name: "lexical_relation_sense_id", model: LexicalRelation{}, fieldName: "SenseID"},
		{name: "entry_learning_signal_cefr_run_id", model: EntryLearningSignal{}, fieldName: "CEFRRunID"},
		{name: "entry_learning_signal_oxford_run_id", model: EntryLearningSignal{}, fieldName: "OxfordRunID"},
		{name: "entry_learning_signal_cet_run_id", model: EntryLearningSignal{}, fieldName: "CETRunID"},
		{name: "entry_learning_signal_frequency_run_id", model: EntryLearningSignal{}, fieldName: "FrequencyRunID"},
		{name: "entry_learning_signal_collins_run_id", model: EntryLearningSignal{}, fieldName: "CollinsRunID"},
		{name: "entry_cefr_source_signal_cefr_run_id", model: EntryCEFRSourceSignal{}, fieldName: "CEFRRunID"},
		{name: "sense_learning_signal_cefr_run_id", model: SenseLearningSignal{}, fieldName: "CEFRRunID"},
		{name: "sense_learning_signal_oxford_run_id", model: SenseLearningSignal{}, fieldName: "OxfordRunID"},
		{name: "sense_cefr_source_signal_cefr_run_id", model: SenseCEFRSourceSignal{}, fieldName: "CEFRRunID"},
		{name: "entry_summary_entry_id", model: EntrySummaryZH{}, fieldName: "EntryID"},
		{name: "entry_summary_source_run_id", model: EntrySummaryZH{}, fieldName: "SourceRunID"},
		{name: "entry_etymology_source_run_id", model: EntryEtymology{}, fieldName: "SourceRunID"},
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

func TestLexicalRelationRelationTypeGORMCheckAllowsControlledRelationTypes(t *testing.T) {
	t.Parallel()

	field := mustStructField(t, LexicalRelation{}, "RelationType")
	tag := field.Tag.Get("gorm")

	for _, fragment := range []string{
		"index:idx_lexical_relations_entry_id_relation_type,priority:2",
		"index:idx_lexical_relations_sense_id_relation_type,priority:2",
	} {
		if !strings.Contains(tag, fragment) {
			t.Fatalf("LexicalRelation.RelationType gorm tag = %q; want fragment %q", tag, fragment)
		}
	}

	const checkPrefix = "check:relation_type IN ("
	start := strings.Index(tag, checkPrefix)
	if start < 0 {
		t.Fatalf("LexicalRelation.RelationType gorm tag = %q; want relation_type check", tag)
	}

	checkValues := tag[start+len(checkPrefix):]
	end := strings.Index(checkValues, ")")
	if end < 0 {
		t.Fatalf("LexicalRelation.RelationType gorm tag = %q; want closed relation_type check", tag)
	}

	got := make(map[string]struct{})
	for _, rawValue := range strings.Split(checkValues[:end], ",") {
		value := strings.Trim(rawValue, "'")
		got[value] = struct{}{}
	}

	want := RelationTypeCodeToName()
	if len(got) != len(want) {
		t.Fatalf("LexicalRelation.RelationType check values = %v; want %d controlled relation types", got, len(want))
	}
	for code := range want {
		if _, ok := got[code]; !ok {
			t.Fatalf("LexicalRelation.RelationType check values = %v; missing controlled relation type %q", got, code)
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
		{name: "entry_source", model: EntryCEFRSourceSignal{}, fieldName: "CEFRSource", wantFragment: "check:cefr_source IN ('oxford','cefrj')"},
		{name: "sense_source", model: SenseCEFRSourceSignal{}, fieldName: "CEFRSource", wantFragment: "check:cefr_source IN ('oxford','cefrj')"},
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
