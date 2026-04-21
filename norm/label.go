package norm

import (
	"maps"

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
	{raw: "colloquial", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal}},
	{raw: "figuratively", mapping: LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative}},
	{raw: "pejorative", mapping: LabelMapping{LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelDerogatory}},
	{raw: "jocular", mapping: LabelMapping{LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous}},
	{raw: "facetious", mapping: LabelMapping{LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous}},
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
}

var labelAliasMap = newLabelAliasMap()

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

// NormalizeLabelAlias resolves a raw label alias to its controlled label type/code.
func NormalizeLabelAlias(raw string) (labelType, labelCode string, ok bool) {
	mapping, ok := labelAliasMap[raw]
	if !ok {
		return "", "", false
	}

	return mapping.LabelType, mapping.LabelCode, true
}

func cloneLabelAliasMap(source map[string]LabelMapping) map[string]LabelMapping {
	clone := make(map[string]LabelMapping, len(source))
	maps.Copy(clone, source)

	return clone
}
