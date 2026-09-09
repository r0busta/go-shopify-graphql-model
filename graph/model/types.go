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

// collectionSourceTypes maps GraphQL type names to the Go types implementing
// CollectionSource.
var collectionSourceTypes = map[string]reflect.Type{
	"CollectionConditionsSource":     reflect.TypeOf(CollectionConditionsSource{}),
	"CollectionSubCollectionsSource": reflect.TypeOf(CollectionSubCollectionsSource{}),
}

// collectionSourceInclusionConditionTypes maps GraphQL type names to the Go types implementing
// CollectionSourceInclusionCondition.
var collectionSourceInclusionConditionTypes = map[string]reflect.Type{
	"CollectionSourceInclusionConditionMetafieldBoolean":        reflect.TypeOf(CollectionSourceInclusionConditionMetafieldBoolean{}),
	"CollectionSourceInclusionConditionMetafieldDecimal":        reflect.TypeOf(CollectionSourceInclusionConditionMetafieldDecimal{}),
	"CollectionSourceInclusionConditionMetafieldInteger":        reflect.TypeOf(CollectionSourceInclusionConditionMetafieldInteger{}),
	"CollectionSourceInclusionConditionMetafieldMetaobject":     reflect.TypeOf(CollectionSourceInclusionConditionMetafieldMetaobject{}),
	"CollectionSourceInclusionConditionMetafieldMetaobjectList": reflect.TypeOf(CollectionSourceInclusionConditionMetafieldMetaobjectList{}),
	"CollectionSourceInclusionConditionMetafieldString":         reflect.TypeOf(CollectionSourceInclusionConditionMetafieldString{}),
	"CollectionSourceInclusionConditionMetafieldStringList":     reflect.TypeOf(CollectionSourceInclusionConditionMetafieldStringList{}),
	"CollectionSourceInclusionConditionProductCategory":         reflect.TypeOf(CollectionSourceInclusionConditionProductCategory{}),
	"CollectionSourceInclusionConditionProductStatus":           reflect.TypeOf(CollectionSourceInclusionConditionProductStatus{}),
	"CollectionSourceInclusionConditionProductTag":              reflect.TypeOf(CollectionSourceInclusionConditionProductTag{}),
	"CollectionSourceInclusionConditionProductTitle":            reflect.TypeOf(CollectionSourceInclusionConditionProductTitle{}),
	"CollectionSourceInclusionConditionProductType":             reflect.TypeOf(CollectionSourceInclusionConditionProductType{}),
	"CollectionSourceInclusionConditionProductVendor":           reflect.TypeOf(CollectionSourceInclusionConditionProductVendor{}),
	"CollectionSourceInclusionConditionUnknown":                 reflect.TypeOf(CollectionSourceInclusionConditionUnknown{}),
	"CollectionSourceInclusionConditionVariantCompareAtPrice":   reflect.TypeOf(CollectionSourceInclusionConditionVariantCompareAtPrice{}),
	"CollectionSourceInclusionConditionVariantInventory":        reflect.TypeOf(CollectionSourceInclusionConditionVariantInventory{}),
	"CollectionSourceInclusionConditionVariantPrice":            reflect.TypeOf(CollectionSourceInclusionConditionVariantPrice{}),
	"CollectionSourceInclusionConditionVariantTitle":            reflect.TypeOf(CollectionSourceInclusionConditionVariantTitle{}),
	"CollectionSourceInclusionConditionVariantWeight":           reflect.TypeOf(CollectionSourceInclusionConditionVariantWeight{}),
}

// collectionSourceExclusionConditionTypes maps GraphQL type names to the Go types implementing
// CollectionSourceExclusionCondition.
var collectionSourceExclusionConditionTypes = map[string]reflect.Type{
	"CollectionSourceExclusionConditionCollection":      reflect.TypeOf(CollectionSourceExclusionConditionCollection{}),
	"CollectionSourceExclusionConditionProductCategory": reflect.TypeOf(CollectionSourceExclusionConditionProductCategory{}),
	"CollectionSourceExclusionConditionProductTag":      reflect.TypeOf(CollectionSourceExclusionConditionProductTag{}),
	"CollectionSourceExclusionConditionProductType":     reflect.TypeOf(CollectionSourceExclusionConditionProductType{}),
	"CollectionSourceExclusionConditionProductVendor":   reflect.TypeOf(CollectionSourceExclusionConditionProductVendor{}),
	"CollectionSourceExclusionConditionUnknown":         reflect.TypeOf(CollectionSourceExclusionConditionUnknown{}),
}
