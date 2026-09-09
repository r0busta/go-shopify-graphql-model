package main

import (
	"encoding/json"
	"testing"
)

const fixture = `{
  "directives": [
    {"name": "include", "description": null, "locations": ["FIELD"], "args": []},
    {"name": "accessRestricted", "description": "Restricted.", "locations": ["FIELD_DEFINITION", "OBJECT"],
     "args": [{"name": "reason", "description": "Why.", "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "defaultValue": "null"}]}
  ],
  "types": [
    {"kind": "SCALAR", "name": "String", "description": "Built in."},
    {"kind": "OBJECT", "name": "__Type", "description": null, "fields": []},
    {"kind": "SCALAR", "name": "URL", "description": "An RFC 3986 URI.\nSecond line."},
    {"kind": "INTERFACE", "name": "Node", "description": "Has an id.", "interfaces": [],
     "fields": [{"name": "id", "description": "The id.", "args": [], "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}, "isDeprecated": false, "deprecationReason": null}]},
    {"kind": "OBJECT", "name": "Product", "description": "A product.", "interfaces": [{"kind": "INTERFACE", "name": "Node", "ofType": null}],
     "fields": [
       {"name": "id", "description": null, "args": [], "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}, "isDeprecated": false, "deprecationReason": null},
       {"name": "tags", "description": "Tags.", "args": [{"name": "first", "description": null, "type": {"kind": "SCALAR", "name": "Int", "ofType": null}, "defaultValue": "10"}],
        "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "LIST", "name": null, "ofType": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}}}, "isDeprecated": false, "deprecationReason": null},
       {"name": "images", "description": "Images.", "args": [{"name": "query", "description": "Filter with a trailing newline.\n", "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "defaultValue": null}],
        "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "isDeprecated": true, "deprecationReason": "Use ` + "`media`" + ` instead."},
       {"name": "old", "description": null, "args": [], "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "isDeprecated": true, "deprecationReason": "No longer supported"}
     ]},
    {"kind": "UNION", "name": "Media", "description": null, "possibleTypes": [{"kind": "OBJECT", "name": "Product", "ofType": null}, {"kind": "OBJECT", "name": "Node", "ofType": null}]},
    {"kind": "ENUM", "name": "Status", "description": "Status. It has a description that is longer than seventy characters wide.",
     "enumValues": [{"name": "ACTIVE", "description": "On.", "isDeprecated": false, "deprecationReason": null}, {"name": "DRAFT", "description": null, "isDeprecated": true, "deprecationReason": "Say \"no\"."}]},
    {"kind": "INPUT_OBJECT", "name": "ProductInput", "description": null,
     "inputFields": [{"name": "title", "description": "Title.", "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "defaultValue": "\"x\""}]},
    {"kind": "OBJECT", "name": "Empty", "description": null, "interfaces": [], "fields": []}
  ]
}`

const want = `"""Restricted."""
directive @accessRestricted(
  """Why."""
  reason: String = null
) on FIELD_DEFINITION | OBJECT

"""
An RFC 3986 URI.
Second line.
"""
scalar URL

"""Has an id."""
interface Node {
  """The id."""
  id: ID!
}

"""A product."""
type Product implements Node {
  id: ID!

  """Tags."""
  tags(first: Int = 10): [String!]!

  """Images."""
  images(
    "Filter with a trailing newline.\n"
    query: String
  ): String @deprecated(reason: "Use ` + "`media`" + ` instead.")
  old: String @deprecated
}

union Media = Product | Node

"""
Status. It has a description that is longer than seventy characters wide.
"""
enum Status {
  """On."""
  ACTIVE
  DRAFT @deprecated(reason: "Say \"no\".")
}

input ProductInput {
  """Title."""
  title: String = "x"
}

type Empty
`

func TestPrintSchema(t *testing.T) {
	var s schema
	if err := json.Unmarshal([]byte(fixture), &s); err != nil {
		t.Fatal(err)
	}
	if got := printSchema(&s); got != want {
		t.Errorf("printSchema mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestPrintBlockString(t *testing.T) {
	cases := map[string]string{
		"one line":          `"""one line"""`,
		"two\nlines":        "\"\"\"\ntwo\nlines\n\"\"\"",
		` leading space`:    `""" leading space"""`,
		`ends with "quote"`: "\"\"\"\nends with \"quote\"\n\"\"\"",
		`has """ inside`:    "\"\"\"has \\\"\"\" inside\"\"\"",
		"first\n  indented": "\"\"\"\nfirst\n  indented\n\"\"\"",
	}
	for in, out := range cases {
		if got := printBlockString(in); got != out {
			t.Errorf("printBlockString(%q) = %q, want %q", in, got, out)
		}
	}
}
