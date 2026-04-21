package norm

import (
	"maps"

	"github.com/simp-lee/isdict-commons/model"
)

var topicToDomainMap = buildTopicToDomainMap()

// TopicToDomainMap returns a defensive copy of the frozen topic-to-domain mapping contract.
func TopicToDomainMap() map[string]string {
	return cloneTopicToDomainMap(topicToDomainMap)
}

var droppedTopics = buildDroppedTopics()

func NormalizeTopic(topic string) (labelCode string, ok bool) {
	labelCode, ok = topicToDomainMap[topic]
	if ok {
		return labelCode, true
	}

	if _, dropped := droppedTopics[topic]; dropped {
		return "", false
	}

	return "", false
}

func buildTopicToDomainMap() map[string]string {
	mappings := make(map[string]string, 87)
	add := func(domain string, topics ...string) {
		for _, topic := range topics {
			mappings[topic] = domain
		}
	}

	add(model.DomainLabelMedicine, "medicine", "pathology", "pharmacology", "anatomy", "surgery")
	add(model.DomainLabelLaw, "law", "legal")
	add(model.DomainLabelComputing, "computing", "programming", "computer-science", "software", "Internet")
	add(model.DomainLabelFinance, "finance", "banking", "economics", "accounting", "stock-market", "insurance")
	add(model.DomainLabelBusiness, "business", "marketing")
	add(model.DomainLabelMusic, "music", "musical-instruments")
	add(model.DomainLabelSports, "sports", "football", "cricket", "baseball", "tennis", "basketball", "ball-games", "martial-arts", "soccer")
	add(model.DomainLabelBiology, "biology", "ecology", "genetics", "microbiology")
	add(model.DomainLabelChemistry, "chemistry", "organic-chemistry", "biochemistry")
	add(model.DomainLabelPhysics, "physics", "optics", "quantum-mechanics", "thermodynamics", "energy", "electricity")
	add(model.DomainLabelEngineering, "engineering", "civil-engineering", "electrical-engineering")
	add(model.DomainLabelMathematics, "mathematics", "geometry", "algebra", "statistics")
	add(model.DomainLabelBotany, "botany", "plants", "mycology")
	add(model.DomainLabelZoology, "zoology", "animals", "ornithology", "entomology", "ichthyology")
	add(model.DomainLabelLinguistics, "linguistics", "grammar", "phonetics", "phonology")
	add(model.DomainLabelMilitary, "military", "weaponry", "army", "navy", "war")
	add(model.DomainLabelArchitecture, "architecture", "construction", "building")
	add(model.DomainLabelReligion, "religion", "Christianity", "Islam", "Buddhism", "theology", "mysticism", "biblical", "Catholicism")
	add(model.DomainLabelPolitics, "politics", "government", "diplomacy", "monarchy")
	add(model.DomainLabelCooking, "cooking", "food", "beverages", "gastronomy")

	return mappings
}

func buildDroppedTopics() map[string]struct{} {
	dropped := make(map[string]struct{}, 18)
	add := func(topics ...string) {
		for _, topic := range topics {
			dropped[topic] = struct{}{}
		}
	}

	add("natural-sciences", "sciences", "physical-sciences", "human-sciences")
	add("lifestyle", "hobbies", "games", "entertainment", "media", "publishing", "literature", "arts", "film", "television", "science-fiction", "tools", "sexuality", "geography")

	return dropped
}

func cloneTopicToDomainMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	maps.Copy(clone, source)

	return clone
}
