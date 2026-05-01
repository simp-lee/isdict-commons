package norm

import (
	"html"
	"strings"
	"unicode"
)

func canonicalLookupText(raw string) string {
	raw = strings.ToLower(html.UnescapeString(strings.TrimSpace(raw)))
	if raw == "" {
		return ""
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(fields) == 0 {
		return ""
	}

	return strings.Join(fields, " ")
}
