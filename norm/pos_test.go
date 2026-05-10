package norm

import (
	"maps"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedRawPOSToLearnerPOS = map[string]string{
	"n":                    model.POSNoun,
	"noun":                 model.POSNoun,
	"v":                    model.POSVerb,
	"vt":                   model.POSVerb,
	"vi":                   model.POSVerb,
	"aux":                  model.POSVerb,
	"modalv":               model.POSVerb,
	"verb":                 model.POSVerb,
	"a":                    model.POSAdjective,
	"s":                    model.POSAdjective,
	"adj":                  model.POSAdjective,
	"adjective":            model.POSAdjective,
	"r":                    model.POSAdverb,
	"adv":                  model.POSAdverb,
	"adverb":               model.POSAdverb,
	"neg":                  model.POSAdverb,
	"pron":                 model.POSPronoun,
	"pronoun":              model.POSPronoun,
	"prep":                 model.POSPreposition,
	"preposition":          model.POSPreposition,
	"conj":                 model.POSConjunction,
	"conjunction":          model.POSConjunction,
	"art":                  model.POSArticle,
	"article":              model.POSArticle,
	"int":                  model.POSInterjection,
	"intj":                 model.POSInterjection,
	"interjection":         model.POSInterjection,
	"exclamation":          model.POSInterjection,
	"det":                  model.POSDeterminer,
	"determiner":           model.POSDeterminer,
	"num":                  model.POSNumber,
	"numeral":              model.POSNumber,
	"number":               model.POSNumber,
	"particle":             model.POSParticle,
	"phrasal_verb":         model.POSPhrasalVerb,
	"ph":                   model.POSPhrase,
	"na":                   model.POSPhrase,
	"un":                   model.POSPhrase,
	"phrase":               model.POSPhrase,
	"idiom":                model.POSPhrase,
	"prep_phrase":          model.POSPhrase,
	"prepositional_phrase": model.POSPhrase,
	"adv_phrase":           model.POSPhrase,
	"abbr":                 model.POSAbbreviation,
	"abbrev":               model.POSAbbreviation,
	"abbreviation":         model.POSAbbreviation,
	"initialism":           model.POSAbbreviation,
	"acronym":              model.POSAbbreviation,
	"symbol":               model.POSSymbol,
	"name":                 model.POSName,
	"proper_noun":          model.POSName,
	"proper_name":          model.POSName,
	"st":                   model.POSProverb,
	"proverb":              model.POSProverb,
	"character":            model.POSCharacter,
	"pref":                 model.POSAffix,
	"prefix":               model.POSAffix,
	"suffix":               model.POSAffix,
	"infix":                model.POSAffix,
	"interfix":             model.POSAffix,
	"circumfix":            model.POSAffix,
	"affix":                model.POSAffix,
	"combining_form":       model.POSAffix,
	"short":                model.POSContraction,
	"contraction":          model.POSContraction,
	"punct":                model.POSPunctuation,
	"punctuation":          model.POSPunctuation,
	"punctuation_mark":     model.POSPunctuation,
	"postp":                model.POSPostposition,
	"postposition":         model.POSPostposition,
}

func TestRawPOSToLearnerPOSContract(t *testing.T) {
	t.Parallel()

	got := RawPOSToLearnerPOS()
	if !maps.Equal(got, expectedRawPOSToLearnerPOS) {
		t.Fatalf("RawPOSToLearnerPOS() mismatch: got %#v; want %#v", got, expectedRawPOSToLearnerPOS)
	}

	validPOSCodes := model.ValidPOSCodes()
	for raw, want := range expectedRawPOSToLearnerPOS {
		if gotCode := CanonicalizePOS(raw); gotCode != want {
			t.Fatalf("CanonicalizePOS(%q) = %q; want %q", raw, gotCode, want)
		}
		if _, ok := validPOSCodes[want]; !ok {
			t.Fatalf("RawPOSToLearnerPOS()[%q] = %q is not a valid learner POS", raw, want)
		}
	}

	if gotCode := CanonicalizePOS("unknown_pos"); gotCode != "" {
		t.Fatalf("CanonicalizePOS(%q) = %q; want empty string", "unknown_pos", gotCode)
	}
}

func TestCanonicalizePOSNormalizesSchoolRawMarkers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want string
	}{
		{raw: " n. ", want: model.POSNoun},
		{raw: "VT.", want: model.POSVerb},
		{raw: "vi", want: model.POSVerb},
		{raw: "aux.", want: model.POSVerb},
		{raw: "modalv.", want: model.POSVerb},
		{raw: "a.", want: model.POSAdjective},
		{raw: "abbr.", want: model.POSAbbreviation},
		{raw: "art.", want: model.POSArticle},
		{raw: "int.", want: model.POSInterjection},
		{raw: "neg.", want: model.POSAdverb},
		{raw: "na.", want: model.POSPhrase},
		{raw: "un.", want: model.POSPhrase},
		{raw: "ph.", want: model.POSPhrase},
		{raw: "pref.", want: model.POSAffix},
		{raw: "short.", want: model.POSContraction},
		{raw: "st.", want: model.POSProverb},
		{raw: "[n", want: model.POSNoun},
		{raw: "[n]", want: model.POSNoun},
		{raw: "[n].", want: model.POSNoun},
		{raw: "[ n ].", want: model.POSNoun},
		{raw: "(adj.)", want: model.POSAdjective},
		{raw: "[化学]", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()
			if got := CanonicalizePOS(tt.raw); got != tt.want {
				t.Fatalf("CanonicalizePOS(%q) = %q; want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestRawPOSToLearnerPOSReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	got := RawPOSToLearnerPOS()
	got["adj"] = model.POSNoun
	got["made-up"] = model.POSVerb

	if gotCode := CanonicalizePOS("adj"); gotCode != model.POSAdjective {
		t.Fatalf("CanonicalizePOS(%q) = %q; want %q after caller mutation", "adj", gotCode, model.POSAdjective)
	}

	fresh := RawPOSToLearnerPOS()
	if fresh["adj"] != model.POSAdjective {
		t.Fatalf("fresh RawPOSToLearnerPOS()[%q] = %q; want %q", "adj", fresh["adj"], model.POSAdjective)
	}
	if _, ok := fresh["made-up"]; ok {
		t.Fatalf("fresh RawPOSToLearnerPOS() unexpectedly retained caller-added key %q", "made-up")
	}
}
