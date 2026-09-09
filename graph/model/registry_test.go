package model

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The type registries must list exactly the schema's implementations, so
// regenerating from a new schema surfaces missing entries here rather than as
// decode errors at runtime.

func loadSchema(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../schema.graphql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	return string(b)
}

func unionMembers(t *testing.T, schema, union string) []string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^union ` + union + `\s*=\s*([^\n]+)`)
	m := re.FindStringSubmatch(schema)
	if m == nil {
		t.Fatalf("union %s not found in schema", union)
	}
	var members []string
	for _, name := range strings.Split(m[1], "|") {
		members = append(members, strings.TrimSpace(name))
	}
	sort.Strings(members)
	return members
}

func interfaceImplementers(t *testing.T, schema, iface string) []string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^type (\w+) implements ([^{]+)\{`)
	var names []string
	for _, m := range re.FindAllStringSubmatch(schema, -1) {
		for _, impl := range strings.Split(m[2], "&") {
			if strings.TrimSpace(impl) == iface {
				names = append(names, m[1])
			}
		}
	}
	if len(names) == 0 {
		t.Fatalf("no implementations of %s found in schema", iface)
	}
	sort.Strings(names)
	return names
}

func registryNames(reg map[string]reflect.Type) []string {
	names := make([]string, 0, len(reg))
	for name := range reg {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func TestMediaRegistryMatchesSchema(t *testing.T) {
	want := interfaceImplementers(t, loadSchema(t), "Media")
	if got := registryNames(mediaTypes); !reflect.DeepEqual(got, want) {
		t.Errorf("mediaTypes = %v, schema has %v", got, want)
	}
	for name, typ := range mediaTypes {
		if _, ok := reflect.New(typ).Interface().(Media); !ok {
			t.Errorf("%s does not implement Media", name)
		}
	}
}

func TestCollectionRuleConditionObjectRegistryMatchesSchema(t *testing.T) {
	want := unionMembers(t, loadSchema(t), "CollectionRuleConditionObject")
	if got := registryNames(collectionRuleConditionObjectTypes); !reflect.DeepEqual(got, want) {
		t.Errorf("collectionRuleConditionObjectTypes = %v, schema has %v", got, want)
	}
	for name, typ := range collectionRuleConditionObjectTypes {
		if _, ok := reflect.New(typ).Interface().(CollectionRuleConditionObject); !ok {
			t.Errorf("%s does not implement CollectionRuleConditionObject", name)
		}
	}
}
