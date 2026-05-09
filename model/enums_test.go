package model

import "testing"

var (
	step35ExpectedPOSCodeToName = map[string]string{
		"noun":         "Noun",
		"verb":         "Verb",
		"adjective":    "Adjective",
		"adverb":       "Adverb",
		"pronoun":      "Pronoun",
		"preposition":  "Preposition",
		"conjunction":  "Conjunction",
		"article":      "Article",
		"interjection": "Interjection",
		"determiner":   "Determiner",
		"number":       "Number",
		"particle":     "Particle",
		"phrasal_verb": "Phrasal Verb",
		"phrase":       "Phrase",
		"abbreviation": "Abbreviation",
		"symbol":       "Symbol",
		"name":         "Name",
		"proverb":      "Proverb",
		"character":    "Character",
		"affix":        "Affix",
		"contraction":  "Contraction",
		"punctuation":  "Punctuation",
		"postposition": "Postposition",
	}

	step35ExpectedAccentCodeToName = map[string]string{
		"unknown":       "Unknown",
		"british":       "British (RP)",
		"american":      "American (GA)",
		"australian":    "Australian",
		"canadian":      "Canadian",
		"irish":         "Irish",
		"scottish":      "Scottish",
		"nz":            "New Zealand",
		"indian":        "Indian",
		"south_african": "South African",
		"other":         "Other Regional",
	}

	step35ExpectedLabelTypeCodeToName = map[string]string{
		"grammar":  "Grammar",
		"register": "Register",
		"region":   "Region",
		"temporal": "Temporal",
		"domain":   "Domain",
		"attitude": "Attitude",
		"variety":  "Variety",
	}

	step35ExpectedLabelCodeToNameByType = map[string]map[string]string{
		"grammar": {
			"transitive":     "Transitive",
			"intransitive":   "Intransitive",
			"ditransitive":   "Ditransitive",
			"ambitransitive": "Ambitransitive",
			"countable":      "Countable",
			"uncountable":    "Uncountable",
			"plural-only":    "Plural Only",
			"singular-only":  "Singular Only",
			"attributive":    "Attributive",
			"predicative":    "Predicative",
			"in-the-plural":  "In The Plural",
		},
		"register": {
			"formal":       "Formal",
			"informal":     "Informal",
			"slang":        "Slang",
			"literary":     "Literary",
			"poetic":       "Poetic",
			"vulgar":       "Vulgar",
			"taboo":        "Taboo",
			"figurative":   "Figurative",
			"idiomatic":    "Idiomatic",
			"by-extension": "By Extension",
			"euphemistic":  "Euphemistic",
		},
		"region": {
			"US":           "US",
			"UK":           "UK",
			"Australia":    "Australia",
			"Canada":       "Canada",
			"New-Zealand":  "New Zealand",
			"Ireland":      "Ireland",
			"Scotland":     "Scotland",
			"India":        "India",
			"Singapore":    "Singapore",
			"South-Africa": "South Africa",
		},
		"temporal": {
			"archaic":    "Archaic",
			"dated":      "Dated",
			"obsolete":   "Obsolete",
			"rare":       "Rare",
			"historical": "Historical",
		},
		"domain": {
			"medicine":          "Medicine",
			"law":               "Law",
			"computing":         "Computing",
			"finance":           "Finance",
			"business":          "Business",
			"music":             "Music",
			"sports":            "Sports",
			"biology":           "Biology",
			"chemistry":         "Chemistry",
			"physics":           "Physics",
			"engineering":       "Engineering",
			"mathematics":       "Mathematics",
			"botany":            "Botany",
			"zoology":           "Zoology",
			"linguistics":       "Linguistics",
			"military":          "Military",
			"architecture":      "Architecture",
			"religion":          "Religion",
			"politics":          "Politics",
			"cooking":           "Cooking",
			"nautical":          "Nautical",
			"astronomy":         "Astronomy",
			"geology":           "Geology",
			"aviation":          "Aviation",
			"electronics":       "Electronics",
			"psychology":        "Psychology",
			"philosophy":        "Philosophy",
			"games":             "Games",
			"media":             "Media",
			"education":         "Education",
			"transport":         "Transport",
			"automotive":        "Automotive",
			"printing":          "Printing",
			"mining":            "Mining",
			"meteorology":       "Meteorology",
			"heraldry":          "Heraldry",
			"materials-science": "Materials Science",
			"mythology":         "Mythology",
			"art":               "Art",
		},
		"attitude": {
			"derogatory":   "Derogatory",
			"offensive":    "Offensive",
			"humorous":     "Humorous",
			"approving":    "Approving",
			"disapproving": "Disapproving",
		},
		"variety": {
			"AAVE":                         "African-American Vernacular English",
			"Multicultural-London-English": "Multicultural London English",
			"non-native-English":           "Non-native English",
			"dialectal":                    "Dialectal",
		},
	}

	step35ExpectedRelationTypeCodeToName = map[string]string{
		"synonym":       "Synonym",
		"antonym":       "Antonym",
		"hypernym":      "Hypernym",
		"hyponym":       "Hyponym",
		"meronym":       "Meronym",
		"holonym":       "Holonym",
		"similar_to":    "Similar To",
		"also_see":      "Also See",
		"derivation":    "Derivation",
		"pertainym":     "Pertainym",
		"domain_topic":  "Domain Topic",
		"domain_region": "Domain Region",
		"exemplifies":   "Exemplifies",
		"attribute":     "Attribute",
		"entails":       "Entails",
		"causes":        "Causes",
		"event":         "Event",
		"agent":         "Agent",
		"result":        "Result",
		"by_means_of":   "By Means Of",
		"undergoer":     "Undergoer",
		"instrument":    "Instrument",
		"uses":          "Uses",
		"state":         "State",
		"property":      "Property",
		"location":      "Location",
		"material":      "Material",
		"vehicle":       "Vehicle",
		"participle":    "Participle",
		"body_part":     "Body Part",
		"destination":   "Destination",
	}

	step35ExpectedHeadwordRelationPOSCodeToName = map[int]string{
		1: "Noun",
		2: "Verb",
		3: "Adjective",
		4: "Adverb",
	}

	step35ExpectedOEWNSourceRelations = map[string]string{
		"members":        "",
		"antonym":        "",
		"derivation":     "",
		"pertainym":      "",
		"hypernym":       "",
		"mero_part":      "",
		"mero_member":    "",
		"mero_substance": "",
		"similar":        "",
		"also":           "",
		"domain_topic":   "",
		"domain_region":  "",
		"exemplifies":    "",
		"attribute":      "",
		"entails":        "",
		"causes":         "",
		"event":          "",
		"agent":          "",
		"result":         "",
		"by_means_of":    "",
		"undergoer":      "",
		"instrument":     "",
		"uses":           "",
		"state":          "",
		"property":       "",
		"location":       "",
		"material":       "",
		"vehicle":        "",
		"participle":     "",
		"body_part":      "",
		"destination":    "",
	}

	step35ExpectedOEWNPartOfSpeechCodeToHeadwordRelationPOSCode = map[string]int{
		"n": 1,
		"v": 2,
		"a": 3,
		"s": 3,
		"r": 4,
	}

	step35ExpectedOEWNSenseTypeToHeadwordRelationPOSCode = map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 3,
	}

	step35ExpectedRelationKindCodeToName = map[string]string{
		"form":  "Form",
		"alias": "Alias",
	}

	step35ExpectedImportRunStatusCodeToName = map[string]string{
		"running":   "Running",
		"completed": "Completed",
		"failed":    "Failed",
	}

	step35ExpectedCEFRSourceCodeToName = map[string]string{
		"":         "None",
		"oxford":   "Oxford",
		"cefrj":    "CEFR-J",
		"octanove": "Octanove",
	}

	step35ExpectedCEFRLevelCodeToName = map[int16]string{
		0: "unknown",
		1: "A1",
		2: "A2",
		3: "B1",
		4: "B2",
		5: "C1",
		6: "C2",
	}

	step35ExpectedOxfordLevelCodeToName = map[int16]string{
		0: "unknown",
		1: "Oxford 3000",
		2: "Oxford 5000",
	}

	step35ExpectedCETLevelCodeToName = map[int16]string{
		0: "unknown",
		1: "CET-4",
		2: "CET-6",
	}

	step35ExpectedSchoolLevelCodeToName = map[int16]string{
		0: "unknown",
		1: "初中",
		2: "高中",
		3: "大学",
	}

	step35ExpectedCollinsStarsCodeToName = map[int16]string{
		0: "unknown",
		1: "1 Star",
		2: "2 Stars",
		3: "3 Stars",
		4: "4 Stars",
		5: "5 Stars",
	}
)

func TestStep35POSMappingsAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, POSCodeToName(), POSNameToCode(), step35ExpectedPOSCodeToName)
	assertStringSetMatchesExpected(t, ValidPOSCodes(), step35ExpectedPOSCodeToName)
}

func TestStep35AccentMappingsAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, AccentCodeToName(), AccentNameToCode(), step35ExpectedAccentCodeToName)
}

func TestStep35LabelTypesAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, LabelTypeCodeToName(), LabelTypeNameToCode(), step35ExpectedLabelTypeCodeToName)
	assertStringSetMatchesExpected(t, ValidLabelTypes(), step35ExpectedLabelTypeCodeToName)
}

func TestStep35ControlledLabelCodesAreCompleteAndDistributed(t *testing.T) {
	t.Parallel()

	gotCodeToNameByType := LabelCodeToNameByType()
	gotNameToCodeByType := LabelNameToCodeByType()
	gotValidCodesByType := ValidLabelCodesByType()

	wantCounts := map[string]int{
		"grammar":  11,
		"register": 11,
		"region":   10,
		"temporal": 5,
		"domain":   39,
		"attitude": 5,
		"variety":  4,
	}

	if got := len(gotCodeToNameByType); got != len(step35ExpectedLabelCodeToNameByType) {
		t.Fatalf("len(LabelCodeToNameByType) = %d; want %d", got, len(step35ExpectedLabelCodeToNameByType))
	}
	if got := len(gotNameToCodeByType); got != len(step35ExpectedLabelCodeToNameByType) {
		t.Fatalf("len(LabelNameToCodeByType) = %d; want %d", got, len(step35ExpectedLabelCodeToNameByType))
	}
	if got := len(gotValidCodesByType); got != len(step35ExpectedLabelCodeToNameByType) {
		t.Fatalf("len(ValidLabelCodesByType) = %d; want %d", got, len(step35ExpectedLabelCodeToNameByType))
	}

	totalCodes := 0
	for labelType, wantCodeToName := range step35ExpectedLabelCodeToNameByType {
		gotCodeToName, ok := gotCodeToNameByType[labelType]
		if !ok {
			t.Fatalf("LabelCodeToNameByType missing label type %q", labelType)
		}

		if got := len(gotCodeToName); got != wantCounts[labelType] {
			t.Fatalf("len(LabelCodeToNameByType[%q]) = %d; want %d", labelType, got, wantCounts[labelType])
		}

		gotNameToCode, ok := gotNameToCodeByType[labelType]
		if !ok {
			t.Fatalf("LabelNameToCodeByType missing label type %q", labelType)
		}

		assertStringEnumBijection(t, gotCodeToName, gotNameToCode, wantCodeToName)

		gotValidCodes, ok := gotValidCodesByType[labelType]
		if !ok {
			t.Fatalf("ValidLabelCodesByType missing label type %q", labelType)
		}
		assertStringSetMatchesExpected(t, gotValidCodes, wantCodeToName)

		totalCodes += len(gotCodeToName)
	}

	if totalCodes != 85 {
		t.Fatalf("total controlled label codes = %d; want %d", totalCodes, 85)
	}
}

func TestStep35RelationTypesAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, RelationTypeCodeToName(), RelationTypeNameToCode(), step35ExpectedRelationTypeCodeToName)
}

func TestStep35HeadwordRelationPOSMappingsAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	assertIntEnumBijection(t, HeadwordRelationPOSCodeToName(), HeadwordRelationPOSNameToCode(), step35ExpectedHeadwordRelationPOSCodeToName)
}

func TestStep35OEWNSourceRelationsMatch2025JSONFields(t *testing.T) {
	t.Parallel()

	assertStringSetMatchesExpected(t, ValidOEWNSourceRelations(), step35ExpectedOEWNSourceRelations)
}

func TestStep35OEWNPOSMappingsMatch2025JSON(t *testing.T) {
	t.Parallel()

	assertStringIntMapMatchesExpected(t, OEWNPartOfSpeechCodeToHeadwordRelationPOSCode(), step35ExpectedOEWNPartOfSpeechCodeToHeadwordRelationPOSCode)
	assertIntIntMapMatchesExpected(t, OEWNSenseTypeToHeadwordRelationPOSCode(), step35ExpectedOEWNSenseTypeToHeadwordRelationPOSCode)

	for _, tt := range []struct {
		name     string
		rawCode  string
		wantCode int
		wantOK   bool
	}{
		{name: "noun", rawCode: OEWNPartOfSpeechCodeNoun, wantCode: HeadwordRelationPOSCodeNoun, wantOK: true},
		{name: "verb", rawCode: OEWNPartOfSpeechCodeVerb, wantCode: HeadwordRelationPOSCodeVerb, wantOK: true},
		{name: "adjective", rawCode: OEWNPartOfSpeechCodeAdjective, wantCode: HeadwordRelationPOSCodeAdjective, wantOK: true},
		{name: "satellite_adjective", rawCode: OEWNPartOfSpeechCodeSatelliteAdjective, wantCode: HeadwordRelationPOSCodeAdjective, wantOK: true},
		{name: "adverb", rawCode: OEWNPartOfSpeechCodeAdverb, wantCode: HeadwordRelationPOSCodeAdverb, wantOK: true},
		{name: "numbered_noun_entry_key", rawCode: "n-2", wantCode: HeadwordRelationPOSCodeNoun, wantOK: true},
		{name: "numbered_verb_entry_key", rawCode: "v-1", wantCode: HeadwordRelationPOSCodeVerb, wantOK: true},
		{name: "unknown", rawCode: "x", wantOK: false},
		{name: "empty", rawCode: "", wantOK: false},
	} {
		gotCode, gotOK := HeadwordRelationPOSCodeFromOEWNPartOfSpeech(tt.rawCode)
		if gotCode != tt.wantCode || gotOK != tt.wantOK {
			t.Fatalf("HeadwordRelationPOSCodeFromOEWNPartOfSpeech(%q) = (%d, %v); want (%d, %v)", tt.rawCode, gotCode, gotOK, tt.wantCode, tt.wantOK)
		}
	}

	for _, tt := range []struct {
		name     string
		rawCode  int
		wantCode int
		wantOK   bool
	}{
		{name: "noun", rawCode: OEWNSenseTypeNoun, wantCode: HeadwordRelationPOSCodeNoun, wantOK: true},
		{name: "verb", rawCode: OEWNSenseTypeVerb, wantCode: HeadwordRelationPOSCodeVerb, wantOK: true},
		{name: "adjective", rawCode: OEWNSenseTypeAdjective, wantCode: HeadwordRelationPOSCodeAdjective, wantOK: true},
		{name: "adverb", rawCode: OEWNSenseTypeAdverb, wantCode: HeadwordRelationPOSCodeAdverb, wantOK: true},
		{name: "satellite_adjective", rawCode: OEWNSenseTypeSatelliteAdjective, wantCode: HeadwordRelationPOSCodeAdjective, wantOK: true},
		{name: "unknown", rawCode: 0, wantOK: false},
	} {
		gotCode, gotOK := HeadwordRelationPOSCodeFromOEWNSenseType(tt.rawCode)
		if gotCode != tt.wantCode || gotOK != tt.wantOK {
			t.Fatalf("HeadwordRelationPOSCodeFromOEWNSenseType(%d) = (%d, %v); want (%d, %v)", tt.rawCode, gotCode, gotOK, tt.wantCode, tt.wantOK)
		}
	}
}

func TestStep35RelationKindsAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, RelationKindCodeToName(), RelationKindNameToCode(), step35ExpectedRelationKindCodeToName)
}

func TestStep35ImportRunStatusMappingsAreControlledAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, ImportRunStatusCodeToName(), ImportRunStatusNameToCode(), step35ExpectedImportRunStatusCodeToName)
}

func TestStep35CEFRSourceMappingsAreControlledAndBidirectional(t *testing.T) {
	t.Parallel()

	assertStringEnumBijection(t, CEFRSourceCodeToName(), CEFRSourceNameToCode(), step35ExpectedCEFRSourceCodeToName)
}

func TestStep35Int16AuxiliaryEnumMappingsAreCompleteAndBidirectional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		gotCodeToName  map[int16]string
		gotNameToCode  map[string]int16
		wantCodeToName map[int16]string
	}{
		{
			name:           "cefr_levels",
			gotCodeToName:  CEFRLevelCodeToName(),
			gotNameToCode:  CEFRLevelNameToCode(),
			wantCodeToName: step35ExpectedCEFRLevelCodeToName,
		},
		{
			name:           "oxford_levels",
			gotCodeToName:  OxfordLevelCodeToName(),
			gotNameToCode:  OxfordLevelNameToCode(),
			wantCodeToName: step35ExpectedOxfordLevelCodeToName,
		},
		{
			name:           "cet_levels",
			gotCodeToName:  CETLevelCodeToName(),
			gotNameToCode:  CETLevelNameToCode(),
			wantCodeToName: step35ExpectedCETLevelCodeToName,
		},
		{
			name:           "school_levels",
			gotCodeToName:  SchoolLevelCodeToName(),
			gotNameToCode:  SchoolLevelNameToCode(),
			wantCodeToName: step35ExpectedSchoolLevelCodeToName,
		},
		{
			name:           "collins_stars",
			gotCodeToName:  CollinsStarsCodeToName(),
			gotNameToCode:  CollinsStarsNameToCode(),
			wantCodeToName: step35ExpectedCollinsStarsCodeToName,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertInt16EnumBijection(t, tt.gotCodeToName, tt.gotNameToCode, tt.wantCodeToName)
		})
	}
}

func TestStep35EnumAccessorsReturnDefensiveCopies(t *testing.T) {
	t.Parallel()

	t.Run("pos", func(t *testing.T) {
		t.Parallel()

		gotCodeToName := POSCodeToName()
		gotNameToCode := POSNameToCode()
		gotValidCodes := ValidPOSCodes()

		gotCodeToName[POSNoun] = "Mutated"
		gotCodeToName["invented"] = "Invented"
		delete(gotCodeToName, POSVerb)

		gotNameToCode["Mutated"] = POSNoun
		gotNameToCode["Invented"] = "invented"
		delete(gotNameToCode, "Verb")

		gotValidCodes["invented"] = struct{}{}
		delete(gotValidCodes, POSVerb)

		assertStringEnumBijection(t, POSCodeToName(), POSNameToCode(), step35ExpectedPOSCodeToName)
		assertStringSetMatchesExpected(t, ValidPOSCodes(), step35ExpectedPOSCodeToName)
	})

	t.Run("labels", func(t *testing.T) {
		t.Parallel()

		gotCodeToNameByType := LabelCodeToNameByType()
		gotNameToCodeByType := LabelNameToCodeByType()
		gotValidCodesByType := ValidLabelCodesByType()

		gotCodeToNameByType[LabelTypeGrammar][GrammarLabelTransitive] = "Mutated"
		gotCodeToNameByType[LabelTypeGrammar]["invented"] = "Invented"
		delete(gotCodeToNameByType, LabelTypeDomain)

		gotNameToCodeByType[LabelTypeGrammar]["Mutated"] = GrammarLabelTransitive
		gotNameToCodeByType[LabelTypeGrammar]["Invented"] = "invented"
		delete(gotNameToCodeByType, LabelTypeRegister)

		gotValidCodesByType[LabelTypeGrammar]["invented"] = struct{}{}
		delete(gotValidCodesByType, LabelTypeTemporal)

		freshCodeToNameByType := LabelCodeToNameByType()
		freshNameToCodeByType := LabelNameToCodeByType()
		freshValidCodesByType := ValidLabelCodesByType()

		if got := len(freshCodeToNameByType); got != len(step35ExpectedLabelCodeToNameByType) {
			t.Fatalf("len(LabelCodeToNameByType) = %d; want %d", got, len(step35ExpectedLabelCodeToNameByType))
		}
		if got := len(freshNameToCodeByType); got != len(step35ExpectedLabelCodeToNameByType) {
			t.Fatalf("len(LabelNameToCodeByType) = %d; want %d", got, len(step35ExpectedLabelCodeToNameByType))
		}
		if got := len(freshValidCodesByType); got != len(step35ExpectedLabelCodeToNameByType) {
			t.Fatalf("len(ValidLabelCodesByType) = %d; want %d", got, len(step35ExpectedLabelCodeToNameByType))
		}

		for labelType, wantCodeToName := range step35ExpectedLabelCodeToNameByType {
			assertStringEnumBijection(t, freshCodeToNameByType[labelType], freshNameToCodeByType[labelType], wantCodeToName)
			assertStringSetMatchesExpected(t, freshValidCodesByType[labelType], wantCodeToName)
		}
	})

	t.Run("school_levels", func(t *testing.T) {
		t.Parallel()

		gotCodeToName := SchoolLevelCodeToName()
		gotNameToCode := SchoolLevelNameToCode()

		gotCodeToName[SchoolLevelMiddleSchool] = "Mutated"
		gotNameToCode["Mutated"] = SchoolLevelMiddleSchool
		delete(gotNameToCode, "初中")

		assertInt16EnumBijection(t, SchoolLevelCodeToName(), SchoolLevelNameToCode(), step35ExpectedSchoolLevelCodeToName)
	})

	t.Run("headword_relation_pos", func(t *testing.T) {
		t.Parallel()

		gotCodeToName := HeadwordRelationPOSCodeToName()
		gotNameToCode := HeadwordRelationPOSNameToCode()

		gotCodeToName[HeadwordRelationPOSCodeNoun] = "Mutated"
		gotNameToCode["Mutated"] = HeadwordRelationPOSCodeNoun
		delete(gotNameToCode, "Noun")

		assertIntEnumBijection(t, HeadwordRelationPOSCodeToName(), HeadwordRelationPOSNameToCode(), step35ExpectedHeadwordRelationPOSCodeToName)
	})

	t.Run("oewn_source_relations", func(t *testing.T) {
		t.Parallel()

		gotValidRelations := ValidOEWNSourceRelations()
		gotValidRelations["related"] = struct{}{}
		delete(gotValidRelations, OEWNSourceRelationMembers)

		assertStringSetMatchesExpected(t, ValidOEWNSourceRelations(), step35ExpectedOEWNSourceRelations)
	})

	t.Run("oewn_pos_mapping", func(t *testing.T) {
		t.Parallel()

		gotPartOfSpeechMap := OEWNPartOfSpeechCodeToHeadwordRelationPOSCode()
		gotSenseTypeMap := OEWNSenseTypeToHeadwordRelationPOSCode()

		gotPartOfSpeechMap[OEWNPartOfSpeechCodeNoun] = HeadwordRelationPOSCodeAdverb
		gotPartOfSpeechMap["x"] = HeadwordRelationPOSCodeNoun
		gotSenseTypeMap[OEWNSenseTypeNoun] = HeadwordRelationPOSCodeAdverb
		gotSenseTypeMap[0] = HeadwordRelationPOSCodeNoun

		assertStringIntMapMatchesExpected(t, OEWNPartOfSpeechCodeToHeadwordRelationPOSCode(), step35ExpectedOEWNPartOfSpeechCodeToHeadwordRelationPOSCode)
		assertIntIntMapMatchesExpected(t, OEWNSenseTypeToHeadwordRelationPOSCode(), step35ExpectedOEWNSenseTypeToHeadwordRelationPOSCode)
	})
}

func assertStringEnumBijection(t *testing.T, gotCodeToName, gotNameToCode, wantCodeToName map[string]string) {
	t.Helper()

	if got := len(gotCodeToName); got != len(wantCodeToName) {
		t.Fatalf("len(codeToName) = %d; want %d", got, len(wantCodeToName))
	}
	if got := len(gotNameToCode); got != len(wantCodeToName) {
		t.Fatalf("len(nameToCode) = %d; want %d", got, len(wantCodeToName))
	}

	for code, wantName := range wantCodeToName {
		gotName, ok := gotCodeToName[code]
		if !ok {
			t.Fatalf("codeToName missing code %q", code)
		}
		if gotName != wantName {
			t.Fatalf("codeToName[%q] = %q; want %q", code, gotName, wantName)
		}

		gotCode, ok := gotNameToCode[wantName]
		if !ok {
			t.Fatalf("nameToCode missing name %q", wantName)
		}
		if gotCode != code {
			t.Fatalf("nameToCode[%q] = %q; want %q", wantName, gotCode, code)
		}
	}

	for code := range gotCodeToName {
		if _, ok := wantCodeToName[code]; !ok {
			t.Fatalf("codeToName contains unexpected code %q", code)
		}
	}
	for name := range gotNameToCode {
		found := false
		for _, wantName := range wantCodeToName {
			if wantName == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("nameToCode contains unexpected name %q", name)
		}
	}
}

func assertInt16EnumBijection(t *testing.T, gotCodeToName map[int16]string, gotNameToCode map[string]int16, wantCodeToName map[int16]string) {
	t.Helper()

	if got := len(gotCodeToName); got != len(wantCodeToName) {
		t.Fatalf("len(codeToName) = %d; want %d", got, len(wantCodeToName))
	}
	if got := len(gotNameToCode); got != len(wantCodeToName) {
		t.Fatalf("len(nameToCode) = %d; want %d", got, len(wantCodeToName))
	}

	for code, wantName := range wantCodeToName {
		gotName, ok := gotCodeToName[code]
		if !ok {
			t.Fatalf("codeToName missing code %d", code)
		}
		if gotName != wantName {
			t.Fatalf("codeToName[%d] = %q; want %q", code, gotName, wantName)
		}

		gotCode, ok := gotNameToCode[wantName]
		if !ok {
			t.Fatalf("nameToCode missing name %q", wantName)
		}
		if gotCode != code {
			t.Fatalf("nameToCode[%q] = %d; want %d", wantName, gotCode, code)
		}
	}

	for code := range gotCodeToName {
		if _, ok := wantCodeToName[code]; !ok {
			t.Fatalf("codeToName contains unexpected code %d", code)
		}
	}
	for name := range gotNameToCode {
		found := false
		for _, wantName := range wantCodeToName {
			if wantName == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("nameToCode contains unexpected name %q", name)
		}
	}
}

func assertIntEnumBijection(t *testing.T, gotCodeToName map[int]string, gotNameToCode map[string]int, wantCodeToName map[int]string) {
	t.Helper()

	if got := len(gotCodeToName); got != len(wantCodeToName) {
		t.Fatalf("len(codeToName) = %d; want %d", got, len(wantCodeToName))
	}
	if got := len(gotNameToCode); got != len(wantCodeToName) {
		t.Fatalf("len(nameToCode) = %d; want %d", got, len(wantCodeToName))
	}

	for code, wantName := range wantCodeToName {
		gotName, ok := gotCodeToName[code]
		if !ok {
			t.Fatalf("codeToName missing code %d", code)
		}
		if gotName != wantName {
			t.Fatalf("codeToName[%d] = %q; want %q", code, gotName, wantName)
		}

		gotCode, ok := gotNameToCode[wantName]
		if !ok {
			t.Fatalf("nameToCode missing name %q", wantName)
		}
		if gotCode != code {
			t.Fatalf("nameToCode[%q] = %d; want %d", wantName, gotCode, code)
		}
	}

	for code := range gotCodeToName {
		if _, ok := wantCodeToName[code]; !ok {
			t.Fatalf("codeToName contains unexpected code %d", code)
		}
	}
	for name := range gotNameToCode {
		found := false
		for _, wantName := range wantCodeToName {
			if wantName == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("nameToCode contains unexpected name %q", name)
		}
	}
}

func assertStringIntMapMatchesExpected(t *testing.T, got, want map[string]int) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d; want %d", len(got), len(want))
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Fatalf("map missing key %q", key)
		}
		if gotValue != wantValue {
			t.Fatalf("map[%q] = %d; want %d", key, gotValue, wantValue)
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok {
			t.Fatalf("map contains unexpected key %q", key)
		}
	}
}

func assertIntIntMapMatchesExpected(t *testing.T, got, want map[int]int) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d; want %d", len(got), len(want))
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Fatalf("map missing key %d", key)
		}
		if gotValue != wantValue {
			t.Fatalf("map[%d] = %d; want %d", key, gotValue, wantValue)
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok {
			t.Fatalf("map contains unexpected key %d", key)
		}
	}
}

func assertStringSetMatchesExpected(t *testing.T, gotSet map[string]struct{}, wantCodeToName map[string]string) {
	t.Helper()

	if got := len(gotSet); got != len(wantCodeToName) {
		t.Fatalf("len(validSet) = %d; want %d", got, len(wantCodeToName))
	}

	for code := range wantCodeToName {
		if _, ok := gotSet[code]; !ok {
			t.Fatalf("validSet missing code %q", code)
		}
	}

	for code := range gotSet {
		if _, ok := wantCodeToName[code]; !ok {
			t.Fatalf("validSet contains unexpected code %q", code)
		}
	}
}
