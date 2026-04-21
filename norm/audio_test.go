package norm

import "testing"

func TestAudioLocalFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: ""},
		{name: "wav_appends_mp3", raw: "LL-Q1860 (eng)-Vealhurl-paint.wav", want: "LL-Q1860_(eng)-Vealhurl-paint.wav.mp3"},
		{name: "ogg_appends_mp3", raw: "En-uk-book.ogg", want: "En-uk-book.ogg.mp3"},
		{name: "mp3_passthrough", raw: "En-us-book.mp3", want: "En-us-book.mp3"},
		{name: "html_entity_ampersand_decodes", raw: "Tom &amp; Jerry.wav", want: "Tom_&_Jerry.wav.mp3"},
		{name: "html_entity_space_decodes_after_space_replacement", raw: "Tom&#32;Jerry.ogg", want: "Tom Jerry.ogg.mp3"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := AudioLocalFilename(tt.raw); got != tt.want {
				t.Fatalf("AudioLocalFilename(%q) = %q; want %q", tt.raw, got, tt.want)
			}
		})
	}
}
