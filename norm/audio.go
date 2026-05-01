package norm

import (
	"html"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/simp-lee/isdict-commons/model"
)

// AudioLocalFilename defines the local MP3 cache filename contract only.
// Callers must pair the returned name with actual MP3 bytes, not renamed wav/ogg/oga payloads.
func AudioLocalFilename(rawAudioField string) string {
	if rawAudioField == "" {
		return ""
	}

	filename := html.UnescapeString(strings.ReplaceAll(rawAudioField, " ", "_"))
	if filepath.Ext(rawAudioField) != ".mp3" {
		return filename + ".mp3"
	}

	return filename
}

// NormalizeAudioAccentCode infers a controlled accent code from high-confidence
// Wiktionary audio filename prefixes.
func NormalizeAudioAccentCode(rawAudioField string) (accentCode string, ok bool) {
	filename := audioAccentBasename(rawAudioField)
	if filename == "" {
		return "", false
	}

	prefixes := []struct {
		prefixes   []string
		accentCode string
	}{
		{prefixes: []string{"en-us"}, accentCode: model.AccentAmerican},
		{prefixes: []string{"en-uk", "en-gb"}, accentCode: model.AccentBritish},
		{prefixes: []string{"en-au"}, accentCode: model.AccentAustralian},
		{prefixes: []string{"en-ca"}, accentCode: model.AccentCanadian},
		{prefixes: []string{"en-nz"}, accentCode: model.AccentNZ},
		{prefixes: []string{"en-in"}, accentCode: model.AccentIndian},
		{prefixes: []string{"en-ie"}, accentCode: model.AccentIrish},
		{prefixes: []string{"en-za"}, accentCode: model.AccentSouthAfrican},
	}

	for _, candidate := range prefixes {
		for _, prefix := range candidate.prefixes {
			if audioFilenameHasAccentPrefix(filename, prefix, true) {
				return candidate.accentCode, true
			}
		}
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(filename)), "uk-") {
		return model.AccentBritish, true
	}

	return "", false
}

func audioAccentBasename(rawAudioField string) string {
	rawAudioField = strings.TrimSpace(rawAudioField)
	if rawAudioField == "" {
		return ""
	}

	value := html.UnescapeString(rawAudioField)
	if parsedURL, err := url.Parse(value); err == nil && parsedURL.Path != "" {
		decodedPath := parsedURL.EscapedPath()
		if decoded, err := url.PathUnescape(decodedPath); err == nil {
			decodedPath = decoded
		}
		value = path.Base(decodedPath)
	} else {
		value = path.Base(value)
	}

	value = html.UnescapeString(value)
	if decoded, err := url.PathUnescape(value); err == nil {
		value = decoded
	}

	return strings.TrimSpace(value)
}

func audioFilenameHasAccentPrefix(filename, prefix string, allowDotBoundary bool) bool {
	filename = strings.ToLower(strings.TrimSpace(filename))
	prefixParts := strings.Split(prefix, "-")

	offset := 0
	for index, part := range prefixParts {
		if !strings.HasPrefix(filename[offset:], part) {
			return false
		}
		offset += len(part)
		if index == len(prefixParts)-1 {
			break
		}
		if offset >= len(filename) || !isAudioFilenamePrefixSeparator(rune(filename[offset])) {
			return false
		}
		offset++
	}

	if offset >= len(filename) {
		return false
	}

	next := rune(filename[offset])
	if isAudioFilenamePrefixSeparator(next) {
		return true
	}

	return allowDotBoundary && next == '.'
}

func isAudioFilenamePrefixSeparator(r rune) bool {
	return r == '-' || r == '_' || r == ' '
}
