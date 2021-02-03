const path = require("path");
const fs = require("fs");
const fetch = require("node-fetch");
const {
    getIntrospectionQuery,
    printSchema,
    buildClientSchema
} = require("graphql");

async function main() {
  const introspectionQuery = getIntrospectionQuery();

    const response = await fetch(
        "https://graphql.myshopify.com/api/graphql",
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "X-Shopify-Storefront-Access-Token": "dd4d4dc146542ba7763305d71d1b3d38"
            },
            body: JSON.stringify({ query: introspectionQuery })
        }
    );

    const { data } = await response.json();

    const schema = buildClientSchema(data);

    const outputFile = path.join(__dirname, "./result.gql");

    await fs.promises.writeFile(outputFile, printSchema(schema));
}

main();
