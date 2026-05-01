package norm

import (
	"testing"

	"github.com/simp-lee/isdict-commons/model"
)

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

func TestNormalizeAudioAccentCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "local_en_us", raw: "en-us-sail.ogg", want: model.AccentAmerican, ok: true},
		{name: "local_en_au_case_variant", raw: "En-au-tag.ogg", want: model.AccentAustralian, ok: true},
		{name: "local_en_uk_space_in_word", raw: "En-uk-to read.ogg", want: model.AccentBritish, ok: true},
		{name: "local_en_ca_oga", raw: "En-ca-slap.oga", want: model.AccentCanadian, ok: true},
		{name: "local_en_gb_mixed_case", raw: "en-GB-far-fetched.ogg", want: model.AccentBritish, ok: true},
		{name: "local_en_in", raw: "en-in-robot.ogg", want: model.AccentIndian, ok: true},
		{name: "local_en_za_space", raw: "en-za-South Africa.ogg", want: model.AccentSouthAfrican, ok: true},
		{name: "local_en_nz", raw: "en-nz-after.ogg", want: model.AccentNZ, ok: true},
		{name: "local_en_ie", raw: "en-ie-after.ogg", want: model.AccentIrish, ok: true},
		{name: "local_uk_prefix", raw: "UK-after.ogg", want: model.AccentBritish, ok: true},
		{name: "underscore_prefix", raw: "en_us_sail.ogg", want: model.AccentAmerican, ok: true},
		{name: "space_prefix", raw: "en us sail.ogg", want: model.AccentAmerican, ok: true},
		{
			name: "transcoded_url_basename",
			raw:  "https://upload.wikimedia.org/wikipedia/commons/transcoded/a/a1/En-us-sail.ogg/En-us-sail.ogg.mp3",
			want: model.AccentAmerican,
			ok:   true,
		},
		{
			name: "url_decoded_basename",
			raw:  "https://upload.wikimedia.org/wikipedia/commons/transcoded/a/a1/En-uk-to%20read.ogg/En-uk-to%20read.ogg.mp3",
			want: model.AccentBritish,
			ok:   true,
		},
		{name: "html_unescaped_basename", raw: "En-au-tag&amp;more.ogg", want: model.AccentAustralian, ok: true},
		{name: "ll_q1860_not_language_prefix", raw: "LL-Q1860 (eng)-Vealhurl-paint.wav", ok: false},
		{name: "meow_usa_not_prefix", raw: "MeowUSA.wav", ok: false},
		{name: "pacific_northwest_not_prefix", raw: "Pacific-Northwest-After.wav", ok: false},
		{name: "unknown_en_prefix", raw: "en-xx-after.ogg", ok: false},
		{name: "embedded_prefix_not_start", raw: "Speech-en-us-after.ogg", ok: false},
		{name: "uk_underscore_not_prefix", raw: "UK_after.ogg", ok: false},
		{name: "blank", raw: "", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := NormalizeAudioAccentCode(tt.raw)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizeAudioAccentCode(%q) = (%q, %t); want (%q, %t)", tt.raw, got, ok, tt.want, tt.ok)
			}
		})
	}
}
