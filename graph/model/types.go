package model

import (
	"fmt"
	"reflect"
	"regexp"

	"gopkg.in/guregu/null.v4"
)

var gidRegex *regexp.Regexp

func init() {
	gidRegex = regexp.MustCompile(`^gid://shopify/(\w+)/\d+$`)
}

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

func concludeObjectType(gid string) (reflect.Type, error) {
	submatches := gidRegex.FindStringSubmatch(gid)
	if len(submatches) != 2 {
		return reflect.TypeOf(nil), fmt.Errorf("malformed gid=`%s`", gid)
	}
	resource := submatches[1]
	switch resource {
	case "MediaImage":
		return reflect.TypeOf(MediaImage{}), nil
	case "Video":
		return reflect.TypeOf(Video{}), nil
	case "Model3d":
		return reflect.TypeOf(Model3d{}), nil
	case "ExternalVideo":
		return reflect.TypeOf(ExternalVideo{}), nil
	default:
		return reflect.TypeOf(nil), fmt.Errorf("`%s` not implemented type", resource)
	}
}
