# Contributing

Thanks for helping out. This library started as a personal tool, so the
process is lightweight, but a few things make pull requests much easier to
merge.

## Before opening a pull request

- Run `gofmt`, `go vet ./...` and `go test ./...`. CI runs the same checks.
- Keep the module path as `github.com/r0busta/go-shopify-graphql-model/v5`.
  If you develop in a fork, use a `replace` directive locally instead of
  renaming the module in `go.mod`.
- `graph/model/models_gen.go` is generated. Do not edit it by hand; change
  the generator (`main.go`, `gqlgen.yml`) or the schema and regenerate. Hand
  written additions such as custom `UnmarshalJSON` methods live in separate
  files in `graph/model/`.
- Keep pull requests focused. Separate generator changes, hand-written model
  fixes and schema updates into different PRs where practical.

## Versioning

The models are generated from a specific Shopify Admin API schema version.
Regenerating from a newer schema removes and renames types, so a schema
update is a breaking change and gets a new major version. That in turn
requires a new major version of
[go-shopify-graphql](https://github.com/r0busta/go-shopify-graphql), which
depends on this module. Fixes that do not change the generated output can
ship as minor or patch releases.

## Regenerating the models

See the README for the fetch and generate steps. The default fetch uses the
public introspection endpoint on shopify.dev and needs no credentials. If
you introspect a store instead, export its access token as `ACCESS_TOKEN`;
if you keep it in `.env`, that file is gitignored, never commit it.
