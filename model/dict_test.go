package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestWordVariantFormTypeTagIncludesKindConstraint(t *testing.T) {
	field, ok := reflect.TypeOf(WordVariant{}).FieldByName("FormType")
	if !ok {
		t.Fatal("WordVariant.FormType field not found")
	}

	tag := field.Tag.Get("gorm")
	requiredFragments := []string{
		"check:form_type BETWEEN 1 AND 9",
		"check:word_variants_kind_form_type",
		"kind = 1 AND form_type IS NOT NULL",
		"kind = 2 AND form_type IS NULL",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(tag, fragment) {
			t.Fatalf("WordVariant.FormType gorm tag %q does not contain %q", tag, fragment)
		}
	}
}
