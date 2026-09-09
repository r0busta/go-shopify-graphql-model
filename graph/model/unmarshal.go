package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// gqlgen generates Go interfaces for GraphQL interfaces and unions, and
// encoding/json cannot decode into an interface value on its own. The
// UnmarshalJSON methods in this file pick the concrete type from the response
// and decode into it. Each method embeds an alias of the generated struct so
// every generated field keeps decoding as usual, and only the interface-typed
// fields are handled by hand.

var jsonNull = []byte("null")

func isJSONNull(raw json.RawMessage) bool {
	return len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), jsonNull)
}

// decodeByTypename decodes raw into a new value of the type registered under
// the object's __typename. When __typename was not selected, the resource
// name of the object's id is used instead. The returned value is a pointer to
// the concrete type.
func decodeByTypename(raw json.RawMessage, types map[string]reflect.Type, kind string) (interface{}, error) {
	var probe struct {
		Typename string `json:"__typename"`
		ID       string `json:"id"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("decode %s: %w", kind, err)
	}

	name := probe.Typename
	if name == "" {
		if m := gidResource.FindStringSubmatch(probe.ID); m != nil {
			name = m[1]
		}
	}
	if name == "" {
		return nil, fmt.Errorf("decode %s: query must select __typename or id", kind)
	}

	t, ok := types[name]
	if !ok {
		return nil, fmt.Errorf("decode %s: unsupported type %q", kind, name)
	}

	v := reflect.New(t).Interface()
	if err := json.Unmarshal(raw, v); err != nil {
		return nil, fmt.Errorf("decode %s %s: %w", kind, name, err)
	}

	return v, nil
}

func decodeMedia(raw json.RawMessage) (Media, error) {
	if isJSONNull(raw) {
		return nil, nil
	}
	v, err := decodeByTypename(raw, mediaTypes, "Media")
	if err != nil {
		return nil, err
	}
	return v.(Media), nil
}

func decodeConditionObject(raw json.RawMessage) (CollectionRuleConditionObject, error) {
	if isJSONNull(raw) {
		return nil, nil
	}
	v, err := decodeByTypename(raw, collectionRuleConditionObjectTypes, "CollectionRuleConditionObject")
	if err != nil {
		return nil, err
	}
	return v.(CollectionRuleConditionObject), nil
}

// UnmarshalJSON decodes the Media node into its concrete type. The query must
// select __typename or id on the node.
func (s *MediaEdge) UnmarshalJSON(b []byte) error {
	type alias MediaEdge
	tmp := struct {
		*alias
		Node json.RawMessage `json:"node"`
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	node, err := decodeMedia(tmp.Node)
	if err != nil {
		return err
	}
	s.Node = node

	return nil
}

// UnmarshalJSON decodes the Media nodes into their concrete types. The query
// must select __typename or id on each node.
func (s *MediaConnection) UnmarshalJSON(b []byte) error {
	type alias MediaConnection
	tmp := struct {
		*alias
		Nodes []json.RawMessage `json:"nodes"`
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	s.Nodes = nil
	if tmp.Nodes != nil {
		s.Nodes = make([]Media, len(tmp.Nodes))
		for i, raw := range tmp.Nodes {
			node, err := decodeMedia(raw)
			if err != nil {
				return err
			}
			s.Nodes[i] = node
		}
	}

	return nil
}

// UnmarshalJSON decodes conditionObject into its concrete type. The query
// must select __typename on conditionObject.
func (s *CollectionRule) UnmarshalJSON(b []byte) error {
	type alias CollectionRule
	tmp := struct {
		*alias
		ConditionObject json.RawMessage `json:"conditionObject"`
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	obj, err := decodeConditionObject(tmp.ConditionObject)
	if err != nil {
		return err
	}
	s.ConditionObject = obj

	return nil
}

// UnmarshalJSON accepts any JSON value in jsonValue. A JSON string is stored
// unquoted; any other value (list, number, boolean, object) is stored as its
// JSON text, which is the same representation the value field uses.
func (s *MetaobjectField) UnmarshalJSON(b []byte) error {
	type alias MetaobjectField
	tmp := struct {
		*alias
		JSONValue json.RawMessage `json:"jsonValue"`
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	s.JSONValue = nil
	if !isJSONNull(tmp.JSONValue) {
		var str string
		if err := json.Unmarshal(tmp.JSONValue, &str); err == nil {
			s.JSONValue = &str
		} else {
			raw := string(bytes.TrimSpace(tmp.JSONValue))
			s.JSONValue = &raw
		}
	}

	return nil
}
