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
  introspection endpoint on shopify.dev (no token needed) and writes it in
  graphql-js `printSchema` format, deprecated arguments and input fields
  included. Instructions are in the README.

## Known gaps

- gqlgen emits Go interfaces for GraphQL interfaces and unions, and
  encoding/json cannot decode into them. `graph/model/unmarshal.go` adds
  decoders for `Media` (edges and nodes), `CollectionRuleConditionObject`,
  `Collection.sources` and `MetaobjectField.jsonValue`, each backed by a type
  registry in `types.go` that `registry_test.go` checks against the schema.
  Fields without a decoder still fail when selected: `Product.featuredMedia`,
  `Order.purchasingEntity`, `Metafield.owner`, `Metafield.reference` and
  `references`, `MetaobjectField.reference`, and the media mutation payloads.
  Add a registry, a decoder and tests in the same style when one is needed.

## Working here

- `go build ./... && go vet ./... && go test ./...` must pass. The test
  compiles the very large generated file, so the first run takes a while.
- `.env` may hold a live store token (`ACCESS_TOKEN`) for
  `cmd/fetchschema -store`; nothing loads it automatically, export it
  yourself. It is gitignored. Never stage it.
- Consumers: `github.com/r0busta/go-shopify-graphql` pins a specific major
  version of this module. A schema bump here forces a major bump there.
