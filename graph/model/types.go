package model

import "gopkg.in/guregu/null.v4"

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
