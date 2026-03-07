package model

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// AC-7: 更新 WordAnnotations 注释以反映 string 类型的 CEFR 语义
func TestWordAnnotationsCommentMentionsStringCEFR(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		wantErr string
	}{
		{
			name:    "accepts current string semantics comment",
			comment: mustReadWordAnnotationsCEFRComment(t),
		},
		{
			name:    "rejects legacy integer semantics comment",
			comment: "CEFR level: A1=1,A2=2,B1=3,B2=4,C1=5,C2=6,0=unknown",
			wantErr: "missing string CEFR values",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateWordAnnotationsCEFRComment(tt.comment)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateWordAnnotationsCEFRComment() error = %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("validateWordAnnotationsCEFRComment() error = nil; want substring %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateWordAnnotationsCEFRComment() error = %q; want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func mustReadWordAnnotationsCEFRComment(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	wordFile := filepath.Join(filepath.Dir(currentFile), "word.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, wordFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parser.ParseFile(%q): %v", wordFile, err)
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "WordAnnotations" {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("WordAnnotations is %T; want struct", typeSpec.Type)
			}

			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					if name.Name != "CEFRLevel" {
						continue
					}

					if field.Comment != nil {
						return field.Comment.Text()
					}
					if field.Doc != nil {
						return field.Doc.Text()
					}
					t.Fatal("WordAnnotations.CEFRLevel comment is missing")
				}
			}

			t.Fatal("WordAnnotations.CEFRLevel field not found")
		}
	}

	t.Fatal("WordAnnotations type declaration not found")
	return ""
}

func validateWordAnnotationsCEFRComment(comment string) error {
	normalized := strings.ToUpper(strings.Join(strings.Fields(comment), " "))
	if strings.Contains(normalized, "A1=1") || strings.Contains(normalized, "0=UNKNOWN") {
		return fmt.Errorf("missing string CEFR values: found legacy integer-coded wording in %q", comment)
	}

	requiredPhrases := []string{"A1", "A2", "B1", "B2", "C1", "C2", "\"\"", "UNKNOWN"}
	for _, phrase := range requiredPhrases {
		if !strings.Contains(normalized, phrase) {
			return fmt.Errorf("missing string CEFR values: %q not found in %q", phrase, comment)
		}
	}

	return nil
}
