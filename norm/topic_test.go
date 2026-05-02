package norm

import (
	"maps"
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

var expectedTopicToDomainMap = map[string]string{
	"medicine":                    model.DomainLabelMedicine,
	"pathology":                   model.DomainLabelMedicine,
	"pharmacology":                model.DomainLabelMedicine,
	"anatomy":                     model.DomainLabelMedicine,
	"surgery":                     model.DomainLabelMedicine,
	"physiology":                  model.DomainLabelMedicine,
	"law":                         model.DomainLabelLaw,
	"legal":                       model.DomainLabelLaw,
	"law-enforcement":             model.DomainLabelLaw,
	"computing":                   model.DomainLabelComputing,
	"programming":                 model.DomainLabelComputing,
	"computer-science":            model.DomainLabelComputing,
	"software":                    model.DomainLabelComputing,
	"Internet":                    model.DomainLabelComputing,
	"object-oriented-programming": model.DomainLabelComputing,
	"java-programming-language":   model.DomainLabelComputing,
	"machine-learning":            model.DomainLabelComputing,
	"artificial-intelligence":     model.DomainLabelComputing,
	"graphical-user-interface":    model.DomainLabelComputing,
	"information-technology":      model.DomainLabelComputing,
	"finance":                     model.DomainLabelFinance,
	"banking":                     model.DomainLabelFinance,
	"economics":                   model.DomainLabelFinance,
	"accounting":                  model.DomainLabelFinance,
	"stock-market":                model.DomainLabelFinance,
	"insurance":                   model.DomainLabelFinance,
	"business":                    model.DomainLabelBusiness,
	"marketing":                   model.DomainLabelBusiness,
	"music":                       model.DomainLabelMusic,
	"musical-instruments":         model.DomainLabelMusic,
	"sports":                      model.DomainLabelSports,
	"football":                    model.DomainLabelSports,
	"cricket":                     model.DomainLabelSports,
	"baseball":                    model.DomainLabelSports,
	"tennis":                      model.DomainLabelSports,
	"basketball":                  model.DomainLabelSports,
	"ball-games":                  model.DomainLabelSports,
	"martial-arts":                model.DomainLabelSports,
	"soccer":                      model.DomainLabelSports,
	"roller-derby":                model.DomainLabelSports,
	"Australian-rules-football":   model.DomainLabelSports,
	"croquet":                     model.DomainLabelSports,
	"field-sports":                model.DomainLabelSports,
	"combat-sports":               model.DomainLabelSports,
	"bobsledding":                 model.DomainLabelSports,
	"ultimate-frisbee":            model.DomainLabelSports,
	"American-football":           model.DomainLabelSports,
	"golf":                        model.DomainLabelSports,
	"racing":                      model.DomainLabelSports,
	"horse-racing":                model.DomainLabelSports,
	"horseracing":                 model.DomainLabelSports,
	"rugby":                       model.DomainLabelSports,
	"bowling":                     model.DomainLabelSports,
	"snooker":                     model.DomainLabelSports,
	"climbing":                    model.DomainLabelSports,
	"hunting":                     model.DomainLabelSports,
	"biology":                     model.DomainLabelBiology,
	"ecology":                     model.DomainLabelBiology,
	"genetics":                    model.DomainLabelBiology,
	"microbiology":                model.DomainLabelBiology,
	"marine-biology":              model.DomainLabelBiology,
	"chemistry":                   model.DomainLabelChemistry,
	"organic-chemistry":           model.DomainLabelChemistry,
	"organic chemistry":           model.DomainLabelChemistry,
	"inorganic-chemistry":         model.DomainLabelChemistry,
	"inorganic chemistry":         model.DomainLabelChemistry,
	"analytical-chemistry":        model.DomainLabelChemistry,
	"analytical chemistry":        model.DomainLabelChemistry,
	"biochemistry":                model.DomainLabelChemistry,
	"physics":                     model.DomainLabelPhysics,
	"optics":                      model.DomainLabelPhysics,
	"quantum-mechanics":           model.DomainLabelPhysics,
	"quantum mechanics":           model.DomainLabelPhysics,
	"nuclear-physics":             model.DomainLabelPhysics,
	"nuclear physics":             model.DomainLabelPhysics,
	"thermodynamics":              model.DomainLabelPhysics,
	"energy":                      model.DomainLabelPhysics,
	"electricity":                 model.DomainLabelPhysics,
	"electromagnetism":            model.DomainLabelPhysics,
	"engineering":                 model.DomainLabelEngineering,
	"civil-engineering":           model.DomainLabelEngineering,
	"electrical-engineering":      model.DomainLabelEngineering,
	"manufacturing":               model.DomainLabelEngineering,
	"technology":                  model.DomainLabelEngineering,
	"mathematics":                 model.DomainLabelMathematics,
	"geometry":                    model.DomainLabelMathematics,
	"differential-geometry":       model.DomainLabelMathematics,
	"differential geometry":       model.DomainLabelMathematics,
	"algebra":                     model.DomainLabelMathematics,
	"statistics":                  model.DomainLabelMathematics,
	"botany":                      model.DomainLabelBotany,
	"plants":                      model.DomainLabelBotany,
	"mycology":                    model.DomainLabelBotany,
	"agriculture":                 model.DomainLabelBotany,
	"horticulture":                model.DomainLabelBotany,
	"zoology":                     model.DomainLabelZoology,
	"animals":                     model.DomainLabelZoology,
	"ornithology":                 model.DomainLabelZoology,
	"entomology":                  model.DomainLabelZoology,
	"ichthyology":                 model.DomainLabelZoology,
	"horses":                      model.DomainLabelZoology,
	"pets":                        model.DomainLabelZoology,
	"linguistics":                 model.DomainLabelLinguistics,
	"grammar":                     model.DomainLabelLinguistics,
	"phonetics":                   model.DomainLabelLinguistics,
	"phonology":                   model.DomainLabelLinguistics,
	"military":                    model.DomainLabelMilitary,
	"weaponry":                    model.DomainLabelMilitary,
	"army":                        model.DomainLabelMilitary,
	"navy":                        model.DomainLabelMilitary,
	"war":                         model.DomainLabelMilitary,
	"architecture":                model.DomainLabelArchitecture,
	"construction":                model.DomainLabelArchitecture,
	"building":                    model.DomainLabelArchitecture,
	"religion":                    model.DomainLabelReligion,
	"Christianity":                model.DomainLabelReligion,
	"Islam":                       model.DomainLabelReligion,
	"Buddhism":                    model.DomainLabelReligion,
	"theology":                    model.DomainLabelReligion,
	"mysticism":                   model.DomainLabelReligion,
	"biblical":                    model.DomainLabelReligion,
	"Catholicism":                 model.DomainLabelReligion,
	"politics":                    model.DomainLabelPolitics,
	"government":                  model.DomainLabelPolitics,
	"diplomacy":                   model.DomainLabelPolitics,
	"monarchy":                    model.DomainLabelPolitics,
	"cooking":                     model.DomainLabelCooking,
	"food":                        model.DomainLabelCooking,
	"beverages":                   model.DomainLabelCooking,
	"gastronomy":                  model.DomainLabelCooking,
	"nautical":                    model.DomainLabelNautical,
	"seafaring":                   model.DomainLabelNautical,
	"marine":                      model.DomainLabelNautical,
	"astronomy":                   model.DomainLabelAstronomy,
	"astrophysics":                model.DomainLabelAstronomy,
	"geology":                     model.DomainLabelGeology,
	"planetary-geology":           model.DomainLabelGeology,
	"planetary geology":           model.DomainLabelGeology,
	"mineralogy":                  model.DomainLabelGeology,
	"aviation":                    model.DomainLabelAviation,
	"aeronautics":                 model.DomainLabelAviation,
	"aerospace":                   model.DomainLabelAviation,
	"electronics":                 model.DomainLabelElectronics,
	"computer-keyboards":          model.DomainLabelElectronics,
	"computer keyboards":          model.DomainLabelElectronics,
	"semiconductors":              model.DomainLabelElectronics,
	"psychology":                  model.DomainLabelPsychology,
	"philosophy":                  model.DomainLabelPhilosophy,
	"games":                       model.DomainLabelGames,
	"board-games":                 model.DomainLabelGames,
	"video-games":                 model.DomainLabelGames,
	"roleplaying-games":           model.DomainLabelGames,
	"puzzles":                     model.DomainLabelGames,
	"chess":                       model.DomainLabelGames,
	"card-games":                  model.DomainLabelGames,
	"poker":                       model.DomainLabelGames,
	"media":                       model.DomainLabelMedia,
	"publishing":                  model.DomainLabelMedia,
	"film":                        model.DomainLabelMedia,
	"television":                  model.DomainLabelMedia,
	"video-editing":               model.DomainLabelMedia,
	"video editing":               model.DomainLabelMedia,
	"sound-recording":             model.DomainLabelMedia,
	"sound recording":             model.DomainLabelMedia,
	"sound-engineering":           model.DomainLabelMedia,
	"sound engineering":           model.DomainLabelMedia,
	"broadcasting":                model.DomainLabelMedia,
	"journalism":                  model.DomainLabelMedia,
	"communications":              model.DomainLabelMedia,
	"education":                   model.DomainLabelEducation,
	"academia":                    model.DomainLabelEducation,
	"transport":                   model.DomainLabelTransport,
	"rail-transport":              model.DomainLabelTransport,
	"rail transport":              model.DomainLabelTransport,
	"railways":                    model.DomainLabelTransport,
	"vehicles":                    model.DomainLabelTransport,
	"automotive":                  model.DomainLabelAutomotive,
	"automobiles":                 model.DomainLabelAutomotive,
	"cars":                        model.DomainLabelAutomotive,
	"printing":                    model.DomainLabelPrinting,
	"typography":                  model.DomainLabelPrinting,
	"mining":                      model.DomainLabelMining,
	"meteorology":                 model.DomainLabelMeteorology,
	"weather":                     model.DomainLabelMeteorology,
	"climatology":                 model.DomainLabelMeteorology,
	"heraldry":                    model.DomainLabelHeraldry,
	"materials-science":           model.DomainLabelMaterialsScience,
	"materials science":           model.DomainLabelMaterialsScience,
	"mythology":                   model.DomainLabelMythology,
	"art":                         model.DomainLabelArt,
	"arts":                        model.DomainLabelArt,
	"fine-arts":                   model.DomainLabelArt,
	"painting":                    model.DomainLabelArt,
	"theater":                     model.DomainLabelArt,
}

var expectedDroppedTopics = map[string]struct{}{
	"natural-sciences":  {},
	"sciences":          {},
	"physical-sciences": {},
	"human-sciences":    {},
	"lifestyle":         {},
	"hobbies":           {},
	"entertainment":     {},
	"literature":        {},
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

func TestNormalizeTopicCanonicalVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "case_variant", raw: "ORGANIC CHEMISTRY", want: model.DomainLabelChemistry, ok: true},
		{name: "underscore_variant", raw: "quantum_mechanics", want: model.DomainLabelPhysics, ok: true},
		{name: "hyphen_space_variant", raw: "materials science", want: model.DomainLabelMaterialsScience, ok: true},
		{name: "raw_gloss_rail_transport", raw: "rail transport", want: model.DomainLabelTransport, ok: true},
		{name: "raw_gloss_video_editing", raw: "video editing", want: model.DomainLabelMedia, ok: true},
		{name: "raw_gloss_sound_recording", raw: "sound recording", want: model.DomainLabelMedia, ok: true},
		{name: "raw_gloss_computer_keyboards", raw: "computer keyboards", want: model.DomainLabelElectronics, ok: true},
		{name: "raw_gloss_semiconductors", raw: "Semiconductors", want: model.DomainLabelElectronics, ok: true},
		{name: "raw_gloss_card_games", raw: "card games", want: model.DomainLabelGames, ok: true},
		{name: "raw_gloss_law_enforcement", raw: "law enforcement", want: model.DomainLabelLaw, ok: true},
		{name: "raw_gloss_graphical_user_interface", raw: "graphical user interface", want: model.DomainLabelComputing, ok: true},
		{name: "real_sample_topic_aerospace", raw: "aerospace", want: model.DomainLabelAviation, ok: true},
		{name: "real_sample_topic_broadcasting", raw: "broadcasting", want: model.DomainLabelMedia, ok: true},
		{name: "real_sample_topic_mineralogy", raw: "mineralogy", want: model.DomainLabelGeology, ok: true},
		{name: "real_sample_topic_railways", raw: "railways", want: model.DomainLabelTransport, ok: true},
		{name: "real_sample_topic_electromagnetism", raw: "electromagnetism", want: model.DomainLabelPhysics, ok: true},
		{name: "real_sample_topic_climatology", raw: "climatology", want: model.DomainLabelMeteorology, ok: true},
		{name: "dropped_variant", raw: "natural sciences", ok: false},
		{name: "substring_false_positive", raw: "videogamesmanship", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := NormalizeTopic(tt.raw)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("NormalizeTopic(%q) = (%q, %t); want (%q, %t)", tt.raw, got, ok, tt.want, tt.ok)
			}
		})
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
