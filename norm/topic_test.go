package norm

import (
	"maps"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedTopicToDomainMap = map[string]string{
	"medicine":               model.DomainLabelMedicine,
	"pathology":              model.DomainLabelMedicine,
	"pharmacology":           model.DomainLabelMedicine,
	"anatomy":                model.DomainLabelMedicine,
	"surgery":                model.DomainLabelMedicine,
	"law":                    model.DomainLabelLaw,
	"legal":                  model.DomainLabelLaw,
	"computing":              model.DomainLabelComputing,
	"programming":            model.DomainLabelComputing,
	"computer-science":       model.DomainLabelComputing,
	"software":               model.DomainLabelComputing,
	"Internet":               model.DomainLabelComputing,
	"finance":                model.DomainLabelFinance,
	"banking":                model.DomainLabelFinance,
	"economics":              model.DomainLabelFinance,
	"accounting":             model.DomainLabelFinance,
	"stock-market":           model.DomainLabelFinance,
	"insurance":              model.DomainLabelFinance,
	"business":               model.DomainLabelBusiness,
	"marketing":              model.DomainLabelBusiness,
	"music":                  model.DomainLabelMusic,
	"musical-instruments":    model.DomainLabelMusic,
	"sports":                 model.DomainLabelSports,
	"football":               model.DomainLabelSports,
	"cricket":                model.DomainLabelSports,
	"baseball":               model.DomainLabelSports,
	"tennis":                 model.DomainLabelSports,
	"basketball":             model.DomainLabelSports,
	"ball-games":             model.DomainLabelSports,
	"martial-arts":           model.DomainLabelSports,
	"soccer":                 model.DomainLabelSports,
	"biology":                model.DomainLabelBiology,
	"ecology":                model.DomainLabelBiology,
	"genetics":               model.DomainLabelBiology,
	"microbiology":           model.DomainLabelBiology,
	"chemistry":              model.DomainLabelChemistry,
	"organic-chemistry":      model.DomainLabelChemistry,
	"biochemistry":           model.DomainLabelChemistry,
	"physics":                model.DomainLabelPhysics,
	"optics":                 model.DomainLabelPhysics,
	"quantum-mechanics":      model.DomainLabelPhysics,
	"thermodynamics":         model.DomainLabelPhysics,
	"energy":                 model.DomainLabelPhysics,
	"electricity":            model.DomainLabelPhysics,
	"engineering":            model.DomainLabelEngineering,
	"civil-engineering":      model.DomainLabelEngineering,
	"electrical-engineering": model.DomainLabelEngineering,
	"mathematics":            model.DomainLabelMathematics,
	"geometry":               model.DomainLabelMathematics,
	"algebra":                model.DomainLabelMathematics,
	"statistics":             model.DomainLabelMathematics,
	"botany":                 model.DomainLabelBotany,
	"plants":                 model.DomainLabelBotany,
	"mycology":               model.DomainLabelBotany,
	"zoology":                model.DomainLabelZoology,
	"animals":                model.DomainLabelZoology,
	"ornithology":            model.DomainLabelZoology,
	"entomology":             model.DomainLabelZoology,
	"ichthyology":            model.DomainLabelZoology,
	"linguistics":            model.DomainLabelLinguistics,
	"grammar":                model.DomainLabelLinguistics,
	"phonetics":              model.DomainLabelLinguistics,
	"phonology":              model.DomainLabelLinguistics,
	"military":               model.DomainLabelMilitary,
	"weaponry":               model.DomainLabelMilitary,
	"army":                   model.DomainLabelMilitary,
	"navy":                   model.DomainLabelMilitary,
	"war":                    model.DomainLabelMilitary,
	"architecture":           model.DomainLabelArchitecture,
	"construction":           model.DomainLabelArchitecture,
	"building":               model.DomainLabelArchitecture,
	"religion":               model.DomainLabelReligion,
	"Christianity":           model.DomainLabelReligion,
	"Islam":                  model.DomainLabelReligion,
	"Buddhism":               model.DomainLabelReligion,
	"theology":               model.DomainLabelReligion,
	"mysticism":              model.DomainLabelReligion,
	"biblical":               model.DomainLabelReligion,
	"Catholicism":            model.DomainLabelReligion,
	"politics":               model.DomainLabelPolitics,
	"government":             model.DomainLabelPolitics,
	"diplomacy":              model.DomainLabelPolitics,
	"monarchy":               model.DomainLabelPolitics,
	"cooking":                model.DomainLabelCooking,
	"food":                   model.DomainLabelCooking,
	"beverages":              model.DomainLabelCooking,
	"gastronomy":             model.DomainLabelCooking,
}

var expectedDroppedTopics = map[string]struct{}{
	"natural-sciences":  {},
	"sciences":          {},
	"physical-sciences": {},
	"human-sciences":    {},
	"lifestyle":         {},
	"hobbies":           {},
	"games":             {},
	"entertainment":     {},
	"media":             {},
	"publishing":        {},
	"literature":        {},
	"arts":              {},
	"film":              {},
	"television":        {},
	"science-fiction":   {},
	"tools":             {},
	"sexuality":         {},
	"geography":         {},
}

func TestTopicToDomainMapContract(t *testing.T) {
	t.Parallel()

	got := TopicToDomainMap()
	if !maps.Equal(got, expectedTopicToDomainMap) {
		t.Fatalf("TopicToDomainMap() mismatch: got %#v; want %#v", got, expectedTopicToDomainMap)
	}
	if !maps.Equal(droppedTopics, expectedDroppedTopics) {
		t.Fatalf("droppedTopics mismatch: got %#v; want %#v", droppedTopics, expectedDroppedTopics)
	}

	domainCodes := model.ValidLabelCodesByType()[model.LabelTypeDomain]
	for topic, wantDomain := range expectedTopicToDomainMap {
		gotDomain, ok := NormalizeTopic(topic)
		if !ok || gotDomain != wantDomain {
			t.Fatalf("NormalizeTopic(%q) = (%q, %t); want (%q, %t)", topic, gotDomain, ok, wantDomain, true)
		}
		if _, ok := domainCodes[wantDomain]; !ok {
			t.Fatalf("TopicToDomainMap()[%q] = %q is not a valid domain label", topic, wantDomain)
		}
	}

	for topic := range expectedDroppedTopics {
		gotDomain, ok := NormalizeTopic(topic)
		if ok || gotDomain != "" {
			t.Fatalf("NormalizeTopic(%q) = (%q, %t); want no label", topic, gotDomain, ok)
		}
		if _, ok := expectedTopicToDomainMap[topic]; ok {
			t.Fatalf("topic %q unexpectedly appears in both retained and dropped contracts", topic)
		}
	}

	if gotDomain, ok := NormalizeTopic("unknown-topic"); ok || gotDomain != "" {
		t.Fatalf("NormalizeTopic(%q) = (%q, %t); want no label", "unknown-topic", gotDomain, ok)
	}
}

func TestTopicToDomainMapReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	got := TopicToDomainMap()
	got["marketing"] = model.DomainLabelFinance
	got["made-up"] = model.DomainLabelBusiness

	gotDomain, ok := NormalizeTopic("marketing")
	if !ok || gotDomain != model.DomainLabelBusiness {
		t.Fatalf("NormalizeTopic(%q) = (%q, %t); want (%q, %t) after caller mutation", "marketing", gotDomain, ok, model.DomainLabelBusiness, true)
	}

	fresh := TopicToDomainMap()
	if fresh["marketing"] != model.DomainLabelBusiness {
		t.Fatalf("fresh TopicToDomainMap()[%q] = %q; want %q", "marketing", fresh["marketing"], model.DomainLabelBusiness)
	}
	if _, ok := fresh["made-up"]; ok {
		t.Fatalf("fresh TopicToDomainMap() unexpectedly retained caller-added key %q", "made-up")
	}
}
