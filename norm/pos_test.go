package norm

import (
	"maps"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedRawPOSToLearnerPOS = map[string]string{
	"noun":                 model.POSNoun,
	"verb":                 model.POSVerb,
	"adj":                  model.POSAdjective,
	"adjective":            model.POSAdjective,
	"adv":                  model.POSAdverb,
	"adverb":               model.POSAdverb,
	"pron":                 model.POSPronoun,
	"pronoun":              model.POSPronoun,
	"prep":                 model.POSPreposition,
	"preposition":          model.POSPreposition,
	"conj":                 model.POSConjunction,
	"conjunction":          model.POSConjunction,
	"article":              model.POSArticle,
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
	"phrase":               model.POSPhrase,
	"idiom":                model.POSPhrase,
	"prep_phrase":          model.POSPhrase,
	"prepositional_phrase": model.POSPhrase,
	"adv_phrase":           model.POSPhrase,
	"abbrev":               model.POSAbbreviation,
	"abbreviation":         model.POSAbbreviation,
	"initialism":           model.POSAbbreviation,
	"acronym":              model.POSAbbreviation,
	"symbol":               model.POSSymbol,
	"name":                 model.POSName,
	"proper_noun":          model.POSName,
	"proper_name":          model.POSName,
	"proverb":              model.POSProverb,
	"character":            model.POSCharacter,
	"prefix":               model.POSAffix,
	"suffix":               model.POSAffix,
	"infix":                model.POSAffix,
	"interfix":             model.POSAffix,
	"circumfix":            model.POSAffix,
	"affix":                model.POSAffix,
	"combining_form":       model.POSAffix,
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
