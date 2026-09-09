package model

import (
	"reflect"
	"regexp"

	"gopkg.in/guregu/null.v4"
)

func NewNullString(v null.String) *null.String {
	return &v
}

func NewString(v string) *string {
	return &v
}

func NewBool(v bool) *bool {
	return &v
}

func NewInt(v int) *int {
	return &v
}

func NewFloat64(v float64) *float64 {
	return &v
}

// gidResource extracts the resource name from a Shopify global ID such as
// gid://shopify/MediaImage/123.
var gidResource = regexp.MustCompile(`^gid://shopify/(\w+)/`)

// mediaTypes maps GraphQL type names to the Go types implementing Media.
var mediaTypes = map[string]reflect.Type{
	"ExternalVideo": reflect.TypeOf(ExternalVideo{}),
	"MediaImage":    reflect.TypeOf(MediaImage{}),
	"Model3d":       reflect.TypeOf(Model3d{}),
	"Video":         reflect.TypeOf(Video{}),
}

// collectionRuleConditionObjectTypes maps GraphQL type names to the Go types
// that are members of the CollectionRuleConditionObject union.
var collectionRuleConditionObjectTypes = map[string]reflect.Type{
	"CollectionRuleCategoryCondition":        reflect.TypeOf(CollectionRuleCategoryCondition{}),
	"CollectionRuleMetafieldCondition":       reflect.TypeOf(CollectionRuleMetafieldCondition{}),
	"CollectionRuleProductCategoryCondition": reflect.TypeOf(CollectionRuleProductCategoryCondition{}),
	"CollectionRuleTextCondition":            reflect.TypeOf(CollectionRuleTextCondition{}),
}
