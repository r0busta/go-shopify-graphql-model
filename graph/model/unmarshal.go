package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func (s *MediaEdge) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	if cursor, ok := m["cursor"].(string); ok {
		s.Cursor = cursor
	}
	if node, ok := m["node"].(map[string]interface{}); ok {
		s.Node, err = decodeMedia(node)
		if err != nil {
			return fmt.Errorf("decode media node: %w", err)
		}
	}
	return nil
}

func (s *MediaConnection) UnmarshalJSON(b []byte) error {
	var (
		m     map[string]interface{}
		mConn struct {
			Edges    []MediaEdge `json:"edges,omitempty"`
			PageInfo *PageInfo   `json:"pageInfo,omitempty"`
		}
	)
	err := json.Unmarshal(b, &mConn)
	if err != nil {
		return err
	}
	s.Edges = mConn.Edges
	s.PageInfo = mConn.PageInfo

	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	if nodes, ok := m["nodes"].([]interface{}); ok {
		s.Nodes = make([]Media, len(nodes))
		for i, n := range nodes {
			if node, ok := n.(map[string]interface{}); ok {
				s.Nodes[i], err = decodeMedia(node)
				if err != nil {
					return fmt.Errorf("decode media node: %w", err)
				}
			} else {
				return fmt.Errorf("expected type map[string]interface{} for Media node, got %T", n)
			}
		}
	}
	return nil
}

func decodeMedia(node map[string]interface{}) (Media, error) {
	if id, ok := node["id"].(string); ok {
		mediaType, err := concludeMediaObjectType(id)
		if err != nil {
			return nil, fmt.Errorf("conclude object type: %w", err)
		}
		media := reflect.New(mediaType).Interface()
		err = mapstructure.Decode(node, media)
		if err != nil {
			return nil, fmt.Errorf("decode media node: %w", err)
		}
		return media.(Media), nil
	}
	return nil, fmt.Errorf("must query id to decode Media")
}

func (s *CollectionRule) UnmarshalJSON(b []byte) error {
	// ConditionObject is CollectionRuleConditionObject, which is an interface

	var cr struct {
		Column          CollectionRuleColumn   `json:"column"`
		Condition       string                 `json:"condition"`
		ConditionObject map[string]interface{} `json:"conditionObject,omitempty"`
		Relation        CollectionRuleRelation `json:"relation"`
	}
	err := json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// assign files from temp object
	s.Column = cr.Column
	s.Condition = cr.Condition
	s.Relation = cr.Relation

	// fiddle with things to get the ConditionObject value
	if len(cr.ConditionObject) > 0 {
		s.ConditionObject, err = decodeConditionObject(cr.ConditionObject)
		if err != nil {
			return fmt.Errorf("decode condition object: %w", err)
		}
	}
	return nil
}

func decodeConditionObject(condObj map[string]interface{}) (CollectionRuleConditionObject, error) {
	if typeName, ok := condObj["__typename"].(string); ok {
		conditionType, err := concludeConditionObjectType(typeName)
		if err != nil {
			return nil, fmt.Errorf("conclude object type: %w", err)
		}
		conditionObject := reflect.New(conditionType).Interface()
		err = mapstructure.Decode(condObj, conditionObject)
		if err != nil {
			return nil, fmt.Errorf("decode condition object: %w", err)
		}
		return conditionObject.(CollectionRuleConditionObject), nil
	}
	return nil, fmt.Errorf("must query __typename to decode condition object")
}

// sometimes the JSONValue value for MetaobjectField's are []string rather than the expected string
// seen where Metaobject.Type == "shopify--color-pattern" for
func (s *MetaobjectField) UnmarshalJSON(b []byte) error {

	var mof struct {
		JSONValue  interface{}                   `json:"jsonValue,omitempty"`
		Key        string                        `json:"key"`
		Reference  MetafieldReference            `json:"reference,omitempty"`
		References *MetafieldReferenceConnection `json:"references,omitempty"`
		Thumbnail  *MetaobjectThumbnail          `json:"thumbnail,omitempty"`
		Type       string                        `json:"type"`
		Value      *string                       `json:"value,omitempty"`
	}
	err := json.Unmarshal(b, &mof)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// assign files from temp object
	s.Key = mof.Key
	s.Reference = mof.Reference
	s.References = mof.References
	s.Thumbnail = mof.Thumbnail
	s.Type = mof.Type
	s.Value = mof.Value

	// deal with the variable JSONValue type
	switch t := mof.JSONValue.(type) {
	case nil:
		// do nothing / save nothing
	case string:
		s.JSONValue = &t
	case []interface{}:
		if len(t) > 0 {
			jvStr := t[0].(string)
			for _, v := range t[1:] {
				jvStr += ", " + v.(string)
			}
			s.JSONValue = &jvStr
		}
	default:
		return errors.New(fmt.Sprintf("Unknown Type for JSONValue: %T", t))
	}

	return nil
}
