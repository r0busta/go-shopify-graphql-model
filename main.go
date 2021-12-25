package main

import (
	"fmt"
	"go/types"
	"os"
	"strings"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin/modelgen"
)

func main() {
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
		os.Exit(2)
	}

	// Attaching the mutation function onto modelgen plugin.
	p := modelgen.Plugin{
		MutateHook: mutateHook,
	}

	err = api.Generate(cfg,
		api.NoPlugins(),
		api.AddPlugin(&p),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}

func isPointer(v interface{}) bool {
	_, ok := v.(*types.Pointer)

	return ok
}

func isSlice(v interface{}) bool {
	_, ok := v.(*types.Slice)

	return ok
}

// Defining mutation function.
func mutateHook(b *modelgen.ModelBuild) *modelgen.ModelBuild {
	for _, model := range b.Models {
		for _, field := range model.Fields {
			if isPointer(field.Type) || isSlice(field.Type) {
				tag := strings.TrimSuffix(field.Tag, `"`)
				field.Tag = fmt.Sprintf(`%v,omitempty"`, tag)
			}
		}
	}

	return b
}
