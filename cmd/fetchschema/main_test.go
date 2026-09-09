package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
       {"name": "images", "description": "Images.", "args": [
          {"name": "query", "description": "Filter with a trailing newline.\n", "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "defaultValue": null},
          {"name": "maxWidth", "description": "Old.", "type": {"kind": "SCALAR", "name": "Int", "ofType": null}, "defaultValue": null, "isDeprecated": true, "deprecationReason": "Use ` + "`transform`" + ` instead."}],
        "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "isDeprecated": true, "deprecationReason": "Use ` + "`media`" + ` instead."},
       {"name": "old", "description": null, "args": [], "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "isDeprecated": true, "deprecationReason": "No longer supported"}
     ]},
    {"kind": "UNION", "name": "Media", "description": null, "possibleTypes": [{"kind": "OBJECT", "name": "Product", "ofType": null}, {"kind": "OBJECT", "name": "Node", "ofType": null}]},
    {"kind": "ENUM", "name": "Status", "description": "Status. It has a description that is longer than seventy characters wide.",
     "enumValues": [{"name": "ACTIVE", "description": "On.", "isDeprecated": false, "deprecationReason": null}, {"name": "DRAFT", "description": null, "isDeprecated": true, "deprecationReason": "Say \"no\"."}]},
    {"kind": "INPUT_OBJECT", "name": "ProductInput", "description": null,
     "inputFields": [
       {"name": "title", "description": "Title.", "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "defaultValue": "\"x\""},
       {"name": "publications", "description": null, "type": {"kind": "LIST", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}, "defaultValue": null, "isDeprecated": true, "deprecationReason": "Use publishablePublish instead."}
     ]},
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

    """Old."""
    maxWidth: Int @deprecated(reason: "Use ` + "`transform`" + ` instead.")
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
  publications: [ID] @deprecated(reason: "Use publishablePublish instead.")
}

type Empty
`

func TestPrintSchema(t *testing.T) {
	var s schema
	if err := json.Unmarshal([]byte(fixture), &s); err != nil {
		t.Fatal(err)
	}
	got, err := render(&s)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("printSchema mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestPrintBlockString(t *testing.T) {
	cases := map[string]string{
		"one line":          `"""one line"""`,
		"two\nlines":        "\"\"\"\ntwo\nlines\n\"\"\"",
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

func TestRenderRejectsMalformedResults(t *testing.T) {
	cases := map[string]string{
		"type reference too deep":   `{"types":[{"kind":"OBJECT","name":"T","fields":[{"name":"f","args":[],"type":{"kind":"NON_NULL","name":null,"ofType":null}}]}]}`,
		"union member without name": `{"types":[{"kind":"UNION","name":"U","possibleTypes":[{"kind":"OBJECT"}]}]}`,
		"unknown kind":              `{"types":[{"name":"X"}]}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			var s schema
			if err := json.Unmarshal([]byte(in), &s); err != nil {
				t.Fatal(err)
			}
			if _, err := render(&s); err == nil || !strings.Contains(err.Error(), "malformed introspection result") {
				t.Errorf("render error = %v, want a malformed-result error", err)
			}
		})
	}
}

func TestIntrospectErrors(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		want   string
	}{
		"http error, body truncated": {http.StatusBadGateway, "<html>" + strings.Repeat("x", 500), "HTTP 502: <html>xxx"},
		"graphql errors":             {http.StatusOK, `{"errors":[{"message":"Invalid API version"},{"message":"and more"}]}`, "introspection failed: Invalid API version; and more"},
		"errors without messages":    {http.StatusOK, `{"errors":[{"extensions":{"code":"THROTTLED"}}]}`, `introspection failed: {"errors":[{"extensions":{"code":"THROTTLED"}}]}`},
		"not json":                   {http.StatusOK, "<html>login</html>", "response is not JSON"},
		"no schema":                  {http.StatusOK, `{"data":{"__schema":{"types":[]}}}`, "response has no schema types"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			_, err := introspect(srv.Client(), srv.URL, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("introspect error = %v, want it to contain %q", err, tc.want)
			}
			if err != nil && len(err.Error()) > 400 {
				t.Errorf("error message is %d bytes long; bodies must be truncated", len(err.Error()))
			}
		})
	}
}

func TestIntrospectTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	client := srv.Client()
	client.Timeout = 20 * time.Millisecond
	if _, err := introspect(client, srv.URL, nil); err == nil {
		t.Error("expected a timeout error")
	}
}

func TestIntrospectSendsHeadersAndQuery(t *testing.T) {
	var gotToken, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Shopify-Access-Token")
		var payload struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		gotQuery = payload.Query
		_, _ = w.Write([]byte(`{"data":{"__schema":` + fixture + `}}`))
	}))
	defer srv.Close()

	s, err := introspect(srv.Client(), srv.URL, map[string]string{"X-Shopify-Access-Token": "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if gotToken != "secret" {
		t.Errorf("token header = %q", gotToken)
	}
	for _, want := range []string{"args(includeDeprecated: true)", "inputFields(includeDeprecated: true)", "fields(includeDeprecated: true)"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("introspection query lacks %q", want)
		}
	}
	if len(s.Types) == 0 {
		t.Error("no types decoded")
	}
}
