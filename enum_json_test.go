package main_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/r0busta/go-shopify-graphql-model/v5/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Enums are plain string types and rely on encoding/json's default handling.
// Newer gqlgen versions can generate MarshalJSON/UnmarshalJSON methods for
// them (see omit_enum_json_marshalers in gqlgen.yml); those methods turn a
// JSON null in a non-pointer enum field into a decode error, which Shopify
// responses trigger whenever a non-null field is nulled out by an error.
func TestGeneratedEnumsHaveNoJSONMethods(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "graph/model/models_gen.go", nil, 0)
	require.NoError(t, err)

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}
		if fn.Name.Name == "MarshalJSON" || fn.Name.Name == "UnmarshalJSON" {
			t.Errorf("%s: generated code defines %s; set omit_enum_json_marshalers in gqlgen.yml",
				fset.Position(fn.Pos()), fn.Name.Name)
		}
	}
}

func TestEnumNullIsNoOp(t *testing.T) {
	var p model.Product
	require.NoError(t, json.Unmarshal([]byte(`{"id":"gid://shopify/Product/1","status":null}`), &p))
	assert.Equal(t, model.ProductStatus(""), p.Status)

	require.NoError(t, json.Unmarshal([]byte(`{"id":"gid://shopify/Product/1","status":"ACTIVE"}`), &p))
	assert.Equal(t, model.ProductStatusActive, p.Status)

	out, err := json.Marshal(struct {
		Status model.ProductStatus `json:"status"`
	}{model.ProductStatusDraft})
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":"DRAFT"}`, string(out))
}
