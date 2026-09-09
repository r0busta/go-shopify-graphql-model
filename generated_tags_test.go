package main_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

// Invariants of the generated models that the generator's mutate hook is
// responsible for: pointer and slice fields carry omitempty, and no json tag
// option is repeated.
func TestGeneratedJSONTags(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "graph/model/models_gen.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse models_gen.go: %v", err)
	}

	var checked int
	ast.Inspect(f, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range st.Fields.List {
			if field.Tag == nil || len(field.Names) == 0 {
				continue
			}
			checked++
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			jsonTag, ok := tag.Lookup("json")
			if !ok {
				continue
			}
			pos := fset.Position(field.Pos())
			name := field.Names[0].Name

			parts := strings.Split(jsonTag, ",")
			opts := parts[1:]
			seen := map[string]bool{}
			for _, opt := range opts {
				if seen[opt] {
					t.Errorf("%s: %s: json option %q repeated in tag %q", pos, name, opt, jsonTag)
				}
				seen[opt] = true
			}

			switch field.Type.(type) {
			case *ast.StarExpr, *ast.ArrayType:
				if !seen["omitempty"] {
					t.Errorf("%s: %s: pointer or slice field without omitempty in tag %q", pos, name, jsonTag)
				}
			}
		}
		return true
	})

	if checked == 0 {
		t.Fatal("no struct fields found in models_gen.go")
	}
}
