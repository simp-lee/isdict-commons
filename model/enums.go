package model

import "maps"

const (
	POSNoun         = "noun"
	POSVerb         = "verb"
	POSAdjective    = "adjective"
	POSAdverb       = "adverb"
	POSPronoun      = "pronoun"
	POSPreposition  = "preposition"
	POSConjunction  = "conjunction"
	POSArticle      = "article"
	POSInterjection = "interjection"
	POSDeterminer   = "determiner"
	POSNumber       = "number"
	POSParticle     = "particle"
	POSPhrasalVerb  = "phrasal_verb"
	POSPhrase       = "phrase"
	POSAbbreviation = "abbreviation"
	POSSymbol       = "symbol"
	POSName         = "name"
	POSProverb      = "proverb"
	POSCharacter    = "character"
	POSAffix        = "affix"
	POSContraction  = "contraction"
	POSPunctuation  = "punctuation"
	POSPostposition = "postposition"
)

const (
	AccentUnknown       = "unknown"
	AccentBritish       = "british"
	AccentAmerican      = "american"
	AccentAustralian    = "australian"
	AccentCanadian      = "canadian"
	AccentIrish         = "irish"
	AccentScottish      = "scottish"
	AccentNZ            = "nz"
	AccentIndian        = "indian"
	AccentSouthAfrican  = "south_african"
	AccentOtherRegional = "other"
)

const (
	LabelTypeGrammar  = "grammar"
	LabelTypeRegister = "register"
	LabelTypeRegion   = "region"
	LabelTypeTemporal = "temporal"
	LabelTypeDomain   = "domain"
	LabelTypeAttitude = "attitude"
)

const (
	GrammarLabelTransitive     = "transitive"
	GrammarLabelIntransitive   = "intransitive"
	GrammarLabelDitransitive   = "ditransitive"
	GrammarLabelAmbitransitive = "ambitransitive"
	GrammarLabelCountable      = "countable"
	GrammarLabelUncountable    = "uncountable"
	GrammarLabelPluralOnly     = "plural-only"
	GrammarLabelSingularOnly   = "singular-only"
	GrammarLabelAttributive    = "attributive"
	GrammarLabelPredicative    = "predicative"
	GrammarLabelInThePlural    = "in-the-plural"
)

const (
	RegisterLabelFormal      = "formal"
	RegisterLabelInformal    = "informal"
	RegisterLabelSlang       = "slang"
	RegisterLabelLiterary    = "literary"
	RegisterLabelPoetic      = "poetic"
	RegisterLabelVulgar      = "vulgar"
	RegisterLabelTaboo       = "taboo"
	RegisterLabelFigurative  = "figurative"
	RegisterLabelIdiomatic   = "idiomatic"
	RegisterLabelByExtension = "by-extension"
	RegisterLabelEuphemistic = "euphemistic"
)

const (
	RegionLabelUS          = "US"
	RegionLabelUK          = "UK"
	RegionLabelAustralia   = "Australia"
	RegionLabelCanada      = "Canada"
	RegionLabelNewZealand  = "New-Zealand"
	RegionLabelIreland     = "Ireland"
	RegionLabelScotland    = "Scotland"
	RegionLabelIndia       = "India"
	RegionLabelSingapore   = "Singapore"
	RegionLabelSouthAfrica = "South-Africa"
)

const (
	TemporalLabelArchaic    = "archaic"
	TemporalLabelDated      = "dated"
	TemporalLabelObsolete   = "obsolete"
	TemporalLabelRare       = "rare"
	TemporalLabelHistorical = "historical"
)

const (
	DomainLabelMedicine     = "medicine"
	DomainLabelLaw          = "law"
	DomainLabelComputing    = "computing"
	DomainLabelFinance      = "finance"
	DomainLabelBusiness     = "business"
	DomainLabelMusic        = "music"
	DomainLabelSports       = "sports"
	DomainLabelBiology      = "biology"
	DomainLabelChemistry    = "chemistry"
	DomainLabelPhysics      = "physics"
	DomainLabelEngineering  = "engineering"
	DomainLabelMathematics  = "mathematics"
	DomainLabelBotany       = "botany"
	DomainLabelZoology      = "zoology"
	DomainLabelLinguistics  = "linguistics"
	DomainLabelMilitary     = "military"
	DomainLabelArchitecture = "architecture"
	DomainLabelReligion     = "religion"
	DomainLabelPolitics     = "politics"
	DomainLabelCooking      = "cooking"
)

const (
	AttitudeLabelDerogatory   = "derogatory"
	AttitudeLabelOffensive    = "offensive"
	AttitudeLabelHumorous     = "humorous"
	AttitudeLabelApproving    = "approving"
	AttitudeLabelDisapproving = "disapproving"
)

const (
	RelationTypeSynonym = "synonym"
	RelationTypeAntonym = "antonym"
	RelationTypeDerived = "derived"
)

const (
	RelationKindForm  = "form"
	RelationKindAlias = "alias"
)

const (
	ImportRunStatusRunning   = "running"
	ImportRunStatusCompleted = "completed"
	ImportRunStatusFailed    = "failed"
)

const (
	CEFRSourceNone   = ""
	CEFRSourceOxford = "oxford"
	CEFRSourceCEFRJ  = "cefrj"
	CEFRSourceBoth   = "both"
)

const (
	CEFRLevelUnknown int16 = 0
	CEFRLevelA1      int16 = 1
	CEFRLevelA2      int16 = 2
	CEFRLevelB1      int16 = 3
	CEFRLevelB2      int16 = 4
	CEFRLevelC1      int16 = 5
	CEFRLevelC2      int16 = 6
)

const (
	OxfordLevelUnknown int16 = 0
	OxfordLevel3000    int16 = 1
	OxfordLevel5000    int16 = 2
)

const (
	CETLevelUnknown int16 = 0
	CETLevel4       int16 = 1
	CETLevel6       int16 = 2
)

const (
	SchoolLevelUnknown      int16 = 0
	SchoolLevelMiddleSchool int16 = 1
	SchoolLevelHighSchool   int16 = 2
	SchoolLevelUniversity   int16 = 3
)

const (
	CollinsStarsUnknown int16 = 0
	CollinsOneStar      int16 = 1
	CollinsTwoStars     int16 = 2
	CollinsThreeStars   int16 = 3
	CollinsFourStars    int16 = 4
	CollinsFiveStars    int16 = 5
)

var (
	posCodeToName = map[string]string{
		POSNoun:         "Noun",
		POSVerb:         "Verb",
		POSAdjective:    "Adjective",
		POSAdverb:       "Adverb",
		POSPronoun:      "Pronoun",
		POSPreposition:  "Preposition",
		POSConjunction:  "Conjunction",
		POSArticle:      "Article",
		POSInterjection: "Interjection",
		POSDeterminer:   "Determiner",
		POSNumber:       "Number",
		POSParticle:     "Particle",
		POSPhrasalVerb:  "Phrasal Verb",
		POSPhrase:       "Phrase",
		POSAbbreviation: "Abbreviation",
		POSSymbol:       "Symbol",
		POSName:         "Name",
		POSProverb:      "Proverb",
		POSCharacter:    "Character",
		POSAffix:        "Affix",
		POSContraction:  "Contraction",
		POSPunctuation:  "Punctuation",
		POSPostposition: "Postposition",
	}
	posNameToCode = invertStringMap(posCodeToName)
	validPOSCodes = keySet(posCodeToName)

	accentCodeToName = map[string]string{
		AccentUnknown:       "Unknown",
		AccentBritish:       "British (RP)",
		AccentAmerican:      "American (GA)",
		AccentAustralian:    "Australian",
		AccentCanadian:      "Canadian",
		AccentIrish:         "Irish",
		AccentScottish:      "Scottish",
		AccentNZ:            "New Zealand",
		AccentIndian:        "Indian",
		AccentSouthAfrican:  "South African",
		AccentOtherRegional: "Other Regional",
	}
	accentNameToCode = invertStringMap(accentCodeToName)

	labelTypeCodeToName = map[string]string{
		LabelTypeGrammar:  "Grammar",
		LabelTypeRegister: "Register",
		LabelTypeRegion:   "Region",
		LabelTypeTemporal: "Temporal",
		LabelTypeDomain:   "Domain",
		LabelTypeAttitude: "Attitude",
	}
	labelTypeNameToCode = invertStringMap(labelTypeCodeToName)
	validLabelTypes     = keySet(labelTypeCodeToName)

	labelCodeToNameByType = map[string]map[string]string{
		LabelTypeGrammar: {
			GrammarLabelTransitive:     "Transitive",
			GrammarLabelIntransitive:   "Intransitive",
			GrammarLabelDitransitive:   "Ditransitive",
			GrammarLabelAmbitransitive: "Ambitransitive",
			GrammarLabelCountable:      "Countable",
			GrammarLabelUncountable:    "Uncountable",
			GrammarLabelPluralOnly:     "Plural Only",
			GrammarLabelSingularOnly:   "Singular Only",
			GrammarLabelAttributive:    "Attributive",
			GrammarLabelPredicative:    "Predicative",
			GrammarLabelInThePlural:    "In The Plural",
		},
		LabelTypeRegister: {
			RegisterLabelFormal:      "Formal",
			RegisterLabelInformal:    "Informal",
			RegisterLabelSlang:       "Slang",
			RegisterLabelLiterary:    "Literary",
			RegisterLabelPoetic:      "Poetic",
			RegisterLabelVulgar:      "Vulgar",
			RegisterLabelTaboo:       "Taboo",
			RegisterLabelFigurative:  "Figurative",
			RegisterLabelIdiomatic:   "Idiomatic",
			RegisterLabelByExtension: "By Extension",
			RegisterLabelEuphemistic: "Euphemistic",
		},
		LabelTypeRegion: {
			RegionLabelUS:          "US",
			RegionLabelUK:          "UK",
			RegionLabelAustralia:   "Australia",
			RegionLabelCanada:      "Canada",
			RegionLabelNewZealand:  "New Zealand",
			RegionLabelIreland:     "Ireland",
			RegionLabelScotland:    "Scotland",
			RegionLabelIndia:       "India",
			RegionLabelSingapore:   "Singapore",
			RegionLabelSouthAfrica: "South Africa",
		},
		LabelTypeTemporal: {
			TemporalLabelArchaic:    "Archaic",
			TemporalLabelDated:      "Dated",
			TemporalLabelObsolete:   "Obsolete",
			TemporalLabelRare:       "Rare",
			TemporalLabelHistorical: "Historical",
		},
		LabelTypeDomain: {
			DomainLabelMedicine:     "Medicine",
			DomainLabelLaw:          "Law",
			DomainLabelComputing:    "Computing",
			DomainLabelFinance:      "Finance",
			DomainLabelBusiness:     "Business",
			DomainLabelMusic:        "Music",
			DomainLabelSports:       "Sports",
			DomainLabelBiology:      "Biology",
			DomainLabelChemistry:    "Chemistry",
			DomainLabelPhysics:      "Physics",
			DomainLabelEngineering:  "Engineering",
			DomainLabelMathematics:  "Mathematics",
			DomainLabelBotany:       "Botany",
			DomainLabelZoology:      "Zoology",
			DomainLabelLinguistics:  "Linguistics",
			DomainLabelMilitary:     "Military",
			DomainLabelArchitecture: "Architecture",
			DomainLabelReligion:     "Religion",
			DomainLabelPolitics:     "Politics",
			DomainLabelCooking:      "Cooking",
		},
		LabelTypeAttitude: {
			AttitudeLabelDerogatory:   "Derogatory",
			AttitudeLabelOffensive:    "Offensive",
			AttitudeLabelHumorous:     "Humorous",
			AttitudeLabelApproving:    "Approving",
			AttitudeLabelDisapproving: "Disapproving",
		},
	}
	labelNameToCodeByType = invertNestedStringMap(labelCodeToNameByType)
	validLabelCodesByType = nestedKeySet(labelCodeToNameByType)

	relationTypeCodeToName = map[string]string{
		RelationTypeSynonym: "Synonym",
		RelationTypeAntonym: "Antonym",
		RelationTypeDerived: "Derived",
	}
	relationTypeNameToCode = invertStringMap(relationTypeCodeToName)

	relationKindCodeToName = map[string]string{
		RelationKindForm:  "Form",
		RelationKindAlias: "Alias",
	}
	relationKindNameToCode = invertStringMap(relationKindCodeToName)

	importRunStatusCodeToName = map[string]string{
		ImportRunStatusRunning:   "Running",
		ImportRunStatusCompleted: "Completed",
		ImportRunStatusFailed:    "Failed",
	}
	importRunStatusNameToCode = invertStringMap(importRunStatusCodeToName)

	cefrSourceCodeToName = map[string]string{
		CEFRSourceNone:   "None",
		CEFRSourceOxford: "Oxford",
		CEFRSourceCEFRJ:  "CEFR-J",
		CEFRSourceBoth:   "Both",
	}
	cefrSourceNameToCode = invertStringMap(cefrSourceCodeToName)

	cefrLevelCodeToName = map[int16]string{
		CEFRLevelUnknown: "unknown",
		CEFRLevelA1:      "A1",
		CEFRLevelA2:      "A2",
		CEFRLevelB1:      "B1",
		CEFRLevelB2:      "B2",
		CEFRLevelC1:      "C1",
		CEFRLevelC2:      "C2",
	}
	cefrLevelNameToCode = invertInt16Map(cefrLevelCodeToName)

	oxfordLevelCodeToName = map[int16]string{
		OxfordLevelUnknown: "unknown",
		OxfordLevel3000:    "Oxford 3000",
		OxfordLevel5000:    "Oxford 5000",
	}
	oxfordLevelNameToCode = invertInt16Map(oxfordLevelCodeToName)

	cetLevelCodeToName = map[int16]string{
		CETLevelUnknown: "unknown",
		CETLevel4:       "CET-4",
		CETLevel6:       "CET-6",
	}
	cetLevelNameToCode = invertInt16Map(cetLevelCodeToName)

	schoolLevelCodeToName = map[int16]string{
		SchoolLevelUnknown:      "unknown",
		SchoolLevelMiddleSchool: "初中",
		SchoolLevelHighSchool:   "高中",
		SchoolLevelUniversity:   "大学",
	}
	schoolLevelNameToCode = invertInt16Map(schoolLevelCodeToName)

	collinsStarsCodeToName = map[int16]string{
		CollinsStarsUnknown: "unknown",
		CollinsOneStar:      "1 Star",
		CollinsTwoStars:     "2 Stars",
		CollinsThreeStars:   "3 Stars",
		CollinsFourStars:    "4 Stars",
		CollinsFiveStars:    "5 Stars",
	}
	collinsStarsNameToCode = invertInt16Map(collinsStarsCodeToName)
)

// The exported accessors below return defensive copies so callers cannot mutate
// the frozen controlled vocabularies held by this package.
func POSCodeToName() map[string]string {
	return cloneMap(posCodeToName)
}

func POSNameToCode() map[string]string {
	return cloneMap(posNameToCode)
}

func ValidPOSCodes() map[string]struct{} {
	return cloneMap(validPOSCodes)
}

func AccentCodeToName() map[string]string {
	return cloneMap(accentCodeToName)
}

func AccentNameToCode() map[string]string {
	return cloneMap(accentNameToCode)
}

func LabelTypeCodeToName() map[string]string {
	return cloneMap(labelTypeCodeToName)
}

func LabelTypeNameToCode() map[string]string {
	return cloneMap(labelTypeNameToCode)
}

func ValidLabelTypes() map[string]struct{} {
	return cloneMap(validLabelTypes)
}

func LabelCodeToNameByType() map[string]map[string]string {
	return cloneNestedStringMap(labelCodeToNameByType)
}

func LabelNameToCodeByType() map[string]map[string]string {
	return cloneNestedStringMap(labelNameToCodeByType)
}

func ValidLabelCodesByType() map[string]map[string]struct{} {
	return cloneNestedStringSet(validLabelCodesByType)
}

func RelationTypeCodeToName() map[string]string {
	return cloneMap(relationTypeCodeToName)
}

func RelationTypeNameToCode() map[string]string {
	return cloneMap(relationTypeNameToCode)
}

func RelationKindCodeToName() map[string]string {
	return cloneMap(relationKindCodeToName)
}

func RelationKindNameToCode() map[string]string {
	return cloneMap(relationKindNameToCode)
}

func ImportRunStatusCodeToName() map[string]string {
	return cloneMap(importRunStatusCodeToName)
}

func ImportRunStatusNameToCode() map[string]string {
	return cloneMap(importRunStatusNameToCode)
}

func CEFRSourceCodeToName() map[string]string {
	return cloneMap(cefrSourceCodeToName)
}

func CEFRSourceNameToCode() map[string]string {
	return cloneMap(cefrSourceNameToCode)
}

func CEFRLevelCodeToName() map[int16]string {
	return cloneMap(cefrLevelCodeToName)
}

func CEFRLevelNameToCode() map[string]int16 {
	return cloneMap(cefrLevelNameToCode)
}

func OxfordLevelCodeToName() map[int16]string {
	return cloneMap(oxfordLevelCodeToName)
}

func OxfordLevelNameToCode() map[string]int16 {
	return cloneMap(oxfordLevelNameToCode)
}

func CETLevelCodeToName() map[int16]string {
	return cloneMap(cetLevelCodeToName)
}

func CETLevelNameToCode() map[string]int16 {
	return cloneMap(cetLevelNameToCode)
}

func SchoolLevelCodeToName() map[int16]string {
	return cloneMap(schoolLevelCodeToName)
}

func SchoolLevelNameToCode() map[string]int16 {
	return cloneMap(schoolLevelNameToCode)
}

func CollinsStarsCodeToName() map[int16]string {
	return cloneMap(collinsStarsCodeToName)
}

func CollinsStarsNameToCode() map[string]int16 {
	return cloneMap(collinsStarsNameToCode)
}

func keySet(values map[string]string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for value := range values {
		set[value] = struct{}{}
	}
	return set
}

func nestedKeySet(groups map[string]map[string]string) map[string]map[string]struct{} {
	set := make(map[string]map[string]struct{}, len(groups))
	for group, values := range groups {
		set[group] = keySet(values)
	}
	return set
}

func cloneMap[K comparable, V any](source map[K]V) map[K]V {
	clone := make(map[K]V, len(source))
	maps.Copy(clone, source)

	return clone
}

func cloneNestedStringMap(source map[string]map[string]string) map[string]map[string]string {
	clone := make(map[string]map[string]string, len(source))
	for key, values := range source {
		clone[key] = cloneMap(values)
	}

	return clone
}

func cloneNestedStringSet(source map[string]map[string]struct{}) map[string]map[string]struct{} {
	clone := make(map[string]map[string]struct{}, len(source))
	for key, values := range source {
		clone[key] = cloneMap(values)
	}

	return clone
}

func invertStringMap(values map[string]string) map[string]string {
	inverted := make(map[string]string, len(values))
	for code, name := range values {
		inverted[name] = code
	}
	return inverted
}

func invertNestedStringMap(groups map[string]map[string]string) map[string]map[string]string {
	inverted := make(map[string]map[string]string, len(groups))
	for group, values := range groups {
		inverted[group] = invertStringMap(values)
	}
	return inverted
}

func invertInt16Map(values map[int16]string) map[string]int16 {
	inverted := make(map[string]int16, len(values))
	for code, name := range values {
		inverted[name] = code
	}
	return inverted
}
