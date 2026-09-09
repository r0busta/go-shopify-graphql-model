package main_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/r0busta/go-shopify-graphql-model/v5/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mediaNodes holds one response object per Media implementation, as returned
// by a query that selects __typename and id plus type-specific fields.
var mediaNodes = map[string]string{
	"MediaImage": `{"__typename":"MediaImage","id":"gid://shopify/MediaImage/1","alt":"front","mimeType":"image/jpeg","mediaContentType":"IMAGE","status":"READY",
		"image":{"url":"https://cdn/1.jpg","width":800,"height":600,"altText":"front"},
		"originalSource":{"fileSize":1024,"url":"https://cdn/1-orig.jpg"}}`,
	"Video": `{"__typename":"Video","id":"gid://shopify/Video/2","alt":"spin","duration":12345,
		"originalSource":{"url":"https://cdn/2.mp4","format":"mp4","mimeType":"video/mp4","height":1080,"width":1920},
		"sources":[{"url":"https://cdn/2-720.mp4","format":"mp4","mimeType":"video/mp4","height":720,"width":1280}],
		"preview":{"image":{"url":"https://cdn/2.jpg"}}}`,
	"Model3d": `{"__typename":"Model3d","id":"gid://shopify/Model3d/3",
		"originalSource":{"url":"https://cdn/3.glb","format":"glb","filesize":4096,"mimeType":"model/gltf-binary"},
		"boundingBox":{"size":{"x":1.5,"y":2.0,"z":0.5}}}`,
	"ExternalVideo": `{"__typename":"ExternalVideo","id":"gid://shopify/ExternalVideo/4","originUrl":"https://youtu.be/x","embedUrl":"https://youtube.com/embed/x","host":"YOUTUBE"}`,
}

// mediaTypeChecks verifies type-specific fields survived decoding.
var mediaTypeChecks = map[string]func(t *testing.T, m model.Media){
	"MediaImage": func(t *testing.T, m model.Media) {
		img, ok := m.(*model.MediaImage)
		require.True(t, ok, "got %T", m)
		assert.Equal(t, "gid://shopify/MediaImage/1", img.ID)
		assert.Equal(t, "image/jpeg", *img.MimeType)
		assert.Equal(t, model.MediaContentTypeImage, img.MediaContentType)
		assert.Equal(t, model.MediaStatusReady, img.Status)
		require.NotNil(t, img.Image)
		assert.Equal(t, "https://cdn/1.jpg", img.Image.URL)
		assert.Equal(t, 800, *img.Image.Width)
		require.NotNil(t, img.OriginalSource)
		assert.Equal(t, 1024, *img.OriginalSource.FileSize)
	},
	"Video": func(t *testing.T, m model.Media) {
		vid, ok := m.(*model.Video)
		require.True(t, ok, "got %T", m)
		assert.Equal(t, 12345, *vid.Duration)
		require.NotNil(t, vid.OriginalSource)
		assert.Equal(t, 1920, vid.OriginalSource.Width)
		require.Len(t, vid.Sources, 1)
		assert.Equal(t, 720, vid.Sources[0].Height)
		require.NotNil(t, vid.Preview)
		require.NotNil(t, vid.Preview.Image)
		assert.Equal(t, "https://cdn/2.jpg", vid.Preview.Image.URL)
	},
	"Model3d": func(t *testing.T, m model.Media) {
		m3d, ok := m.(*model.Model3d)
		require.True(t, ok, "got %T", m)
		require.NotNil(t, m3d.OriginalSource)
		assert.Equal(t, 4096, m3d.OriginalSource.Filesize)
		require.NotNil(t, m3d.BoundingBox)
		assert.Equal(t, 1.5, m3d.BoundingBox.Size.X)
	},
	"ExternalVideo": func(t *testing.T, m model.Media) {
		ext, ok := m.(*model.ExternalVideo)
		require.True(t, ok, "got %T", m)
		assert.Equal(t, "https://youtu.be/x", ext.OriginURL)
		assert.Equal(t, "https://youtube.com/embed/x", ext.EmbedURL)
		assert.Equal(t, model.MediaHostYoutube, ext.Host)
	},
}

func mediaConnectionJSON(nodes ...string) string {
	edges := make([]string, len(nodes))
	for i, n := range nodes {
		edges[i] = fmt.Sprintf(`{"cursor":"c%d","node":%s}`, i, n)
	}
	return fmt.Sprintf(`{"edges":[%s],"nodes":[%s],"pageInfo":{"hasNextPage":true,"hasPreviousPage":false,"endCursor":"c%d"}}`,
		strings.Join(edges, ","), strings.Join(nodes, ","), len(nodes)-1)
}

func TestMediaDecodesEachImplementation(t *testing.T) {
	for name, node := range mediaNodes {
		t.Run(name, func(t *testing.T) {
			var conn model.MediaConnection
			require.NoError(t, json.Unmarshal([]byte(mediaConnectionJSON(node)), &conn))

			require.Len(t, conn.Edges, 1)
			require.Len(t, conn.Nodes, 1)
			assert.Equal(t, "c0", conn.Edges[0].Cursor)
			require.NotNil(t, conn.PageInfo)
			assert.True(t, conn.PageInfo.HasNextPage)

			for _, m := range []model.Media{conn.Edges[0].Node, conn.Nodes[0]} {
				require.NotNil(t, m)
				assert.Equal(t, reflect.Ptr, reflect.TypeOf(m).Kind(), "decoded Media must be a pointer, got %T", m)
				mediaTypeChecks[name](t, m)
			}
		})
	}
}

func TestMediaDecodesMixedConnection(t *testing.T) {
	var conn model.MediaConnection
	require.NoError(t, json.Unmarshal([]byte(mediaConnectionJSON(
		mediaNodes["MediaImage"], mediaNodes["Video"], mediaNodes["Model3d"], mediaNodes["ExternalVideo"],
	)), &conn))
	require.Len(t, conn.Edges, 4)
	require.Len(t, conn.Nodes, 4)
	assert.IsType(t, &model.MediaImage{}, conn.Edges[0].Node)
	assert.IsType(t, &model.Video{}, conn.Edges[1].Node)
	assert.IsType(t, &model.Model3d{}, conn.Edges[2].Node)
	assert.IsType(t, &model.ExternalVideo{}, conn.Edges[3].Node)
	assert.IsType(t, &model.ExternalVideo{}, conn.Nodes[3])
	assert.Equal(t, "c3", conn.Edges[3].Cursor)
}

func TestMediaTypeSelection(t *testing.T) {
	tests := []struct {
		name    string
		node    string
		want    interface{}
		wantErr string
	}{
		{name: "typename only", node: `{"__typename":"Video","alt":"x"}`, want: &model.Video{}},
		{name: "id only", node: `{"id":"gid://shopify/Model3d/3"}`, want: &model.Model3d{}},
		{name: "typename wins over id", node: `{"__typename":"Video","id":"gid://shopify/MediaImage/1"}`, want: &model.Video{}},
		{name: "neither selected", node: `{"alt":"x"}`, wantErr: "__typename or id"},
		{name: "unknown typename", node: `{"__typename":"Hologram","id":"gid://shopify/Hologram/9"}`, wantErr: `"Hologram"`},
		{name: "id of another resource", node: `{"id":"gid://shopify/Product/1"}`, wantErr: `"Product"`},
		{name: "malformed id", node: `{"id":"not-a-gid"}`, wantErr: "__typename or id"},
		{name: "wrong field type", node: `{"__typename":"Video","duration":"long"}`, wantErr: "Video"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var edge model.MediaEdge
			err := json.Unmarshal([]byte(`{"cursor":"c","node":`+tt.node+`}`), &edge)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.IsType(t, tt.want, edge.Node)
			assert.Equal(t, "c", edge.Cursor)
		})
	}
}

func TestMediaEmptyAndNull(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty object", in: `{}`},
		{name: "empty lists", in: `{"edges":[],"nodes":[]}`},
		{name: "null lists", in: `{"edges":null,"nodes":null}`},
		{name: "null node", in: `{"edges":[{"cursor":"c","node":null}],"nodes":[null]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var conn model.MediaConnection
			require.NoError(t, json.Unmarshal([]byte(tt.in), &conn))
			for _, e := range conn.Edges {
				assert.Nil(t, e.Node)
			}
			for _, n := range conn.Nodes {
				assert.Nil(t, n)
			}
		})
	}
}

func TestMediaDecodesWhenNested(t *testing.T) {
	t.Run("product", func(t *testing.T) {
		var p model.Product
		require.NoError(t, json.Unmarshal([]byte(`{"id":"gid://shopify/Product/1","title":"T","media":`+mediaConnectionJSON(mediaNodes["Video"])+`}`), &p))
		assert.Equal(t, "T", p.Title)
		require.NotNil(t, p.Media)
		require.Len(t, p.Media.Edges, 1)
		assert.IsType(t, &model.Video{}, p.Media.Edges[0].Node)
	})
	t.Run("product variant", func(t *testing.T) {
		var v model.ProductVariant
		require.NoError(t, json.Unmarshal([]byte(`{"id":"gid://shopify/ProductVariant/1","media":`+mediaConnectionJSON(mediaNodes["MediaImage"])+`}`), &v))
		require.NotNil(t, v.Media)
		require.Len(t, v.Media.Nodes, 1)
		assert.IsType(t, &model.MediaImage{}, v.Media.Nodes[0])
	})
	t.Run("product list", func(t *testing.T) {
		var products []model.Product
		require.NoError(t, json.Unmarshal([]byte(`[{"id":"gid://shopify/Product/1","media":`+mediaConnectionJSON(mediaNodes["Model3d"])+`},{"id":"gid://shopify/Product/2"}]`), &products))
		require.Len(t, products, 2)
		require.NotNil(t, products[0].Media)
		assert.IsType(t, &model.Model3d{}, products[0].Media.Edges[0].Node)
		assert.Nil(t, products[1].Media)
	})
}

func TestMediaMarshalRoundTrip(t *testing.T) {
	var conn model.MediaConnection
	require.NoError(t, json.Unmarshal([]byte(mediaConnectionJSON(mediaNodes["MediaImage"], mediaNodes["Video"])), &conn))

	out, err := json.Marshal(conn)
	require.NoError(t, err)

	var again model.MediaConnection
	require.NoError(t, json.Unmarshal(out, &again))
	assert.Equal(t, conn, again)
}

// conditionObjects holds one response object per CollectionRuleConditionObject
// member, as returned by a query that selects __typename.
var conditionObjects = map[string]string{
	"CollectionRuleCategoryCondition":        `{"__typename":"CollectionRuleCategoryCondition","value":{"id":"gid://shopify/TaxonomyCategory/aa-1","name":"Apparel","fullName":"Apparel","isLeaf":false,"isRoot":true,"level":1}}`,
	"CollectionRuleMetafieldCondition":       `{"__typename":"CollectionRuleMetafieldCondition","metafieldDefinition":{"id":"gid://shopify/MetafieldDefinition/1","key":"colour","namespace":"custom","name":"Colour","ownerType":"PRODUCT"}}`,
	"CollectionRuleProductCategoryCondition": `{"__typename":"CollectionRuleProductCategoryCondition","value":{"id":"gid://shopify/ProductTaxonomyNode/2","name":"Shoes","fullName":"Apparel > Shoes","isLeaf":true,"isRoot":false}}`,
	"CollectionRuleTextCondition":            `{"__typename":"CollectionRuleTextCondition","value":"sale"}`,
}

var conditionObjectChecks = map[string]func(t *testing.T, o model.CollectionRuleConditionObject){
	"CollectionRuleCategoryCondition": func(t *testing.T, o model.CollectionRuleConditionObject) {
		c, ok := o.(*model.CollectionRuleCategoryCondition)
		require.True(t, ok, "got %T", o)
		require.NotNil(t, c.Value)
		assert.Equal(t, "Apparel", c.Value.Name)
		assert.True(t, c.Value.IsRoot)
	},
	"CollectionRuleMetafieldCondition": func(t *testing.T, o model.CollectionRuleConditionObject) {
		c, ok := o.(*model.CollectionRuleMetafieldCondition)
		require.True(t, ok, "got %T", o)
		require.NotNil(t, c.MetafieldDefinition)
		assert.Equal(t, "colour", c.MetafieldDefinition.Key)
		assert.Equal(t, model.MetafieldOwnerTypeProduct, c.MetafieldDefinition.OwnerType)
	},
	"CollectionRuleProductCategoryCondition": func(t *testing.T, o model.CollectionRuleConditionObject) {
		c, ok := o.(*model.CollectionRuleProductCategoryCondition)
		require.True(t, ok, "got %T", o)
		require.NotNil(t, c.Value)
		assert.Equal(t, "Apparel > Shoes", c.Value.FullName)
		assert.True(t, c.Value.IsLeaf)
	},
	"CollectionRuleTextCondition": func(t *testing.T, o model.CollectionRuleConditionObject) {
		c, ok := o.(*model.CollectionRuleTextCondition)
		require.True(t, ok, "got %T", o)
		assert.Equal(t, "sale", c.Value)
	},
}

func TestCollectionRuleDecodesEachConditionObject(t *testing.T) {
	for name, obj := range conditionObjects {
		t.Run(name, func(t *testing.T) {
			var rule model.CollectionRule
			require.NoError(t, json.Unmarshal([]byte(`{"column":"TAG","condition":"sale","relation":"EQUALS","conditionObject":`+obj+`}`), &rule))
			assert.Equal(t, model.CollectionRuleColumnTag, rule.Column)
			assert.Equal(t, "sale", rule.Condition)
			assert.Equal(t, model.CollectionRuleRelationEquals, rule.Relation)
			require.NotNil(t, rule.ConditionObject)
			assert.Equal(t, reflect.Ptr, reflect.TypeOf(rule.ConditionObject).Kind())
			conditionObjectChecks[name](t, rule.ConditionObject)
		})
	}
}

func TestCollectionRuleConditionObjectSelection(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		wantNil bool
		wantErr string
	}{
		{name: "absent", rule: `{"column":"TAG","condition":"sale","relation":"EQUALS"}`, wantNil: true},
		{name: "null", rule: `{"column":"TAG","condition":"sale","relation":"EQUALS","conditionObject":null}`, wantNil: true},
		{name: "missing typename", rule: `{"column":"TITLE","condition":"x","relation":"CONTAINS","conditionObject":{"value":"x"}}`, wantErr: "__typename"},
		{name: "unknown typename", rule: `{"column":"TITLE","condition":"x","relation":"CONTAINS","conditionObject":{"__typename":"CollectionRuleFutureCondition","value":"x"}}`, wantErr: `"CollectionRuleFutureCondition"`},
		{name: "wrong field type", rule: `{"column":"TITLE","condition":"x","relation":"CONTAINS","conditionObject":{"__typename":"CollectionRuleTextCondition","value":42}}`, wantErr: "CollectionRuleTextCondition"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule model.CollectionRule
			err := json.Unmarshal([]byte(tt.rule), &rule)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Nil(t, rule.ConditionObject)
			assert.NotEmpty(t, rule.Column)
		})
	}
}

func TestCollectionRuleDecodesWhenNested(t *testing.T) {
	var c model.Collection
	require.NoError(t, json.Unmarshal([]byte(`{"id":"gid://shopify/Collection/1","ruleSet":{"appliedDisjunctively":true,"rules":[
		{"column":"TAG","condition":"sale","relation":"EQUALS"},
		{"column":"TITLE","condition":"x","relation":"CONTAINS","conditionObject":`+conditionObjects["CollectionRuleTextCondition"]+`}
	]}}`), &c))
	require.NotNil(t, c.RuleSet)
	assert.True(t, c.RuleSet.AppliedDisjunctively)
	require.Len(t, c.RuleSet.Rules, 2)
	assert.Nil(t, c.RuleSet.Rules[0].ConditionObject)
	assert.IsType(t, &model.CollectionRuleTextCondition{}, c.RuleSet.Rules[1].ConditionObject)
}

func TestCollectionRuleMarshalRoundTrip(t *testing.T) {
	var rule model.CollectionRule
	require.NoError(t, json.Unmarshal([]byte(`{"column":"TAG","condition":"sale","relation":"EQUALS","conditionObject":`+conditionObjects["CollectionRuleMetafieldCondition"]+`}`), &rule))

	out, err := json.Marshal(rule)
	require.NoError(t, err)

	var again model.CollectionRule
	require.NoError(t, json.Unmarshal([]byte(strings.Replace(string(out), `"conditionObject":{`, `"conditionObject":{"__typename":"CollectionRuleMetafieldCondition",`, 1)), &again))
	assert.Equal(t, rule, again)
}

func TestMetaobjectFieldJSONValueShapes(t *testing.T) {
	tests := []struct {
		name      string
		jsonValue string
		want      *string
	}{
		{name: "string", jsonValue: `"Ash"`, want: model.NewString("Ash")},
		{name: "string with comma", jsonValue: `"Ash, Oak"`, want: model.NewString("Ash, Oak")},
		{name: "empty string", jsonValue: `""`, want: model.NewString("")},
		{name: "list of strings", jsonValue: `["gid://shopify/TaxonomyValue/3"]`, want: model.NewString(`["gid://shopify/TaxonomyValue/3"]`)},
		{name: "list of several strings", jsonValue: `["a","b"]`, want: model.NewString(`["a","b"]`)},
		{name: "list of numbers", jsonValue: `[1,2]`, want: model.NewString(`[1,2]`)},
		{name: "nested list", jsonValue: `[["a"]]`, want: model.NewString(`[["a"]]`)},
		{name: "empty list", jsonValue: `[]`, want: model.NewString(`[]`)},
		{name: "integer", jsonValue: `42`, want: model.NewString(`42`)},
		{name: "decimal", jsonValue: `4.5`, want: model.NewString(`4.5`)},
		{name: "boolean", jsonValue: `true`, want: model.NewString(`true`)},
		{name: "object", jsonValue: `{"value":1.5,"unit":"KILOGRAMS"}`, want: model.NewString(`{"value":1.5,"unit":"KILOGRAMS"}`)},
		{name: "null", jsonValue: `null`, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f model.MetaobjectField
			require.NoError(t, json.Unmarshal([]byte(`{"key":"k","type":"t","jsonValue":`+tt.jsonValue+`}`), &f))
			assert.Equal(t, "k", f.Key)
			assert.Equal(t, "t", f.Type)
			if tt.want == nil {
				assert.Nil(t, f.JSONValue)
				return
			}
			require.NotNil(t, f.JSONValue)
			assert.Equal(t, *tt.want, *f.JSONValue)
		})
	}

	t.Run("absent", func(t *testing.T) {
		var f model.MetaobjectField
		require.NoError(t, json.Unmarshal([]byte(`{"key":"k","type":"t"}`), &f))
		assert.Nil(t, f.JSONValue)
	})
}

func TestMetaobjectFieldKeepsAllSelectedFields(t *testing.T) {
	in := `{
		"key": "label",
		"type": "single_line_text_field",
		"value": "Ash",
		"jsonValue": "Ash",
		"definition": {"key": "label", "name": "Label", "required": true, "description": "d"},
		"thumbnail": {"hex": "#ff0000"},
		"references": {"pageInfo": {"hasNextPage": false, "hasPreviousPage": false}}
	}`
	var f model.MetaobjectField
	require.NoError(t, json.Unmarshal([]byte(in), &f))

	assert.Equal(t, "label", f.Key)
	assert.Equal(t, "single_line_text_field", f.Type)
	assert.Equal(t, "Ash", *f.Value)
	assert.Equal(t, "Ash", *f.JSONValue)
	require.NotNil(t, f.Definition)
	assert.Equal(t, "Label", f.Definition.Name)
	assert.True(t, f.Definition.Required)
	require.NotNil(t, f.Thumbnail)
	assert.Equal(t, "#ff0000", *f.Thumbnail.Hex)
	require.NotNil(t, f.References)
	require.NotNil(t, f.References.PageInfo)
	assert.False(t, f.References.PageInfo.HasNextPage)
}

func TestMetaobjectFieldDecodesWhenNested(t *testing.T) {
	var mo model.Metaobject
	require.NoError(t, json.Unmarshal([]byte(`{"id":"gid://shopify/Metaobject/1","type":"shopify--color-pattern","fields":[
		{"key":"label","type":"single_line_text_field","jsonValue":"Ash"},
		{"key":"color_taxonomy_reference","type":"list.product_taxonomy_value_reference","jsonValue":["gid://shopify/TaxonomyValue/3"]}
	]}`), &mo))
	require.Len(t, mo.Fields, 2)
	assert.Equal(t, "Ash", *mo.Fields[0].JSONValue)
	assert.Equal(t, `["gid://shopify/TaxonomyValue/3"]`, *mo.Fields[1].JSONValue)
}

// Selecting reference on a metaobject field is not supported yet: the field
// is a GraphQL union and decoding it must fail rather than drop the value.
func TestMetaobjectFieldReferenceIsRejected(t *testing.T) {
	var f model.MetaobjectField
	err := json.Unmarshal([]byte(`{"key":"product","type":"product_reference","jsonValue":"gid://shopify/Product/1","reference":{"__typename":"Product","id":"gid://shopify/Product/1"}}`), &f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MetafieldReference")
}

func TestMetaobjectFieldMarshalRoundTrip(t *testing.T) {
	var f model.MetaobjectField
	require.NoError(t, json.Unmarshal([]byte(`{"key":"k","type":"t","value":"[1,2]","jsonValue":[1,2],"definition":{"key":"k","name":"K","required":false}}`), &f))

	out, err := json.Marshal(f)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"jsonValue":"[1,2]"`)

	var again model.MetaobjectField
	require.NoError(t, json.Unmarshal(out, &again))
	assert.Equal(t, f, again)
}
