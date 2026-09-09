// Command fetchschema downloads the Shopify Admin API GraphQL schema for one
// API version and writes it as SDL, ready for the model generator.
//
// By default it uses the public introspection endpoint on shopify.dev, which
// needs no store or access token:
//
//	go run ./cmd/fetchschema -version 2026-07
//
// To introspect a specific store instead, pass -store and set ACCESS_TOKEN:
//
//	ACCESS_TOKEN=shpat_... go run ./cmd/fetchschema -version 2026-07 -store my-store
//
// The output follows graphql-js 16.7's printSchema so that schema diffs
// between API versions stay readable; for 2026-07 it is byte-identical apart
// from a trailing newline. Unlike graphql-js's default introspection query,
// deprecated arguments and input fields are included, since the API still
// accepts them. The `schema { ... }` block is omitted because the generator
// does not need it.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const proxyURL = "https://shopify.dev/admin-graphql-direct-proxy/"

var (
	versionPattern = regexp.MustCompile(`^(\d{4}-\d{2}|unstable)$`)
	storePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

func main() {
	version := flag.String("version", "", "Admin API version, for example 2026-07 (required)")
	store := flag.String("store", "", "myshopify store name; introspects that store using the ACCESS_TOKEN environment variable instead of the public proxy")
	out := flag.String("o", "schema.graphql", "output file, or - for stdout")
	timeout := flag.Duration("timeout", 2*time.Minute, "HTTP timeout")
	flag.Parse()

	if !versionPattern.MatchString(*version) {
		fmt.Fprintf(os.Stderr, "fetchschema: -version must look like 2026-07 or be \"unstable\", got %q\n", *version)
		flag.Usage()
		os.Exit(2)
	}

	url := proxyURL + *version
	headers := map[string]string{"Content-Type": "application/json"}
	if *store != "" {
		if !storePattern.MatchString(*store) {
			fmt.Fprintf(os.Stderr, "fetchschema: -store must be the store's myshopify subdomain, got %q\n", *store)
			os.Exit(2)
		}
		token := os.Getenv("ACCESS_TOKEN")
		if token == "" {
			fmt.Fprintln(os.Stderr, "fetchschema: ACCESS_TOKEN must be set when -store is given")
			os.Exit(2)
		}
		url = fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/graphql.json", *store, *version)
		headers["X-Shopify-Access-Token"] = token
	}

	client := &http.Client{Timeout: *timeout}
	schema, err := introspect(client, url, headers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetchschema:", err)
		os.Exit(1)
	}

	sdl, err := render(schema)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetchschema:", err)
		os.Exit(1)
	}

	if *out == "-" {
		_, err = os.Stdout.WriteString(sdl)
	} else {
		err = os.WriteFile(*out, []byte(sdl), 0o644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetchschema:", err)
		os.Exit(1)
	}
}

func introspect(client *http.Client, url string, headers map[string]string) (*schema, error) {
	body, err := json.Marshal(map[string]string{"query": introspectionQuery})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d: %s", url, resp.StatusCode, excerpt(raw))
	}

	var result struct {
		Data struct {
			Schema *schema `json:"__schema"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("%s: response is not JSON (%v): %s", url, err, excerpt(raw))
	}
	if len(result.Errors) > 0 {
		var msgs []string
		for _, e := range result.Errors {
			if e.Message != "" {
				msgs = append(msgs, e.Message)
			}
		}
		if len(msgs) == 0 {
			return nil, fmt.Errorf("introspection failed: %s", excerpt(raw))
		}
		return nil, fmt.Errorf("introspection failed: %s", strings.Join(msgs, "; "))
	}
	if result.Data.Schema == nil || len(result.Data.Schema.Types) == 0 {
		return nil, fmt.Errorf("response has no schema types: %s", excerpt(raw))
	}

	return result.Data.Schema, nil
}

// excerpt returns the start of a response body for error messages.
func excerpt(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

// render prints the schema, turning malformed introspection data into an
// error instead of a panic.
func render(s *schema) (sdl string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("malformed introspection result: %v", r)
		}
	}()
	return printSchema(s), nil
}

// Introspection result types, following the GraphQL introspection schema.

type schema struct {
	Types      []typeDef      `json:"types"`
	Directives []directiveDef `json:"directives"`
}

type typeDef struct {
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	Description   *string      `json:"description"`
	Fields        []fieldDef   `json:"fields"`
	InputFields   []inputValue `json:"inputFields"`
	Interfaces    []typeRef    `json:"interfaces"`
	EnumValues    []enumValue  `json:"enumValues"`
	PossibleTypes []typeRef    `json:"possibleTypes"`
}

type fieldDef struct {
	Name              string       `json:"name"`
	Description       *string      `json:"description"`
	Args              []inputValue `json:"args"`
	Type              typeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason *string      `json:"deprecationReason"`
}

type inputValue struct {
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	Type              typeRef `json:"type"`
	DefaultValue      *string `json:"defaultValue"`
	IsDeprecated      bool    `json:"isDeprecated"`
	DeprecationReason *string `json:"deprecationReason"`
}

type enumValue struct {
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	IsDeprecated      bool    `json:"isDeprecated"`
	DeprecationReason *string `json:"deprecationReason"`
}

type typeRef struct {
	Kind   string   `json:"kind"`
	Name   *string  `json:"name"`
	OfType *typeRef `json:"ofType"`
}

var errTypeRefTooDeep = errors.New("type reference nested deeper than the introspection query supports")

func (t typeRef) String() string {
	switch t.Kind {
	case "NON_NULL":
		if t.OfType == nil {
			panic(errTypeRefTooDeep)
		}
		return t.OfType.String() + "!"
	case "LIST":
		if t.OfType == nil {
			panic(errTypeRefTooDeep)
		}
		return "[" + t.OfType.String() + "]"
	default:
		return t.name()
	}
}

func (t typeRef) name() string {
	if t.Name == nil {
		panic(fmt.Sprintf("type reference of kind %q without a name", t.Kind))
	}
	return *t.Name
}

type directiveDef struct {
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	Locations   []string     `json:"locations"`
	Args        []inputValue `json:"args"`
}

// Printing, modelled on graphql-js printSchema.

var specifiedDirectives = map[string]bool{"include": true, "skip": true, "deprecated": true, "specifiedBy": true}
var specifiedScalars = map[string]bool{"String": true, "Int": true, "Float": true, "Boolean": true, "ID": true}

func printSchema(s *schema) string {
	var blocks []string

	for _, d := range s.Directives {
		if specifiedDirectives[d.Name] {
			continue
		}
		blocks = append(blocks, printDirective(d))
	}
	for _, t := range s.Types {
		if strings.HasPrefix(t.Name, "__") || specifiedScalars[t.Name] {
			continue
		}
		blocks = append(blocks, printType(t))
	}

	return strings.Join(blocks, "\n\n") + "\n"
}

func printDirective(d directiveDef) string {
	return printDescription(d.Description, "", true) +
		"directive @" + d.Name + printArgs(d.Args, "") +
		" on " + strings.Join(d.Locations, " | ")
}

func printType(t typeDef) string {
	desc := printDescription(t.Description, "", true)
	switch t.Kind {
	case "SCALAR":
		return desc + "scalar " + t.Name
	case "OBJECT":
		return desc + "type " + t.Name + printImplements(t.Interfaces) + printFields(t.Fields)
	case "INTERFACE":
		return desc + "interface " + t.Name + printImplements(t.Interfaces) + printFields(t.Fields)
	case "UNION":
		names := make([]string, len(t.PossibleTypes))
		for i, p := range t.PossibleTypes {
			names[i] = p.name()
		}
		return desc + "union " + t.Name + " = " + strings.Join(names, " | ")
	case "ENUM":
		items := make([]string, len(t.EnumValues))
		for i, v := range t.EnumValues {
			items[i] = printDescription(v.Description, "  ", i == 0) + "  " + v.Name + printDeprecated(v.IsDeprecated, v.DeprecationReason)
		}
		return desc + "enum " + t.Name + printBlock(items)
	case "INPUT_OBJECT":
		items := make([]string, len(t.InputFields))
		for i, f := range t.InputFields {
			items[i] = printDescription(f.Description, "  ", i == 0) + "  " + printInputValue(f)
		}
		return desc + "input " + t.Name + printBlock(items)
	default:
		panic(fmt.Sprintf("type %q has unexpected kind %q", t.Name, t.Kind))
	}
}

func printImplements(interfaces []typeRef) string {
	if len(interfaces) == 0 {
		return ""
	}
	names := make([]string, len(interfaces))
	for i, r := range interfaces {
		names[i] = r.name()
	}
	return " implements " + strings.Join(names, " & ")
}

func printFields(fields []fieldDef) string {
	items := make([]string, len(fields))
	for i, f := range fields {
		items[i] = printDescription(f.Description, "  ", i == 0) + "  " + f.Name + printArgs(f.Args, "  ") +
			": " + f.Type.String() + printDeprecated(f.IsDeprecated, f.DeprecationReason)
	}
	return printBlock(items)
}

func printBlock(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return " {\n" + strings.Join(items, "\n") + "\n}"
}

func printArgs(args []inputValue, indentation string) string {
	if len(args) == 0 {
		return ""
	}

	described := false
	for _, a := range args {
		if a.Description != nil && *a.Description != "" {
			described = true
			break
		}
	}
	if !described {
		items := make([]string, len(args))
		for i, a := range args {
			items[i] = printInputValue(a)
		}
		return "(" + strings.Join(items, ", ") + ")"
	}

	items := make([]string, len(args))
	for i, a := range args {
		items[i] = printDescription(a.Description, "  "+indentation, i == 0) + "  " + indentation + printInputValue(a)
	}
	return "(\n" + strings.Join(items, "\n") + "\n" + indentation + ")"
}

func printInputValue(v inputValue) string {
	s := v.Name + ": " + v.Type.String()
	if v.DefaultValue != nil {
		s += " = " + *v.DefaultValue
	}
	return s + printDeprecated(v.IsDeprecated, v.DeprecationReason)
}

func printDeprecated(deprecated bool, reason *string) string {
	if !deprecated {
		return ""
	}
	if reason == nil || *reason == "" || *reason == "No longer supported" {
		return " @deprecated"
	}
	return " @deprecated(reason: " + printString(*reason) + ")"
}

func printDescription(desc *string, indentation string, firstInBlock bool) string {
	if desc == nil || *desc == "" {
		return ""
	}
	block := printString(*desc)
	if isPrintableAsBlockString(*desc) {
		block = printBlockString(*desc)
	}
	prefix := indentation
	if indentation != "" && !firstInBlock {
		prefix = "\n" + indentation
	}
	return prefix + strings.ReplaceAll(block, "\n", "\n"+indentation) + "\n"
}

// printBlockString mirrors graphql-js's printBlockString.
func printBlockString(value string) string {
	escaped := strings.ReplaceAll(value, `"""`, `\"""`)
	lines := strings.Split(strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(escaped), "\n")
	isSingleLine := len(lines) == 1

	forceLeadingNewLine := len(lines) > 1
	for _, l := range lines[1:] {
		if l != "" && !isWhiteSpace(l[0]) {
			forceLeadingNewLine = false
			break
		}
	}

	hasTrailingTripleQuotes := strings.HasSuffix(escaped, `\"""`)
	hasTrailingQuote := strings.HasSuffix(value, `"`) && !hasTrailingTripleQuotes
	hasTrailingSlash := strings.HasSuffix(value, `\`)
	forceTrailingNewline := hasTrailingQuote || hasTrailingSlash

	printAsMultipleLines := !isSingleLine || len([]rune(value)) > 70 || forceTrailingNewline || forceLeadingNewLine || hasTrailingTripleQuotes

	var b strings.Builder
	skipLeadingNewLine := isSingleLine && value != "" && isWhiteSpace(value[0])
	if (printAsMultipleLines && !skipLeadingNewLine) || forceLeadingNewLine {
		b.WriteString("\n")
	}
	b.WriteString(escaped)
	if printAsMultipleLines || forceTrailingNewline {
		b.WriteString("\n")
	}
	return `"""` + b.String() + `"""`
}

// isPrintableAsBlockString mirrors graphql-js: descriptions with control
// characters, leading or trailing empty lines, or a common indent are printed
// as ordinary strings instead.
func isPrintableAsBlockString(value string) bool {
	if value == "" {
		return true
	}
	isEmptyLine, hasIndent, hasCommonIndent, seenNonEmptyLine := true, false, true, false
	for _, r := range value {
		switch {
		case r <= 0x08 || r == 0x0b || r == 0x0c || r == 0x0e || r == 0x0f || r == 0x0d:
			return false
		case r == '\n':
			if isEmptyLine && !seenNonEmptyLine {
				return false
			}
			seenNonEmptyLine = true
			isEmptyLine = true
			hasIndent = false
		case r == '\t' || r == ' ':
			hasIndent = hasIndent || isEmptyLine
		default:
			hasCommonIndent = hasCommonIndent && hasIndent
			isEmptyLine = false
		}
	}
	if isEmptyLine {
		return false
	}
	if hasCommonIndent && seenNonEmptyLine {
		return false
	}
	return true
}

func isWhiteSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// printString prints a GraphQL string literal. It uses JSON escaping, which
// differs from graphql-js only for control characters and U+2028/U+2029.
func printString(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(buf.String(), "\n")
}

const introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types { ...FullType }
    directives {
      name
      description
      locations
      args(includeDeprecated: true) { ...InputValue }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args(includeDeprecated: true) { ...InputValue }
    type { ...TypeRef }
    isDeprecated
    deprecationReason
  }
  inputFields(includeDeprecated: true) { ...InputValue }
  interfaces { ...TypeRef }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes { ...TypeRef }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
  isDeprecated
  deprecationReason
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`
