const path = require("path")
const fs = require("fs")
const fetch = require("node-fetch")
const { getIntrospectionQuery, printSchema, buildClientSchema } = require("graphql")

async function main() {
    const introspectionQuery = getIntrospectionQuery()

    const response = await fetch(`https://${process.env.STORE}.myshopify.com/admin/api/2021-04/graphql.json`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Shopify-Access-Token": process.env.PASSWORD,
        },
        body: JSON.stringify({ query: introspectionQuery }),
    })

    const { data } = await response.json()

    const schema = buildClientSchema(data)

    const outputFile = path.join(__dirname, "./result.graphql")

    await fs.promises.writeFile(outputFile, printSchema(schema))
}

main()
