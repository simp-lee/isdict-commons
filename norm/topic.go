package norm

import (
	"maps"

	"github.com/simp-lee/isdict-commons/model"
)

var topicToDomainMap = buildTopicToDomainMap()
var normalizedTopicToDomainMap = buildNormalizedTopicToDomainMap(topicToDomainMap)

// TopicToDomainMap returns a defensive copy of the frozen topic-to-domain mapping contract.
func TopicToDomainMap() map[string]string {
	return cloneTopicToDomainMap(topicToDomainMap)
}

var droppedTopics = buildDroppedTopics()
var normalizedDroppedTopics = buildNormalizedDroppedTopics(droppedTopics)

func NormalizeTopic(topic string) (labelCode string, ok bool) {
	labelCode, ok = topicToDomainMap[topic]
	if ok {
		return labelCode, true
	}
	if _, dropped := droppedTopics[topic]; dropped {
		return "", false
	}

	normalized := canonicalLookupText(topic)
	if normalized == "" {
		return "", false
	}
	labelCode, ok = normalizedTopicToDomainMap[normalized]
	if ok {
		return labelCode, true
	}
	if _, dropped := normalizedDroppedTopics[normalized]; dropped {
		return "", false
	}

	return "", false
}

func buildTopicToDomainMap() map[string]string {
	mappings := make(map[string]string, 200)
	add := func(domain string, topics ...string) {
		for _, topic := range topics {
			mappings[topic] = domain
		}
	}

	add(model.DomainLabelMedicine, "medicine", "pathology", "pharmacology", "anatomy", "surgery", "physiology")
	add(model.DomainLabelLaw, "law", "legal", "law-enforcement")
	add(model.DomainLabelComputing, "computing", "programming", "computer-science", "software", "Internet", "object-oriented-programming", "java-programming-language", "machine-learning", "artificial-intelligence", "graphical-user-interface", "information-technology")
	add(model.DomainLabelFinance, "finance", "banking", "economics", "accounting", "stock-market", "insurance")
	add(model.DomainLabelBusiness, "business", "marketing")
	add(model.DomainLabelMusic, "music", "musical-instruments")
	add(model.DomainLabelSports, "sports", "football", "cricket", "baseball", "tennis", "basketball", "ball-games", "martial-arts", "soccer", "roller-derby", "Australian-rules-football", "croquet", "field-sports", "combat-sports", "bobsledding", "ultimate-frisbee", "American-football", "golf", "racing", "horse-racing", "horseracing", "rugby", "bowling", "snooker", "climbing", "hunting")
	add(model.DomainLabelBiology, "biology", "ecology", "genetics", "microbiology", "marine-biology")
	add(model.DomainLabelChemistry, "chemistry", "organic-chemistry", "organic chemistry", "inorganic-chemistry", "inorganic chemistry", "analytical-chemistry", "analytical chemistry", "biochemistry")
	add(model.DomainLabelPhysics, "physics", "optics", "quantum-mechanics", "quantum mechanics", "nuclear-physics", "nuclear physics", "thermodynamics", "energy", "electricity", "electromagnetism")
	add(model.DomainLabelEngineering, "engineering", "civil-engineering", "electrical-engineering", "manufacturing", "technology")
	add(model.DomainLabelMathematics, "mathematics", "geometry", "differential-geometry", "differential geometry", "algebra", "statistics")
	add(model.DomainLabelBotany, "botany", "plants", "mycology", "agriculture", "horticulture")
	add(model.DomainLabelZoology, "zoology", "animals", "ornithology", "entomology", "ichthyology", "horses", "pets")
	add(model.DomainLabelLinguistics, "linguistics", "grammar", "phonetics", "phonology")
	add(model.DomainLabelMilitary, "military", "weaponry", "army", "navy", "war")
	add(model.DomainLabelArchitecture, "architecture", "construction", "building")
	add(model.DomainLabelReligion, "religion", "Christianity", "Islam", "Buddhism", "theology", "mysticism", "biblical", "Catholicism")
	add(model.DomainLabelPolitics, "politics", "government", "diplomacy", "monarchy")
	add(model.DomainLabelCooking, "cooking", "food", "beverages", "gastronomy")
	add(model.DomainLabelNautical, "nautical", "seafaring", "marine")
	add(model.DomainLabelAstronomy, "astronomy", "astrophysics")
	add(model.DomainLabelGeology, "geology", "planetary-geology", "planetary geology", "mineralogy")
	add(model.DomainLabelAviation, "aviation", "aeronautics", "aerospace")
	add(model.DomainLabelElectronics, "electronics", "computer-keyboards", "computer keyboards", "semiconductors")
	add(model.DomainLabelPsychology, "psychology")
	add(model.DomainLabelPhilosophy, "philosophy")
	add(model.DomainLabelGames, "games", "board-games", "video-games", "roleplaying-games", "puzzles", "chess", "card-games", "poker")
	add(model.DomainLabelMedia, "media", "publishing", "film", "television", "video-editing", "video editing", "sound-recording", "sound recording", "sound-engineering", "sound engineering", "broadcasting", "journalism", "communications")
	add(model.DomainLabelEducation, "education", "academia")
	add(model.DomainLabelTransport, "transport", "rail-transport", "rail transport", "railways", "vehicles")
	add(model.DomainLabelAutomotive, "automotive", "automobiles", "cars")
	add(model.DomainLabelPrinting, "printing", "typography")
	add(model.DomainLabelMining, "mining")
	add(model.DomainLabelMeteorology, "meteorology", "weather", "climatology")
	add(model.DomainLabelHeraldry, "heraldry")
	add(model.DomainLabelMaterialsScience, "materials-science", "materials science")
	add(model.DomainLabelMythology, "mythology")
	add(model.DomainLabelArt, "art", "arts", "fine-arts", "painting", "theater")

	return mappings
}

func buildDroppedTopics() map[string]struct{} {
	dropped := make(map[string]struct{}, 11)
	add := func(topics ...string) {
		for _, topic := range topics {
			dropped[topic] = struct{}{}
		}
	}

	add("natural-sciences", "sciences", "physical-sciences", "human-sciences")
	add("lifestyle", "hobbies", "entertainment", "literature", "science-fiction", "tools", "sexuality", "geography")

	return dropped
}

func buildNormalizedTopicToDomainMap(source map[string]string) map[string]string {
	mappings := make(map[string]string, len(source))
	for topic, domain := range source {
		if normalized := canonicalLookupText(topic); normalized != "" {
			mappings[normalized] = domain
		}
	}

	return mappings
}

func buildNormalizedDroppedTopics(source map[string]struct{}) map[string]struct{} {
	dropped := make(map[string]struct{}, len(source))
	for topic := range source {
		if normalized := canonicalLookupText(topic); normalized != "" {
			dropped[normalized] = struct{}{}
		}
	}

	return dropped
}

func cloneTopicToDomainMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	maps.Copy(clone, source)

	return clone
}
