package norm

import (
	"maps"
	"strings"

	"github.com/simp-lee/isdict-commons/model"
)

// LabelMapping binds a raw label alias to a controlled label type/code pair.
type LabelMapping struct {
	LabelType string
	LabelCode string
}

var labelAliasEntries = []struct {
	raw     string
	mapping LabelMapping
}{
	{raw: "transitive", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelTransitive}},
	{raw: "intransitive", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelIntransitive}},
	{raw: "ditransitive", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelDitransitive}},
	{raw: "ambitransitive", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelAmbitransitive}},
	{raw: "countable", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelCountable}},
	{raw: "uncountable", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelUncountable}},
	{raw: "plural-only", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelPluralOnly}},
	{raw: "singular-only", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelSingularOnly}},
	{raw: "in the plural", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelInThePlural}},
	{raw: "formal", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFormal}},
	{raw: "informal", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal}},
	{raw: "colloquial", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal}},
	{raw: "slang", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelSlang}},
	{raw: "literary", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelLiterary}},
	{raw: "poetic", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelPoetic}},
	{raw: "vulgar", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelVulgar}},
	{raw: "taboo", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelTaboo}},
	{raw: "figurative", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative}},
	{raw: "figuratively", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative}},
	{raw: "idiomatic", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelIdiomatic}},
	{raw: "by extension", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelByExtension}},
	{raw: "euphemistic", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelEuphemistic}},
	{raw: "pejorative", mapping: LabelMapping{LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelDerogatory}},
	{raw: "jocular", mapping: LabelMapping{LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous}},
	{raw: "facetious", mapping: LabelMapping{LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous}},
	{raw: "archaic", mapping: LabelMapping{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelArchaic}},
	{raw: "obsolete", mapping: LabelMapping{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelObsolete}},
	{raw: "historical", mapping: LabelMapping{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelHistorical}},
	{raw: "rare", mapping: LabelMapping{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelRare}},
	{raw: "dated", mapping: LabelMapping{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelDated}},
	{raw: "American-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelUS}},
	{raw: "British-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelUK}},
	{raw: "Australian-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelAustralia}},
	{raw: "Canadian-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelCanada}},
	{raw: "Irish-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIreland}},
	{raw: "Hiberno-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIreland}},
	{raw: "Scottish-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelScotland}},
	{raw: "Indian-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIndia}},
	{raw: "Singlish", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelSingapore}},
	{raw: "South-African-English", mapping: LabelMapping{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelSouthAfrica}},
	{raw: "plurale-tantum", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelPluralOnly}},
	{raw: "singulare-tantum", mapping: LabelMapping{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelSingularOnly}},
	{raw: "African-American Vernacular", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelAAVE}},
	{raw: "AAVE", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelAAVE}},
	{raw: "Multicultural-London-English", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelMulticulturalLondonEnglish}},
	{raw: "MLE", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelMulticulturalLondonEnglish}},
	{raw: "non-native speakers' English", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelNonNativeEnglish}},
	{raw: "non-native-English", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelNonNativeEnglish}},
	{raw: "dialectal", mapping: LabelMapping{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelDialectal}},
	{raw: "law", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelLaw}},
	{raw: "computing", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelComputing}},
	{raw: "biology", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelBiology}},
	{raw: "military", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelMilitary}},
	{raw: "sports", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelSports}},
	{raw: "games", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelGames}},
	{raw: "media", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelMedia}},
	{raw: "art", mapping: LabelMapping{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelArt}},
}

var (
	labelAliasMap           = newLabelAliasMap()
	normalizedLabelAliasMap = newNormalizedLabelAliasMap(labelAliasMap)
	labelTextPhraseMap      = newLabelTextPhraseMap()
)

// LabelAliasMap returns a defensive copy of the frozen label alias mapping contract.
func LabelAliasMap() map[string]LabelMapping {
	return cloneLabelAliasMap(labelAliasMap)
}

func newLabelAliasMap() map[string]LabelMapping {
	aliases := make(map[string]LabelMapping, len(labelAliasEntries))
	for _, entry := range labelAliasEntries {
		aliases[entry.raw] = entry.mapping
	}

	return aliases
}

func newNormalizedLabelAliasMap(source map[string]LabelMapping) map[string]LabelMapping {
	aliases := make(map[string]LabelMapping, len(source))
	for raw, mapping := range source {
		if normalized := canonicalLookupText(raw); normalized != "" {
			aliases[normalized] = mapping
		}
	}

	return aliases
}

// NormalizeLabelAlias resolves a raw label alias to its controlled label type/code.
func NormalizeLabelAlias(raw string) (labelType, labelCode string, ok bool) {
	mapping, ok := labelAliasMap[raw]
	if !ok {
		mapping, ok = normalizedLabelAliasMap[canonicalLookupText(raw)]
		if !ok {
			return "", "", false
		}
	}

	return mapping.LabelType, mapping.LabelCode, true
}

// NormalizeLabelText resolves a Wiktionary qualifier/raw_gloss label string to
// one or more controlled label mappings. Matches are allowlisted and ordered.
func NormalizeLabelText(raw string) []LabelMapping {
	parts := splitLabelText(raw)
	mappings := make([]LabelMapping, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	add := func(mapping LabelMapping) {
		key := mapping.LabelType + "\x00" + mapping.LabelCode
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		mappings = append(mappings, mapping)
	}

	for _, part := range parts {
		normalized := canonicalLookupText(part)
		if normalized == "" {
			continue
		}
		if phraseMappings, ok := labelTextPhraseMap[normalized]; ok {
			for _, mapping := range phraseMappings {
				add(mapping)
			}
			continue
		}
		if labelType, labelCode, ok := NormalizeLabelAlias(part); ok {
			add(LabelMapping{LabelType: labelType, LabelCode: labelCode})
		}
	}

	if len(mappings) == 0 {
		return nil
	}

	return mappings
}

func splitLabelText(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})
	if len(parts) == 0 {
		return nil
	}

	trimmed := parts[:0]
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			trimmed = append(trimmed, part)
		}
	}

	return trimmed
}

func newLabelTextPhraseMap() map[string][]LabelMapping {
	phrases := make(map[string][]LabelMapping, 40)
	add := func(raw string, mappings ...LabelMapping) {
		if normalized := canonicalLookupText(raw); normalized != "" {
			phrases[normalized] = append([]LabelMapping(nil), mappings...)
		}
	}
	one := func(labelType, labelCode string) LabelMapping {
		return LabelMapping{LabelType: labelType, LabelCode: labelCode}
	}

	add("African-American Vernacular", one(model.LabelTypeVariety, model.VarietyLabelAAVE))
	add("chiefly African-American Vernacular", one(model.LabelTypeVariety, model.VarietyLabelAAVE))
	add("including African-American Vernacular", one(model.LabelTypeVariety, model.VarietyLabelAAVE))
	add("Multicultural-London-English", one(model.LabelTypeVariety, model.VarietyLabelMulticulturalLondonEnglish))
	add("non-native speakers' English", one(model.LabelTypeVariety, model.VarietyLabelNonNativeEnglish))
	add("dialectal", one(model.LabelTypeVariety, model.VarietyLabelDialectal))
	add("Scots law", one(model.LabelTypeRegion, model.RegionLabelScotland), one(model.LabelTypeDomain, model.DomainLabelLaw))

	add("object-oriented programming", one(model.LabelTypeDomain, model.DomainLabelComputing))
	add("Java programming language", one(model.LabelTypeDomain, model.DomainLabelComputing))
	add("machine learning", one(model.LabelTypeDomain, model.DomainLabelComputing))
	add("artificial intelligence", one(model.LabelTypeDomain, model.DomainLabelComputing))
	add("marine biology", one(model.LabelTypeDomain, model.DomainLabelBiology))
	add("Royal Navy", one(model.LabelTypeDomain, model.DomainLabelMilitary))
	add("RAF World War II code name", one(model.LabelTypeDomain, model.DomainLabelMilitary))
	add("procedure word", one(model.LabelTypeDomain, model.DomainLabelMilitary))
	add("social media", one(model.LabelTypeDomain, model.DomainLabelMedia))

	for _, raw := range []string{
		"roller derby",
		"Australian rules football",
		"croquet",
		"field sports",
		"combat sports",
		"bobsledding",
		"ultimate frisbee",
		"American football",
	} {
		add(raw, one(model.LabelTypeDomain, model.DomainLabelSports))
	}

	for _, raw := range []string{
		"board games",
		"roleplaying games",
		"cribbage",
		"crosswording",
		"puzzles",
		"turn-based games",
		"three card brag",
		"chess",
		"video games",
	} {
		add(raw, one(model.LabelTypeDomain, model.DomainLabelGames))
	}

	for _, raw := range []string{"painting", "fine arts", "theater lighting"} {
		add(raw, one(model.LabelTypeDomain, model.DomainLabelArt))
	}

	return phrases
}

func cloneLabelAliasMap(source map[string]LabelMapping) map[string]LabelMapping {
	clone := make(map[string]LabelMapping, len(source))
	maps.Copy(clone, source)

	return clone
}
