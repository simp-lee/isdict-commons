package norm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/liuzl/da"
	"github.com/liuzl/gocc"
)

func TestZhNormalizerNormalize(t *testing.T) {
	normalizer := mustNewZhNormalizer(t)

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "traditional_to_simplified", raw: "計算機科學", want: "计算机科学"},
		{name: "full_width_punctuation_preserved", raw: "你好，世界！「測試」", want: "你好，世界！「测试」"},
		{name: "zwsp_and_bom_removed", raw: "\ufeff計\u200b算\u200b機", want: "计算机"},
		{name: "nbsp_and_repeated_spaces_collapsed", raw: "\u00a0中文\u00a0\u00a0 測試\u3000\t例子  ", want: "中文 测试 例子"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizer.Normalize(tt.raw); got != tt.want {
				t.Fatalf("Normalize(%q) = %q; want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNewZhNormalizerReturnsReusableInstances(t *testing.T) {
	first := mustNewZhNormalizer(t)
	second := mustNewZhNormalizer(t)

	if first == second {
		t.Fatal("NewZhNormalizer() returned the same normalizer instance; want independent instances")
	}

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "first_call", raw: "計算機科學", want: "计算机科学"},
		{name: "reused_after_setup", raw: "\ufeff繁\u200b體\u00a0\u00a0中文", want: "繁体 中文"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := first.Normalize(tt.raw); got != tt.want {
				t.Fatalf("first.Normalize(%q) = %q; want %q", tt.raw, got, tt.want)
			}

			if got := second.Normalize(tt.raw); got != tt.want {
				t.Fatalf("second.Normalize(%q) = %q; want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestZhNormalizerNormalizeConcurrentSharedInstance(t *testing.T) {
	normalizer := mustNewZhNormalizer(t)

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "traditional_to_simplified", raw: "計算機科學", want: "计算机科学"},
		{name: "full_width_punctuation_preserved", raw: "你好，世界！「測試」", want: "你好，世界！「测试」"},
		{name: "zwsp_and_bom_removed", raw: "\ufeff計\u200b算\u200b機", want: "计算机"},
		{name: "nbsp_and_repeated_spaces_collapsed", raw: "\u00a0中文\u00a0\u00a0 測試\u3000\t例子  ", want: "中文 测试 例子"},
	}

	const iterations = 64
	errCh := make(chan error, len(tests)*iterations)
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		for _, tt := range tests {
			tt := tt
			wg.Add(1)
			go func() {
				defer wg.Done()

				if got := normalizer.Normalize(tt.raw); got != tt.want {
					errCh <- fmt.Errorf("%s: Normalize(%q) = %q; want %q", tt.name, tt.raw, got, tt.want)
				}
			}()
		}
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// AC-R008: 当 OpenCC Convert 失败时，ZhNormalizer.Normalize 必须显式暴露故障，不能静默返回未转换文本。
func TestZhNormalizerNormalizePanicsOnConvertFailure(t *testing.T) {
	t.Parallel()

	// REG-008
	// Ledger Key: norm/zh|convert-error-silent-fallback
	normalizer := &ZhNormalizer{
		converter:                newBrokenOpenCCConverterForTest(t),
		controlCharacterReplacer: newZhControlCharacterReplacer(),
	}
	raw := "\ufeff測\u200b試\u00a0文字"
	wantFallback := normalizeZhWhitespace(newZhControlCharacterReplacer().Replace(raw))
	if wantFallback == "" {
		t.Fatal("fallback normalization unexpectedly empty")
	}

	var got string
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()

		got = normalizer.Normalize(raw)
	}()

	if recovered == nil {
		t.Fatalf("Normalize(%q) completed without panic; want convert failure to be exposed", raw)
	}

	err, ok := recovered.(error)
	if !ok {
		t.Fatalf("panic = %T (%v); want error", recovered, recovered)
	}

	if !strings.Contains(err.Error(), "zh normalization convert failed") {
		t.Fatalf("panic error = %q; want wrapper context", err.Error())
	}

	if !strings.Contains(err.Error(), "Trie is nil") {
		t.Fatalf("panic error = %q; want underlying convert failure", err.Error())
	}

	if got == wantFallback {
		t.Fatalf("Normalize(%q) returned untranslated fallback %q; want panic", raw, got)
	}

	if got != "" {
		t.Fatalf("Normalize(%q) returned %q before panic; want no returned value", raw, got)
	}
}

func mustNewZhNormalizer(t *testing.T) *ZhNormalizer {
	t.Helper()

	normalizer, err := NewZhNormalizer()
	if err != nil {
		t.Fatalf("NewZhNormalizer() error = %v; want nil", err)
	}

	return normalizer
}

func newBrokenOpenCCConverterForTest(t *testing.T) *gocc.OpenCC {
	t.Helper()

	brokenGroup := &gocc.Group{}
	groupValue := reflect.ValueOf(brokenGroup).Elem()
	dictSliceType := reflect.TypeOf([]*da.Dict{})
	dictFieldIndex := -1
	for i := 0; i < groupValue.NumField(); i++ {
		field := groupValue.Field(i)
		if !field.CanSet() || field.Type() != dictSliceType {
			continue
		}
		if dictFieldIndex != -1 {
			t.Fatalf("gocc.Group exposes multiple []*da.Dict fields; cannot build deterministic broken converter")
		}
		dictFieldIndex = i
	}

	if dictFieldIndex == -1 {
		t.Fatalf("gocc.Group does not expose a []*da.Dict field; cannot build broken converter")
	}

	groupValue.Field(dictFieldIndex).Set(reflect.ValueOf([]*da.Dict{{}}))

	return &gocc.OpenCC{DictChains: []*gocc.Group{brokenGroup}}
}
