package norm

import (
	"maps"

	"github.com/simp-lee/isdict-commons/model"
)

var rawPOSToLearnerPOS = buildRawPOSToLearnerPOS()

// RawPOSToLearnerPOS returns a defensive copy of the frozen raw POS mapping contract.
func RawPOSToLearnerPOS() map[string]string {
	return cloneRawPOSToLearnerPOS(rawPOSToLearnerPOS)
}

func CanonicalizePOS(raw string) string {
	return rawPOSToLearnerPOS[raw]
}

func buildRawPOSToLearnerPOS() map[string]string {
	mappings := make(map[string]string, 52)
	add := func(learnerPOS string, rawValues ...string) {
		for _, rawValue := range rawValues {
			mappings[rawValue] = learnerPOS
		}
	}

	add(model.POSNoun, "noun")
	add(model.POSVerb, "verb")
	add(model.POSAdjective, "adj", "adjective")
	add(model.POSAdverb, "adv", "adverb")
	add(model.POSPronoun, "pron", "pronoun")
	add(model.POSPreposition, "prep", "preposition")
	add(model.POSConjunction, "conj", "conjunction")
	add(model.POSArticle, "article")
	add(model.POSInterjection, "intj", "interjection", "exclamation")
	add(model.POSDeterminer, "det", "determiner")
	add(model.POSNumber, "num", "numeral", "number")
	add(model.POSParticle, "particle")
	add(model.POSPhrasalVerb, "phrasal_verb")
	add(model.POSPhrase, "phrase", "idiom", "prep_phrase", "prepositional_phrase", "adv_phrase")
	add(model.POSAbbreviation, "abbrev", "abbreviation", "initialism", "acronym")
	add(model.POSSymbol, "symbol")
	add(model.POSName, "name", "proper_noun", "proper_name")
	add(model.POSProverb, "proverb")
	add(model.POSCharacter, "character")
	add(model.POSAffix, "prefix", "suffix", "infix", "interfix", "circumfix", "affix", "combining_form")
	add(model.POSContraction, "contraction")
	add(model.POSPunctuation, "punct", "punctuation", "punctuation_mark")
	add(model.POSPostposition, "postp", "postposition")

	return mappings
}

func cloneRawPOSToLearnerPOS(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	maps.Copy(clone, source)

	return clone
}
