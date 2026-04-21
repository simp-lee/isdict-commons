package norm

import (
	"maps"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedLabelAliasMap = map[string]LabelMapping{
	"colloquial":            {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal},
	"figuratively":          {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative},
	"pejorative":            {LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelDerogatory},
	"jocular":               {LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous},
	"facetious":             {LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous},
	"American-English":      {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelUS},
	"British-English":       {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelUK},
	"Australian-English":    {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelAustralia},
	"Canadian-English":      {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelCanada},
	"Irish-English":         {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIreland},
	"Hiberno-English":       {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIreland},
	"Scottish-English":      {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelScotland},
	"Indian-English":        {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIndia},
	"Singlish":              {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelSingapore},
	"South-African-English": {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelSouthAfrica},
	"plurale-tantum":        {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelPluralOnly},
	"singulare-tantum":      {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelSingularOnly},
}

func TestLabelAliasMapContract(t *testing.T) {
	t.Parallel()

	got := LabelAliasMap()
	if !maps.Equal(got, expectedLabelAliasMap) {
		t.Fatalf("LabelAliasMap() mismatch: got %#v; want %#v", got, expectedLabelAliasMap)
	}

	validLabelTypes := model.ValidLabelTypes()
	validLabelCodesByType := model.ValidLabelCodesByType()
	for raw, want := range expectedLabelAliasMap {
		gotLabelType, gotLabelCode, ok := NormalizeLabelAlias(raw)
		if !ok || gotLabelType != want.LabelType || gotLabelCode != want.LabelCode {
			t.Fatalf(
				"NormalizeLabelAlias(%q) = (%q, %q, %t); want (%q, %q, %t)",
				raw,
				gotLabelType,
				gotLabelCode,
				ok,
				want.LabelType,
				want.LabelCode,
				true,
			)
		}
		if _, ok := validLabelTypes[want.LabelType]; !ok {
			t.Fatalf("LabelAliasMap()[%q] uses invalid label type %q", raw, want.LabelType)
		}
		if _, ok := validLabelCodesByType[want.LabelType][want.LabelCode]; !ok {
			t.Fatalf("LabelAliasMap()[%q] uses invalid label code %q for type %q", raw, want.LabelCode, want.LabelType)
		}
	}

	if gotLabelType, gotLabelCode, ok := NormalizeLabelAlias("unknown-label"); ok || gotLabelType != "" || gotLabelCode != "" {
		t.Fatalf("NormalizeLabelAlias(%q) = (%q, %q, %t); want empty miss", "unknown-label", gotLabelType, gotLabelCode, ok)
	}
}

func TestLabelAliasMapReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	got := LabelAliasMap()
	got["colloquial"] = LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFormal}
	got["made-up"] = LabelMapping{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelSlang}

	gotLabelType, gotLabelCode, ok := NormalizeLabelAlias("colloquial")
	if !ok || gotLabelType != model.LabelTypeRegister || gotLabelCode != model.RegisterLabelInformal {
		t.Fatalf(
			"NormalizeLabelAlias(%q) = (%q, %q, %t); want (%q, %q, %t) after caller mutation",
			"colloquial",
			gotLabelType,
			gotLabelCode,
			ok,
			model.LabelTypeRegister,
			model.RegisterLabelInformal,
			true,
		)
	}

	fresh := LabelAliasMap()
	if fresh["colloquial"].LabelCode != model.RegisterLabelInformal {
		t.Fatalf("fresh LabelAliasMap()[%q] = %#v; want informal mapping", "colloquial", fresh["colloquial"])
	}
	if _, ok := fresh["made-up"]; ok {
		t.Fatalf("fresh LabelAliasMap() unexpectedly retained caller-added key %q", "made-up")
	}
}
