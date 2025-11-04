package model

// PosCodeToName maps database POS enum values to canonical names
var PosCodeToName = map[int]string{
	0:  "unknown",
	1:  "noun",
	2:  "verb",
	3:  "adjective",
	4:  "adverb",
	5:  "pronoun",
	6:  "preposition",
	7:  "conjunction",
	8:  "article",
	9:  "interjection",
	10: "determiner",
	11: "numeral",
	12: "modal",
	13: "auxiliary",
	14: "particle",
	15: "phrasal_verb",
	16: "idiom",
	17: "abbreviation",
	18: "character",
	19: "affix",
	20: "contraction",
	21: "punctuation",
	22: "postposition",
}

// FormTypeMapping maps form type tags to database enum values
var FormTypeMapping = map[string]int{
	"past":                  1, // FormPast
	"past_tense":            1, // FormPast
	"preterite":             1, // FormPast
	"past_participle":       2, // FormPastParticiple
	"past_part":             2, // FormPastParticiple
	"present_3rd":           3, // FormPresent3rd
	"3rd_person_singular":   3, // FormPresent3rd
	"third_person_singular": 3, // FormPresent3rd
	"gerund":                4, // FormGerund
	"present_participle":    4, // FormGerund
	"ing_form":              4, // FormGerund
	"plural":                5, // FormPlural
	"comparative":           6, // FormComparative
	"superlative":           7, // FormSuperlative
	"possessive":            8, // FormPossessive
	"genitive":              8, // FormPossessive
	"infinitive":            9, // FormInfinitive
	"to_infinitive":         9, // FormInfinitive
}

// FormTypeCodeToName maps form type enum codes to their canonical names
var FormTypeCodeToName = map[int]string{
	1: "past",
	2: "past_participle",
	3: "present_3rd",
	4: "gerund",
	5: "plural",
	6: "comparative",
	7: "superlative",
	8: "possessive",
	9: "infinitive",
}

// VariantKindToName maps variant kind enum codes to their canonical names
var VariantKindToName = map[int]string{
	int(VariantForm):  "form",
	int(VariantAlias): "alias",
}

// AccentCodeToName maps accent codes to lowercase names (for API consistency)
var AccentCodeToName = map[int]string{
	0:  "unknown",
	1:  "british",
	2:  "american",
	3:  "australian",
	4:  "newzealand",
	5:  "canadian",
	6:  "irish",
	7:  "scottish",
	8:  "indian",
	9:  "southafrican",
	10: "other",
}

// PosNameToCode is the reverse mapping of PosCodeToName (for API query parameter parsing)
var PosNameToCode = makeReverseMap(PosCodeToName)

// AccentNameToCode is the reverse mapping of AccentCodeToName (for API query parameter parsing)
var AccentNameToCode = makeReverseMap(AccentCodeToName)

// makeReverseMap creates a reverse mapping from name to code
func makeReverseMap(forward map[int]string) map[string]int {
	reverse := make(map[string]int, len(forward))
	for code, name := range forward {
		reverse[name] = code
	}
	return reverse
}

// OxfordLevelCodeToName maps Oxford level enum codes to their canonical names
var OxfordLevelCodeToName = map[int]string{
	0: "",
	1: "Oxford 3000",
	2: "Oxford 5000",
}

// GetPOSName returns the canonical name for a POS code
func GetPOSName(code int) string {
	if name, ok := PosCodeToName[code]; ok {
		return name
	}
	return "unknown"
}

// GetAccentName returns the canonical name for an accent code
func GetAccentName(code int) string {
	if name, ok := AccentCodeToName[code]; ok {
		return name
	}
	return "unknown"
}

// GetFormTypeName returns the canonical name for a form type code
func GetFormTypeName(code int) string {
	if name, ok := FormTypeCodeToName[code]; ok {
		return name
	}
	return ""
}

// GetVariantKindName returns the canonical name for a variant kind code
func GetVariantKindName(code int) string {
	if name, ok := VariantKindToName[code]; ok {
		return name
	}
	return "unknown"
}

// ParsePOS converts a POS name to its code
func ParsePOS(name string) (int, bool) {
	code, ok := PosNameToCode[name]
	return code, ok
}

// ParseAccent converts an accent name to its code
func ParseAccent(name string) (int, bool) {
	code, ok := AccentNameToCode[name]
	return code, ok
}

// GetOxfordLevelName returns the canonical name for an Oxford level code
func GetOxfordLevelName(code int) string {
	if name, ok := OxfordLevelCodeToName[code]; ok {
		return name
	}
	return ""
}

// OxfordLevelFromString converts oxford source string to level code
func OxfordLevelFromString(source string) int {
	switch source {
	case "oxford_3000":
		return 1
	case "oxford_5000":
		return 2
	default:
		return 0
	}
}
