package norm

import (
	"maps"
	"reflect"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedLabelAliasMap = map[string]LabelMapping{
	"transitive":                   {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelTransitive},
	"intransitive":                 {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelIntransitive},
	"ditransitive":                 {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelDitransitive},
	"ambitransitive":               {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelAmbitransitive},
	"countable":                    {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelCountable},
	"uncountable":                  {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelUncountable},
	"plural-only":                  {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelPluralOnly},
	"singular-only":                {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelSingularOnly},
	"in the plural":                {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelInThePlural},
	"formal":                       {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFormal},
	"informal":                     {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal},
	"colloquial":                   {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal},
	"slang":                        {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelSlang},
	"literary":                     {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelLiterary},
	"poetic":                       {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelPoetic},
	"vulgar":                       {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelVulgar},
	"taboo":                        {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelTaboo},
	"figurative":                   {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative},
	"figuratively":                 {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative},
	"idiomatic":                    {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelIdiomatic},
	"by extension":                 {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelByExtension},
	"euphemistic":                  {LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelEuphemistic},
	"pejorative":                   {LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelDerogatory},
	"jocular":                      {LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous},
	"facetious":                    {LabelType: model.LabelTypeAttitude, LabelCode: model.AttitudeLabelHumorous},
	"archaic":                      {LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelArchaic},
	"obsolete":                     {LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelObsolete},
	"historical":                   {LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelHistorical},
	"rare":                         {LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelRare},
	"dated":                        {LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelDated},
	"American-English":             {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelUS},
	"British-English":              {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelUK},
	"Australian-English":           {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelAustralia},
	"Canadian-English":             {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelCanada},
	"Irish-English":                {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIreland},
	"Hiberno-English":              {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIreland},
	"Scottish-English":             {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelScotland},
	"Indian-English":               {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelIndia},
	"Singlish":                     {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelSingapore},
	"South-African-English":        {LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelSouthAfrica},
	"plurale-tantum":               {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelPluralOnly},
	"singulare-tantum":             {LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelSingularOnly},
	"African-American Vernacular":  {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelAAVE},
	"AAVE":                         {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelAAVE},
	"Multicultural-London-English": {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelMulticulturalLondonEnglish},
	"MLE":                          {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelMulticulturalLondonEnglish},
	"non-native speakers' English": {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelNonNativeEnglish},
	"non-native-English":           {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelNonNativeEnglish},
	"dialectal":                    {LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelDialectal},
	"law":                          {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelLaw},
	"computing":                    {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelComputing},
	"biology":                      {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelBiology},
	"military":                     {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelMilitary},
	"sports":                       {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelSports},
	"games":                        {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelGames},
	"media":                        {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelMedia},
	"art":                          {LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelArt},
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

func TestNormalizeLabelAliasCanonicalVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		wantType  string
		wantCode  string
		wantMatch bool
	}{
		{name: "case_variant", raw: "COLLOQUIAL", wantType: model.LabelTypeRegister, wantCode: model.RegisterLabelInformal, wantMatch: true},
		{name: "space_variant", raw: "American English", wantType: model.LabelTypeRegion, wantCode: model.RegionLabelUS, wantMatch: true},
		{name: "underscore_variant", raw: "south_african_english", wantType: model.LabelTypeRegion, wantCode: model.RegionLabelSouthAfrica, wantMatch: true},
		{name: "hyphen_variant", raw: "by-extension", wantType: model.LabelTypeRegister, wantCode: model.RegisterLabelByExtension, wantMatch: true},
		{name: "unknown", raw: "colloquialism", wantMatch: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotType, gotCode, gotMatch := NormalizeLabelAlias(tt.raw)
			if gotType != tt.wantType || gotCode != tt.wantCode || gotMatch != tt.wantMatch {
				t.Fatalf(
					"NormalizeLabelAlias(%q) = (%q, %q, %t); want (%q, %q, %t)",
					tt.raw,
					gotType,
					gotCode,
					gotMatch,
					tt.wantType,
					tt.wantCode,
					tt.wantMatch,
				)
			}
		})
	}
}

func TestNormalizeLabelText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []LabelMapping
	}{
		{
			name: "comma_split",
			raw:  "transitive, obsolete",
			want: []LabelMapping{
				{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelTransitive},
				{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelObsolete},
			},
		},
		{
			name: "raw_gloss_registers",
			raw:  "in the plural; by extension; figuratively; colloquial; idiomatic; slang; informal; archaic; historical; rare; dated",
			want: []LabelMapping{
				{LabelType: model.LabelTypeGrammar, LabelCode: model.GrammarLabelInThePlural},
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelByExtension},
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelFigurative},
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal},
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelIdiomatic},
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelSlang},
				{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelArchaic},
				{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelHistorical},
				{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelRare},
				{LabelType: model.LabelTypeTemporal, LabelCode: model.TemporalLabelDated},
			},
		},
		{
			name: "variety_phrases",
			raw:  "chiefly African-American Vernacular; including African-American Vernacular; Multicultural_London_English; non-native speakers' English; dialectal",
			want: []LabelMapping{
				{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelAAVE},
				{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelMulticulturalLondonEnglish},
				{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelNonNativeEnglish},
				{LabelType: model.LabelTypeVariety, LabelCode: model.VarietyLabelDialectal},
			},
		},
		{
			name: "scots_law",
			raw:  "Scots law",
			want: []LabelMapping{
				{LabelType: model.LabelTypeRegion, LabelCode: model.RegionLabelScotland},
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelLaw},
			},
		},
		{
			name: "computing_and_biology_phrases",
			raw:  "object-oriented programming; Java programming language; machine learning; artificial intelligence; marine biology",
			want: []LabelMapping{
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelComputing},
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelBiology},
			},
		},
		{
			name: "military_phrase",
			raw:  "Royal Navy; RAF World War II code name; procedure word",
			want: []LabelMapping{{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelMilitary}},
		},
		{
			name: "sports_games_media_art",
			raw:  "roller derby; Australian rules football; board games; video games; social media; fine arts; theater lighting",
			want: []LabelMapping{
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelSports},
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelGames},
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelMedia},
				{LabelType: model.LabelTypeDomain, LabelCode: model.DomainLabelArt},
			},
		},
		{
			name: "dedupe_preserves_first_order",
			raw:  "slang, slang; colloquial; informal",
			want: []LabelMapping{
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelSlang},
				{LabelType: model.LabelTypeRegister, LabelCode: model.RegisterLabelInformal},
			},
		},
		{name: "substring_false_positive", raw: "slanguage; object-orientedness; MeowUSA", want: nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NormalizeLabelText(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("NormalizeLabelText(%q) = %#v; want %#v", tt.raw, got, tt.want)
			}
		})
	}
}
