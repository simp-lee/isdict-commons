package norm

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/liuzl/gocc"
)

type ZhNormalizer struct {
	converter                *gocc.OpenCC
	controlCharacterReplacer *strings.Replacer
}

func NewZhNormalizer() (*ZhNormalizer, error) {
	converter, err := gocc.New("t2s")
	if err != nil {
		return nil, err
	}

	return &ZhNormalizer{
		converter:                converter,
		controlCharacterReplacer: newZhControlCharacterReplacer(),
	}, nil
}

func (z *ZhNormalizer) Normalize(raw string) string {
	normalized := normalizeZhWhitespace(z.controlCharacterReplacer.Replace(raw))

	converted, err := z.converter.Convert(normalized)
	if err != nil {
		panic(fmt.Errorf("zh normalization convert failed: %w", err))
	}

	return strings.TrimSpace(converted)
}

func newZhControlCharacterReplacer() *strings.Replacer {
	return strings.NewReplacer(
		"\u200b", "",
		"\ufeff", "",
		"\u200e", "",
		"\u200f", "",
	)
}

func normalizeZhWhitespace(raw string) string {
	var builder strings.Builder
	builder.Grow(len(raw))

	previousWasSpace := false
	for _, r := range raw {
		if isZhWhitespace(r) {
			if builder.Len() == 0 || previousWasSpace {
				previousWasSpace = true
				continue
			}

			builder.WriteByte(' ')
			previousWasSpace = true
			continue
		}

		builder.WriteRune(r)
		previousWasSpace = false
	}

	return strings.TrimSpace(builder.String())
}

func isZhWhitespace(r rune) bool {
	return r == '\u00a0' || r == '\u3000' || unicode.IsSpace(r)
}
