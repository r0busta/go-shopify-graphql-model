package main_test

import (
	"encoding/json"
	"testing"

	"github.com/r0busta/go-shopify-graphql-model/v4/graph/model"
	"github.com/stretchr/testify/assert"
	null "gopkg.in/guregu/null.v4"
)

func TestSanity(t *testing.T) {
	product := model.Product{
		Title: "T-shirt, gopher print",
		PriceRangeV2: &model.ProductPriceRangeV2{
			MinVariantPrice: &model.MoneyV2{
				Amount:       *model.NewNullString(null.StringFrom("9.99")),
				CurrencyCode: model.CurrencyCodeUsd,
			},
			MaxVariantPrice: &model.MoneyV2{
				Amount:       *model.NewNullString(null.StringFrom("10.01")),
				CurrencyCode: model.CurrencyCodeUsd,
			},
		},
	}

	want := `{
    "createdAt": "",
    "defaultCursor": "",
    "description": "",
    "descriptionHtml": "",
    "descriptionPlainSummary": "",
    "handle": "",
    "hasOnlyDefaultVariant": false,
    "hasOutOfStockVariants": false,
    "hasVariantsThatRequiresComponents": false,
    "id": "",
    "inCollection": false,
    "isGiftCard": false,
    "legacyResourceId": "",
    "priceRangeV2": {
        "maxVariantPrice": {
            "amount": "10.01",
            "currencyCode": "USD"
        },
        "minVariantPrice": {
            "amount": "9.99",
            "currencyCode": "USD"
        }
    },
    "productType": "",
    "publicationCount": 0,
    "publishedInContext": false,
    "publishedOnChannel": false,
    "publishedOnCurrentChannel": false,
    "publishedOnCurrentPublication": false,
    "publishedOnPublication": false,
    "requiresSellingPlan": false,
    "sellingPlanGroupCount": 0,
    "status": "",
    "storefrontId": "",
    "title": "T-shirt, gopher print",
    "totalInventory": 0,
    "totalVariants": 0,
    "tracksInventory": false,
    "updatedAt": "",
    "vendor": ""
}`

	got, _ := json.MarshalIndent(product, "", "    ")
	assert.Equal(t, want, string(got))
}
