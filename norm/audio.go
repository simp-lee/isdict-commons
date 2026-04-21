package norm

import (
	"html"
	"path/filepath"
	"strings"
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
