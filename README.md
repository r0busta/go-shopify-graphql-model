# go-shopify-graphql-model

Go types for the Shopify GraphQL Admin API, generated from the API schema
with [gqlgen](https://github.com/99designs/gqlgen).

```go
import "github.com/r0busta/go-shopify-graphql-model/v5/graph/model"
```

The generated models live in `graph/model`. The major version of this module
follows the Admin API version the models were generated from; see
`schema.graphql` for the exact schema. A client that uses these models is
[go-shopify-graphql](https://github.com/r0busta/go-shopify-graphql).

## Regenerating the models

1. Fetch the schema for an API version. This uses the public introspection
   endpoint on shopify.dev and needs no store or token:

    ```bash
    go run ./cmd/fetchschema -version 2026-07
    ```

   To introspect a specific store instead, pass `-store my-store` and export
   `ACCESS_TOKEN` with an Admin API access token. Deprecated arguments and
   input fields are included since the API still accepts them.

2. Generate the models:

    ```bash
    go run .
    ```

3. Run the tests:

    ```bash
    go test ./...
    ```

Both commands are safe to re-run; they overwrite `schema.graphql` and
`graph/model/models_gen.go`.
