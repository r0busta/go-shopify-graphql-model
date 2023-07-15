import { promises } from "fs"
import fetch from "node-fetch"
import { getIntrospectionQuery, printSchema, buildClientSchema } from "graphql"

async function main() {
    const introspectionQuery = getIntrospectionQuery()

    const response = await fetch(`https://${process.env.STORE}.myshopify.com/admin/api/${process.env.API_VERSION}/graphql.json`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Shopify-Access-Token": process.env.ACCESS_TOKEN,
        },
        body: JSON.stringify({ query: introspectionQuery }),
    })

    const { data } = await response.json()

    const schema = buildClientSchema(data)

    const outputFile = "./result.graphql"

    await promises.writeFile(outputFile, printSchema(schema))
}

main()
