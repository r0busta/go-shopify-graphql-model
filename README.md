# go-shopify-graphql-model

This is a simple library to help you use Shopify's GraphQL objects in your Go code.

## Getting started

0. Install dependencies

    ```bash
    yarn
    ```

1. Fetch the Shopify graphql schema

    ```bash
    STORE=my-store PASSWORD=my-pass yarn fetch
    ```

2. Rename or copy its content from `result.graphql` to `schema.graphql`
3. Remove the following declaration from `schema.graphql` so that models can be generated

    ```graphql
    schema {
        query: QueryRoot
        mutation: Mutation
    }
    ```

4. Generate models

    ```bash
    go run main.go 
    ```
