package main

var commonLinks = []Schema{
	{
		File: "gotd_bot_api.json",
		Link: "https://raw.githubusercontent.com/gotd/botapi/main/_oas/openapi.json",
	},
	{
		File: "superset.json",
		Link: "https://raw.githubusercontent.com/apache/superset/master/docs/static/resources/openapi.json",
	},
	{
		File:       "dapr.json",
		Link:       "https://raw.githubusercontent.com/dapr/dapr/master/swagger/swagger.json",
		SkipReason: "securities",
	},
	{
		File:       "medusa/admin-spec3.json",
		Link:       "https://raw.githubusercontent.com/medusajs/medusa/master/docs/api/admin-spec3.json",
		SkipReason: "invalid schema: paths: /draft-orders/{id}/line-items:  path parameter not specified: \"id\"",
	},
	{
		File:       "grocy.json",
		Link:       "https://raw.githubusercontent.com/grocy/grocy/master/grocy.openapi.json",
		SkipReason: "invalid schema: unknown ref \"#/components/schemas/ExposedEntity_NotIncludingNotListable\"",
	},
	{
		File:       "redoc-demo.json",
		Link:       "https://raw.githubusercontent.com/Redocly/redoc/master/demo/big-openapi.json",
		SkipReason: "invalid schema: parse enum values: expected type \"string\", got \"number\"",
	},
	// GPL 3.0
	// {
	// 	File: "radarr.json",
	// 	Link: "https://raw.githubusercontent.com/Radarr/Radarr/develop/src/Radarr.Api.V3/swagger.json",
	// },
	// {
	// 	File: "netdata.json",
	// 	Link: "https://raw.githubusercontent.com/netdata/netdata/master/web/api/netdata-swagger.json",
	// },
}

var redocFixtures = []Schema{
	{
		File: "redoc/callback.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/callback.json",
	},
	{
		File: "redoc/discriminator.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/discriminator.json",
	},
	{
		File:       "redoc/fields.json",
		Link:       "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/fields.json",
		SkipReason: "invalid parameter: \"testParam\": path parameters must be required",
	},
	{
		File: "redoc/oneOfHoist.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/oneOfHoist.json",
	},
	{
		File: "redoc/oneOfTitles.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/oneOfTitles.json",
	},
	{
		File: "redoc/siblingRefDescription.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/siblingRefDescription.json",
	},

	// 3.1
	{
		File: "redoc/pathItems.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/3.1/pathItems.json",
	},
	{
		File: "redoc/schemaDefinition.json",
		Link: "https://raw.githubusercontent.com/Redocly/redoc/master/src/services/__tests__/fixtures/3.1/schemaDefinition.json",
	},
}

var autoRestLinks = []Schema{
	{
		File: "autorest/ApiManagementClient-openapi.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/ApiManagementClient-openapi.json",
	},
	{
		File: "autorest/additionalProperties.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/additionalProperties.json",
	},
	{
		File: "autorest/complex-model.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/complex-model.json",
	},
	{
		File:       "autorest/default-response.json",
		Link:       "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/default-response.json",
		SkipReason: "invalid schema: unexpected schema type: \"file\"",
	},
	{
		File:       "autorest/exec-service.json",
		Link:       "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/exec-service.json",
		SkipReason: "invalid schema: unexpected schema type: \"file\"",
	},
	{
		File: "autorest/extensible-enums-swagger.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/extensible-enums-swagger.json",
	},
	{
		File: "autorest/header.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/header.json",
	},
	{
		File: "autorest/lro.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/lro.json",
	},
	{
		File: "autorest/luis.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/luis.json",
	},
	{
		File: "autorest/storage.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/storage.json",
	},
	{
		File: "autorest/url-multi-collectionFormat.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/url-multi-collectionFormat.json",
	},
	{
		File:       "autorest/url.json",
		Link:       "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/url.json",
		SkipReason: "unsupported schema: unexpected style: spaceDelimited",
	},
	{
		File:       "autorest/validation.json",
		Link:       "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/validation.json",
		SkipReason: "invalid schema: path parameter not specified: \"apiVersion\"",
	},
	{
		File: "autorest/xml-service.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/xml-service.json",
	},
	{
		File: "autorest/xms-error-responses.json",
		Link: "https://raw.githubusercontent.com/Azure/autorest/main/packages/libs/oai2-to-oai3/test/resources/conversion/oai3/xms-error-responses.json",
	},
}

var linkSets = [][]Schema{
	commonLinks,
	redocFixtures,
	autoRestLinks,
}
