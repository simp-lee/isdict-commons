package norm

import (
	"maps"
	"strings"

	"github.com/simp-lee/isdict-commons/model"
)

var accentTagMap = buildAccentTagMap()

// AccentTagMap returns a defensive copy of the frozen exact-tag accent mapping contract.
func AccentTagMap() map[string]string {
	return cloneAccentTagMap(accentTagMap)
}

var accentTokenMap = buildAccentTokenMap(accentTagMap)

// NormalizeAccentCode keeps the commons baseline limited to single-tag exact and token-level matching.
func NormalizeAccentCode(rawTag string) string {
	rawTag = strings.TrimSpace(rawTag)
	if rawTag == "" {
		return model.AccentUnknown
	}

	if accentCode, ok := accentTagMap[rawTag]; ok {
		return accentCode
	}

	for _, candidate := range accentTokenCandidates(rawTag) {
		if accentCode, ok := accentTokenMap[candidate]; ok {
			return accentCode
		}
	}

	return model.AccentOtherRegional
}

func buildAccentTagMap() map[string]string {
	mappings := make(map[string]string, 64)
	add := func(accentCode string, rawTags ...string) {
		for _, rawTag := range rawTags {
			mappings[rawTag] = accentCode
		}
	}

	add(model.AccentUnknown, "unknown", "Unknown")

	add(model.AccentBritish,
		"Received-Pronunciation",
		"Received Pronunciation",
		"RP",
		"Conservative-RP",
		"Conservative RP",
		"UK",
		"GB",
		"England",
		"Standard-Southern-British",
	)

	add(model.AccentAmerican,
		"General-American",
		"General American",
		"GA",
		"GenAm",
		"US",
		"USA",
		"U.S.",
		"U.S.A.",
		"United-States",
		"North-American",
	)

	add(model.AccentAustralian,
		"General-Australian",
		"General Australian",
		"Australia",
	)

	add(model.AccentCanadian,
		"Canada",
		"Standard-Canadian",
	)

	add(model.AccentIrish,
		"Ireland",
		"Irish",
		"Hiberno-English",
		"Ulster",
		"Munster",
	)

	add(model.AccentScottish,
		"Scotland",
		"Scottish",
		"Scots",
		"Glaswegian",
		"Edinburgh",
	)

	add(model.AccentNZ,
		"New-Zealand",
		"NZ",
	)

	add(model.AccentIndian,
		"India",
		"Indian-English",
		"South-Asian",
		"Pakistan",
	)

	add(model.AccentSouthAfrican,
		"South-Africa",
		"South-African",
		"South-African-English",
	)

	add(model.AccentOtherRegional,
		"Jamaica",
		"Northern-Ireland",
		"Northern-England",
		"Southern-England",
		"Multicultural-London-English",
		"Cockney",
		"Estuary-English",
		"Geordie",
		"Scouse",
		"Yorkshire",
		"Lancashire",
		"West-Country",
		"East-Anglia",
		"Norfolk",
		"Essex",
		"London",
		"Cornwall",
		"New-York",
		"Boston",
		"Philadelphia",
		"California",
		"Texas",
		"Southern-US",
		"Appalachian",
		"New-England",
		"Mid-Atlantic",
		"Midwest",
		"Singapore",
		"Singlish",
		"Philippines",
		"Hong-Kong",
		"Malaysia",
		"Caribbean",
		"Nigeria",
		"Kenya",
		"Ghana",
		"Welsh",
		"Northumbria",
	)

	return mappings
}

func buildAccentTokenMap(source map[string]string) map[string]string {
	mappings := make(map[string]string, len(source)+32)
	for rawTag, accentCode := range source {
		mappings[strings.ToLower(rawTag)] = accentCode
	}

	add := func(accentCode string, tokens ...string) {
		for _, token := range tokens {
			mappings[token] = accentCode
		}
	}

	add(model.AccentBritish, "british", "britain")
	add(model.AccentAmerican, "american")
	add(model.AccentAustralian, "australian")
	add(model.AccentCanadian, "canadian")
	add(model.AccentIrish, "ireland", "irish", "ulster", "munster", "hiberno")
	add(model.AccentScottish, "scotland", "scottish", "scots", "glaswegian", "edinburgh")
	add(model.AccentIndian, "india", "indian", "pakistan", "south-asian")
	add(model.AccentSouthAfrican, "south-africa", "south-african")
	add(model.AccentOtherRegional,
		"jamaica",
		"cockney",
		"estuary-english",
		"geordie",
		"scouse",
		"yorkshire",
		"lancashire",
		"west-country",
		"east-anglia",
		"norfolk",
		"essex",
		"london",
		"cornwall",
		"new-york",
		"boston",
		"philadelphia",
		"california",
		"texas",
		"southern-us",
		"appalachian",
		"new-england",
		"mid-atlantic",
		"midwest",
		"singapore",
		"singlish",
		"philippines",
		"hong-kong",
		"malaysia",
		"caribbean",
		"nigeria",
		"kenya",
		"ghana",
		"welsh",
		"northumbria",
	)

	return mappings
}

func accentTokenCandidates(rawTag string) []string {
	lower := strings.ToLower(strings.TrimSpace(rawTag))
	fields := strings.FieldsFunc(lower, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		default:
			return true
		}
	})
	if len(fields) == 0 {
		return nil
	}

	candidates := make([]string, 0, 1+len(fields)*(len(fields)+1)/2)
	seen := make(map[string]struct{}, 1+len(fields)*(len(fields)+1)/2)
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	add(lower)
	for width := len(fields); width >= 2; width-- {
		for start := 0; start+width <= len(fields); start++ {
			add(strings.Join(fields[start:start+width], "-"))
		}
	}
	for _, field := range fields {
		add(field)
	}

	return candidates
}

func cloneAccentTagMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	maps.Copy(clone, source)

	return clone
}
