package norm

import (
	"maps"
	"strings"

	"github.com/simp-lee/isdict-commons/model"
)

var rawPOSToLearnerPOS = buildRawPOSToLearnerPOS()

// RawPOSToLearnerPOS returns a defensive copy of the frozen raw POS mapping contract.
func RawPOSToLearnerPOS() map[string]string {
	return cloneRawPOSToLearnerPOS(rawPOSToLearnerPOS)
}

func CanonicalizePOS(raw string) string {
	return rawPOSToLearnerPOS[normalizeRawPOSLookupKey(raw)]
}

func normalizeRawPOSLookupKey(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.TrimRight(normalized, ".")
	normalized = strings.Trim(normalized, "[]()")
	normalized = strings.TrimSpace(normalized)
	normalized = strings.TrimRight(normalized, ".")
	return normalized
}

func buildRawPOSToLearnerPOS() map[string]string {
	mappings := make(map[string]string, 75)
	add := func(learnerPOS string, rawValues ...string) {
		for _, rawValue := range rawValues {
			mappings[rawValue] = learnerPOS
		}
	}

	add(model.POSNoun, "n", "noun")
	add(model.POSVerb, "v", "vt", "vi", "aux", "modalv", "verb")
	add(model.POSAdjective, "a", "s", "adj", "adjective")
	add(model.POSAdverb, "r", "adv", "adverb", "neg")
	add(model.POSPronoun, "pron", "pronoun")
	add(model.POSPreposition, "prep", "preposition")
	add(model.POSConjunction, "conj", "conjunction")
	add(model.POSArticle, "art", "article")
	add(model.POSInterjection, "int", "intj", "interjection", "exclamation")
	add(model.POSDeterminer, "det", "determiner")
	add(model.POSNumber, "num", "numeral", "number")
	add(model.POSParticle, "particle")
	add(model.POSPhrasalVerb, "phrasal_verb")
	add(model.POSPhrase, "ph", "na", "un", "phrase", "idiom", "prep_phrase", "prepositional_phrase", "adv_phrase")
	add(model.POSAbbreviation, "abbr", "abbrev", "abbreviation", "initialism", "acronym")
	add(model.POSSymbol, "symbol")
	add(model.POSName, "name", "proper_noun", "proper_name")
	add(model.POSProverb, "st", "proverb")
	add(model.POSCharacter, "character")
	add(model.POSAffix, "pref", "prefix", "suffix", "infix", "interfix", "circumfix", "affix", "combining_form")
	add(model.POSContraction, "short", "contraction")
	add(model.POSPunctuation, "punct", "punctuation", "punctuation_mark")
	add(model.POSPostposition, "postp", "postposition")

	return mappings
}

func cloneRawPOSToLearnerPOS(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	maps.Copy(clone, source)

	return clone
}
