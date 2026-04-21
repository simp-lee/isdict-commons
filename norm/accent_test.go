package norm

import (
	"maps"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedAccentTagMap = map[string]string{
	"unknown":                      model.AccentUnknown,
	"Unknown":                      model.AccentUnknown,
	"Received-Pronunciation":       model.AccentBritish,
	"Received Pronunciation":       model.AccentBritish,
	"RP":                           model.AccentBritish,
	"Conservative-RP":              model.AccentBritish,
	"Conservative RP":              model.AccentBritish,
	"UK":                           model.AccentBritish,
	"GB":                           model.AccentBritish,
	"England":                      model.AccentBritish,
	"Standard-Southern-British":    model.AccentBritish,
	"General-American":             model.AccentAmerican,
	"General American":             model.AccentAmerican,
	"GA":                           model.AccentAmerican,
	"GenAm":                        model.AccentAmerican,
	"US":                           model.AccentAmerican,
	"USA":                          model.AccentAmerican,
	"U.S.":                         model.AccentAmerican,
	"U.S.A.":                       model.AccentAmerican,
	"United-States":                model.AccentAmerican,
	"North-American":               model.AccentAmerican,
	"General-Australian":           model.AccentAustralian,
	"General Australian":           model.AccentAustralian,
	"Australia":                    model.AccentAustralian,
	"Canada":                       model.AccentCanadian,
	"Standard-Canadian":            model.AccentCanadian,
	"Ireland":                      model.AccentIrish,
	"Irish":                        model.AccentIrish,
	"Hiberno-English":              model.AccentIrish,
	"Ulster":                       model.AccentIrish,
	"Munster":                      model.AccentIrish,
	"Scotland":                     model.AccentScottish,
	"Scottish":                     model.AccentScottish,
	"Scots":                        model.AccentScottish,
	"Glaswegian":                   model.AccentScottish,
	"Edinburgh":                    model.AccentScottish,
	"New-Zealand":                  model.AccentNZ,
	"NZ":                           model.AccentNZ,
	"India":                        model.AccentIndian,
	"Indian-English":               model.AccentIndian,
	"South-Asian":                  model.AccentIndian,
	"Pakistan":                     model.AccentIndian,
	"South-Africa":                 model.AccentSouthAfrican,
	"South-African":                model.AccentSouthAfrican,
	"South-African-English":        model.AccentSouthAfrican,
	"Jamaica":                      model.AccentOtherRegional,
	"Northern-Ireland":             model.AccentOtherRegional,
	"Northern-England":             model.AccentOtherRegional,
	"Southern-England":             model.AccentOtherRegional,
	"Multicultural-London-English": model.AccentOtherRegional,
	"Cockney":                      model.AccentOtherRegional,
	"Estuary-English":              model.AccentOtherRegional,
	"Geordie":                      model.AccentOtherRegional,
	"Scouse":                       model.AccentOtherRegional,
	"Yorkshire":                    model.AccentOtherRegional,
	"Lancashire":                   model.AccentOtherRegional,
	"West-Country":                 model.AccentOtherRegional,
	"East-Anglia":                  model.AccentOtherRegional,
	"Norfolk":                      model.AccentOtherRegional,
	"Essex":                        model.AccentOtherRegional,
	"London":                       model.AccentOtherRegional,
	"Cornwall":                     model.AccentOtherRegional,
	"New-York":                     model.AccentOtherRegional,
	"Boston":                       model.AccentOtherRegional,
	"Philadelphia":                 model.AccentOtherRegional,
	"California":                   model.AccentOtherRegional,
	"Texas":                        model.AccentOtherRegional,
	"Southern-US":                  model.AccentOtherRegional,
	"Appalachian":                  model.AccentOtherRegional,
	"New-England":                  model.AccentOtherRegional,
	"Mid-Atlantic":                 model.AccentOtherRegional,
	"Midwest":                      model.AccentOtherRegional,
	"Singapore":                    model.AccentOtherRegional,
	"Singlish":                     model.AccentOtherRegional,
	"Philippines":                  model.AccentOtherRegional,
	"Hong-Kong":                    model.AccentOtherRegional,
	"Malaysia":                     model.AccentOtherRegional,
	"Caribbean":                    model.AccentOtherRegional,
	"Nigeria":                      model.AccentOtherRegional,
	"Kenya":                        model.AccentOtherRegional,
	"Ghana":                        model.AccentOtherRegional,
	"Welsh":                        model.AccentOtherRegional,
	"Northumbria":                  model.AccentOtherRegional,
}

func TestAccentTagMapContract(t *testing.T) {
	t.Parallel()

	got := AccentTagMap()
	if !maps.Equal(got, expectedAccentTagMap) {
		t.Fatalf("AccentTagMap() mismatch: got %#v; want %#v", got, expectedAccentTagMap)
	}

	validAccentCodes := model.AccentCodeToName()
	for rawTag, wantAccent := range expectedAccentTagMap {
		if gotAccent := NormalizeAccentCode(rawTag); gotAccent != wantAccent {
			t.Fatalf("NormalizeAccentCode(%q) = %q; want %q", rawTag, gotAccent, wantAccent)
		}
		if _, ok := validAccentCodes[wantAccent]; !ok {
			t.Fatalf("AccentTagMap()[%q] = %q is not a valid controlled accent code", rawTag, wantAccent)
		}
	}
}

func TestNormalizeAccentCodeTokenFallbacksAndDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rawTag string
		want   string
	}{
		{name: "token_standard_southern_british_spacing_variant", rawTag: "standard southern british", want: model.AccentBritish},
		{name: "token_north_american_spacing_variant", rawTag: "north american", want: model.AccentAmerican},
		{name: "token_standard_canadian_spacing_variant", rawTag: "standard canadian", want: model.AccentCanadian},
		{name: "token_new_zealand_spacing_variant", rawTag: "new zealand", want: model.AccentNZ},
		{name: "token_south_african_spacing_variant", rawTag: "south african", want: model.AccentSouthAfrican},
		{name: "regression_northern_ireland_hyphenated", rawTag: "Northern-Ireland", want: model.AccentOtherRegional},
		{name: "regression_northern_ireland_spacing_variant", rawTag: "Northern Ireland", want: model.AccentOtherRegional},
		{name: "blank_tag", rawTag: "", want: model.AccentUnknown},
		{name: "whitespace_only_tag", rawTag: " \t\n ", want: model.AccentUnknown},
		{name: "unknown_tag", rawTag: "Atlantis", want: model.AccentOtherRegional},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if gotAccent := NormalizeAccentCode(tt.rawTag); gotAccent != tt.want {
				t.Fatalf("NormalizeAccentCode(%q) = %q; want %q", tt.rawTag, gotAccent, tt.want)
			}
		})
	}
}

func TestAccentTagMapReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	got := AccentTagMap()
	got["Received-Pronunciation"] = model.AccentAmerican
	got["made-up"] = model.AccentBritish

	if gotAccent := NormalizeAccentCode("Received-Pronunciation"); gotAccent != model.AccentBritish {
		t.Fatalf("NormalizeAccentCode(%q) = %q; want %q after caller mutation", "Received-Pronunciation", gotAccent, model.AccentBritish)
	}

	fresh := AccentTagMap()
	if fresh["Received-Pronunciation"] != model.AccentBritish {
		t.Fatalf("fresh AccentTagMap()[%q] = %q; want %q", "Received-Pronunciation", fresh["Received-Pronunciation"], model.AccentBritish)
	}
	if _, ok := fresh["made-up"]; ok {
		t.Fatalf("fresh AccentTagMap() unexpectedly retained caller-added key %q", "made-up")
	}
}
