# go-shopify-graphql-model

Go types for the Shopify GraphQL Admin API, generated with gqlgen's modelgen
plugin. Module path `github.com/r0busta/go-shopify-graphql-model/v5`.

## Layout

- `schema.graphql`: the Shopify Admin API schema the models were generated
  from. Its API version determines the major version of this module.
- `graph/model/models_gen.go`: generated output. Never hand-edit. Regenerate
  with `go run .` after changing the schema or generator.
- `graph/model/types.go` and other non-`_gen` files: hand-written helpers and
  custom JSON handling that gqlgen cannot produce (for example unmarshaling
  fields typed as GraphQL interfaces).
- `graphql/`: custom scalar types (`Money`, `Decimal`) mapped in
  `gqlgen.yml`.
- `main.go`: the generator entry point. Its mutate hook adds `omitempty` to
  pointer and slice fields. `gqlgen.yml` sets `omit_enum_json_marshalers`
  so enums stay plain strings; `enum_json_test.go` guards that.
- `cmd/fetchschema`: fetches the schema for an API version from the public
  introspection endpoint on shopify.dev (no token needed) and prints it in
  graphql-js `printSchema` format. Instructions are in the README.

## Known gaps

- gqlgen emits Go interfaces for GraphQL interfaces, so JSON responses that
  include interface-typed fields (`Media`, `CollectionRuleConditionObject`,
  `MetaobjectField.jsonValue`) fail to unmarshal without custom code. Work on
  this is tracked in the issues.

## Working here

- `go build ./... && go vet ./... && go test ./...` must pass. The test
  compiles the very large generated file, so the first run takes a while.
- `.env` may hold a live store token for `cmd/fetchschema -store` and is
  gitignored. Never stage it.
- Consumers: `github.com/r0busta/go-shopify-graphql` pins a specific major
  version of this module. A schema bump here forces a major bump there.
