import { join } from "path"
import { promises } from "fs"
import fetch from "node-fetch"
import { getIntrospectionQuery, printSchema, buildClientSchema } from "graphql"

const __dirname = new URL(".", import.meta.url).pathname

async function main() {
    const introspectionQuery = getIntrospectionQuery()

    const response = await fetch(`https://${process.env.STORE}.myshopify.com/admin/api/unstable/graphql.json`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Shopify-Access-Token": process.env.PASSWORD,
        },
        body: JSON.stringify({ query: introspectionQuery }),
    })

    const { data } = await response.json()

    const schema = buildClientSchema(data)

    const outputFile = join(__dirname, "./result.graphql")

    await promises.writeFile(outputFile, printSchema(schema))
}

main()
