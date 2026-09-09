package main_test

import (
	"encoding/json"
	"testing"

	"github.com/r0busta/go-shopify-graphql-model/v5/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const collectionWithSourcesJSON = `{
  "id": "gid://shopify/Collection/1",
  "handle": "summer",
  "title": "Summer",
  "sources": [
    {"__typename": "CollectionConditionsSource", "id": "gid://shopify/CollectionConditionsSource/1", "title": "Rules", "shareable": false, "targetType": "PRODUCTS",
     "inclusion": {"matchType": "ALL", "conditions": [
       {"__typename": "CollectionSourceInclusionConditionProductTag", "id": "gid://shopify/CollectionSourceInclusionCondition/1", "matchType": "ANY", "relation": "EQUALS", "values": ["summer"]},
       {"__typename": "CollectionSourceInclusionConditionMetafieldBoolean", "id": "gid://shopify/CollectionSourceInclusionCondition/2", "relation": "EQUALS", "value": true, "definition": {"id": "gid://shopify/MetafieldDefinition/1", "key": "featured", "namespace": "custom", "name": "Featured", "ownerType": "PRODUCT"}}
     ]},
     "exclusion": {"matchType": "ANY", "conditions": [
       {"__typename": "CollectionSourceExclusionConditionProductVendor", "id": "gid://shopify/CollectionSourceExclusionCondition/3", "matchType": "ANY", "relation": "EQUALS", "values": ["Acme"]}
     ]}},
    {"__typename": "CollectionSubCollectionsSource", "id": "gid://shopify/CollectionSubCollectionsSource/2", "title": "Children", "description": "Nested",
     "collections": [{"id": "gid://shopify/Collection/2", "handle": "hats", "sources": []}]}
  ],
  "ruleSet": {"appliedDisjunctively": false, "rules": [{"column": "TAG", "condition": "summer", "relation": "EQUALS"}]}
}`

func TestCollectionSourcesDecoding(t *testing.T) {
	var c model.Collection
	require.NoError(t, json.Unmarshal([]byte(collectionWithSourcesJSON), &c))

	assert.Equal(t, "summer", c.Handle)
	require.NotNil(t, c.RuleSet, "sibling fields must keep decoding")
	require.Len(t, c.RuleSet.Rules, 1)
	require.Len(t, c.Sources, 2)

	cond, ok := c.Sources[0].(*model.CollectionConditionsSource)
	require.True(t, ok, "source 0 is %T", c.Sources[0])
	assert.Equal(t, "Rules", cond.Title)
	assert.Equal(t, model.CollectionSourceTargetTypeProducts, cond.TargetType)
	require.NotNil(t, cond.Inclusion)
	require.Len(t, cond.Inclusion.Conditions, 2)
	tag, ok := cond.Inclusion.Conditions[0].(*model.CollectionSourceInclusionConditionProductTag)
	require.True(t, ok, "condition 0 is %T", cond.Inclusion.Conditions[0])
	assert.Equal(t, []string{"summer"}, tag.Values)
	boolean, ok := cond.Inclusion.Conditions[1].(*model.CollectionSourceInclusionConditionMetafieldBoolean)
	require.True(t, ok, "condition 1 is %T", cond.Inclusion.Conditions[1])
	assert.True(t, boolean.Value)
	require.NotNil(t, boolean.Definition)
	assert.Equal(t, "featured", boolean.Definition.Key)
	require.NotNil(t, cond.Exclusion)
	require.Len(t, cond.Exclusion.Conditions, 1)
	vendor, ok := cond.Exclusion.Conditions[0].(*model.CollectionSourceExclusionConditionProductVendor)
	require.True(t, ok, "exclusion 0 is %T", cond.Exclusion.Conditions[0])
	assert.Equal(t, []string{"Acme"}, vendor.Values)

	sub, ok := c.Sources[1].(*model.CollectionSubCollectionsSource)
	require.True(t, ok, "source 1 is %T", c.Sources[1])
	assert.Equal(t, "Nested", *sub.Description)
	require.Len(t, sub.Collections, 1)
	assert.Equal(t, "hats", sub.Collections[0].Handle)
	assert.Empty(t, sub.Collections[0].Sources)

	// Sources are reached through the interface's getters too.
	assert.Equal(t, "Children", c.Sources[1].GetTitle())
}

func TestCollectionSourcesEdgeCases(t *testing.T) {
	cases := map[string]struct {
		in      string
		wantLen int
		wantNil bool
		wantErr string
	}{
		"not selected":               {in: `{"id":"gid://shopify/Collection/1"}`, wantNil: true},
		"null":                       {in: `{"id":"gid://shopify/Collection/1","sources":null}`, wantNil: true},
		"empty":                      {in: `{"id":"gid://shopify/Collection/1","sources":[]}`, wantLen: 0},
		"by id only":                 {in: `{"id":"gid://shopify/Collection/1","sources":[{"id":"gid://shopify/CollectionConditionsSource/1"}]}`, wantLen: 1},
		"no typename":                {in: `{"id":"gid://shopify/Collection/1","sources":[{"title":"x"}]}`, wantErr: "must select __typename"},
		"unknown type":               {in: `{"id":"gid://shopify/Collection/1","sources":[{"__typename":"CollectionOtherSource"}]}`, wantErr: "unsupported type"},
		"condition without typename": {in: `{"id":"gid://shopify/Collection/1","sources":[{"__typename":"CollectionConditionsSource","inclusion":{"conditions":[{"relation":"EQUALS"}]}}]}`, wantErr: "decode CollectionSourceInclusionCondition: query must select __typename"},
		"null and empty conditions":  {in: `{"id":"gid://shopify/Collection/1","sources":[{"__typename":"CollectionConditionsSource","inclusion":{"conditions":[]},"exclusion":{"conditions":null}}]}`, wantLen: 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var c model.Collection
			err := json.Unmarshal([]byte(tc.in), &c)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, c.Sources)
			} else {
				assert.Len(t, c.Sources, tc.wantLen)
			}
		})
	}
}

func TestCollectionSourcesNested(t *testing.T) {
	var conn model.CollectionConnection
	err := json.Unmarshal([]byte(`{"edges":[{"node":`+collectionWithSourcesJSON+`}],"nodes":[`+collectionWithSourcesJSON+`]}`), &conn)
	require.NoError(t, err)
	require.Len(t, conn.Edges, 1)
	require.Len(t, conn.Edges[0].Node.Sources, 2)
	require.Len(t, conn.Nodes, 1)
	require.Len(t, conn.Nodes[0].Sources, 2)
}

func TestCollectionSourcesMarshalRoundTrip(t *testing.T) {
	var c model.Collection
	require.NoError(t, json.Unmarshal([]byte(collectionWithSourcesJSON), &c))
	out, err := json.Marshal(c)
	require.NoError(t, err)

	// __typename is not part of the generated structs, so a round trip relies
	// on the ids. Condition ids share one resource name, so their concrete
	// types cannot be recovered and the decode must say so.
	var again model.Collection
	require.ErrorContains(t, json.Unmarshal(out, &again), "unsupported type")

	c.Sources[0].(*model.CollectionConditionsSource).Inclusion = nil
	c.Sources[0].(*model.CollectionConditionsSource).Exclusion = nil
	out, err = json.Marshal(c)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(out, &again))
	require.Len(t, again.Sources, 2)
	_, ok := again.Sources[0].(*model.CollectionConditionsSource)
	assert.True(t, ok)
}
