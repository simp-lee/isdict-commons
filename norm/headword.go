package norm

import "strings"

var headwordNormalizationReplacer = strings.NewReplacer(
	"\u2018", "'",
	"\u2019", "'",
	"-", "",
	" ", "",
	"_", "",
)

func NormalizeHeadword(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return headwordNormalizationReplacer.Replace(s)
}

func IsMultiword(headword string) bool {
	return strings.Contains(headword, " ") || strings.Contains(headword, "-")
}
