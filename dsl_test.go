package ogen_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

const (
	pathWithID    = "/path/with/{id}"
	refPathWithID = "/ref/path/with/id"
	pathWithBody  = "/path/with/body"
)

var (
	_extensions = ogen.Extensions(nil)
	_common     = ogen.OpenAPICommon{
		Extensions: _extensions,
	}

	// reusable query param
	_queryParam = ogen.NewParameter().
			InQuery().
			SetName("auth").
			SetDescription("Optional bearer token").
			ToNamed("authInQuery")
	// reusable header param
	_headerParam = ogen.NewNamedParameter(
		"authInHeader",
		ogen.NewParameter().
			SetIn("header").
			SetName("Authorization").
			SetDescription("Optional bearer token"),
	)
	// reusable cookie param
	_cookieParam = ogen.NewParameter().
			InCookie().
			SetName("csrf").
			SetDescription("CSRF token").
			ToNamed("csrf")
	// reusable pet schema
	_petSchema = ogen.NewNamedSchema(
		"Pet",
		ogen.NewSchema().
			SetDescription("A Pet").
			AddRequiredProperties(
				ogen.Int().ToProperty("required_Int"),
				ogen.Int32().ToProperty("required_Int32"),
				ogen.Int64().ToProperty("required_Int64"),
				ogen.Float().ToProperty("required_Float"),
				ogen.Double().ToProperty("required_Double"),
				ogen.String().ToProperty("required_String"),
				ogen.Bytes().ToProperty("required_Bytes"),
				ogen.Binary().ToProperty("required_Binary"),
				ogen.Bool().ToProperty("required_Bool"),
				ogen.Date().ToProperty("required_Date"),
				ogen.DateTime().ToProperty("required_DateTime"),
				ogen.Password().ToProperty("required_Password"),
				ogen.Int32().AsArray().ToProperty("required_array_Int32"),
				ogen.Int32().AsEnum(json.RawMessage("0"), json.RawMessage("0"), json.RawMessage("1")).
					ToProperty("required_enum_Int32"),
				ogen.Int32().AsEnum(json.RawMessage(`"off"`), json.RawMessage(`"0"`), json.RawMessage(`"1"`)).
					ToProperty("required_enum_String"),
			).
			AddOptionalProperties(
				ogen.UUID().ToProperty("optional_UUID"),
				ogen.Int32().ToProperty("optional_Int32"),
				ogen.Int64().ToProperty("optional_Int64"),
				ogen.Float().ToProperty("optional_Float"),
				ogen.Double().ToProperty("optional_Double"),
				ogen.String().ToProperty("optional_String"),
				ogen.Bytes().ToProperty("optional_Bytes"),
				ogen.Binary().ToProperty("optional_Binary"),
				ogen.Bool().ToProperty("optional_Bool"),
				ogen.Date().ToProperty("optional_Date"),
				ogen.DateTime().ToProperty("optional_DateTime"),
				ogen.Password().ToProperty("optional_Password"),
			),
	)
	// reusable toy schema
	_toySchema = ogen.NewSchema().
			SetDescription("A toy of a Pet").
			ToNamed("User")
)

func TestBuilder(t *testing.T) {
	// expected result
	ex := &ogen.Spec{
		OpenAPI: "3.1.0",
		Info: ogen.Info{
			Title:          "title",
			Description:    "description",
			TermsOfService: "terms of service",
			Contact: &ogen.Contact{
				Name:  "Name",
				URL:   "url",
				Email: "email",
			},
			License: &ogen.License{
				Name: "Name",
				URL:  "url",
			},
			Version: "0.1.0",
		},
		Servers: []ogen.Server{
			{"staging", "staging.api.com", nil, _common},
			{"production", "api.com", nil, _common},
		},
		Paths: map[string]*ogen.PathItem{
			pathWithID: {
				Description: "This is my first path",
				Parameters: []*ogen.Parameter{
					{Ref: "#/components/parameters/authInQuery"},
					{Ref: "#/components/parameters/authInHeader"},
					{Ref: "#/components/parameters/csrf"},
				},
				Get: &ogen.Operation{
					Tags:        []string{"default"},
					Description: "Description for my path",
					OperationID: "path-with-id",
					Parameters: []*ogen.Parameter{
						{
							Name:        "id",
							In:          "path",
							Description: "ID Parameter in path",
							Required:    true,
							Schema:      &ogen.Schema{Type: "integer", Format: "int32"},
						},
					},
					Responses: ogen.Responses{
						"error": {Ref: "#/components/responses/error"},
						"ok": {
							Description: "Success",
							Content: map[string]ogen.Media{
								ir.EncodingJSON.String(): {Schema: &ogen.Schema{
									Type:        "object",
									Description: "Success",
									Properties: []ogen.Property{
										{Name: "prop1", Schema: &ogen.Schema{Type: "integer", Format: "int32"}},
										{Name: "prop2", Schema: &ogen.Schema{Type: "string"}},
									},
								}},
							},
						},
					},
				},
			},
			refPathWithID: {
				Ref: "#/paths/~1path~1with~1{id}",
			},
			pathWithBody: {
				Post: &ogen.Operation{
					Tags:        []string{"post"},
					Description: "Description for my path with body",
					OperationID: "path-with-body",
					Parameters: []*ogen.Parameter{
						{Ref: "#/components/parameters/authInQuery"},
						{Ref: "#/components/parameters/authInHeader"},
						{Ref: "#/components/parameters/csrf"},
					},
					Responses:   ogen.Responses{"error": {Ref: "#/components/responses/error"}},
					RequestBody: &ogen.RequestBody{Ref: "#/components/requestBodies/~1path~1with~1body"},
				},
			},
		},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				_petSchema.Name: _petSchema.Schema,
				_toySchema.Name: _toySchema.Schema,
			},
			Responses: ogen.Responses{
				"error": {
					Description: "An Error Response",
					Content: map[string]ogen.Media{
						ir.EncodingJSON.String(): {Schema: &ogen.Schema{
							Type:        "object",
							Description: "Error Response Schema",
							Properties: []ogen.Property{
								{Name: "code", Schema: &ogen.Schema{Type: "integer", Format: "int32"}},
								{Name: "status", Schema: &ogen.Schema{Type: "string"}},
							},
						}},
					},
				},
			},
			Parameters: map[string]*ogen.Parameter{
				"authInQuery": {
					Name:        "auth",
					In:          "query",
					Description: "Optional bearer token",
				},
				"authInHeader": {
					Name:        "Authorization",
					In:          "header",
					Description: "Optional bearer token",
				},
				"csrf": {
					Name:        "csrf",
					In:          "cookie",
					Description: "CSRF token",
				},
			},
			RequestBodies: map[string]*ogen.RequestBody{
				pathWithBody: {
					Description: "Referenced RequestBody",
					Content: map[string]ogen.Media{
						ir.EncodingJSON.String(): {
							Schema: &ogen.Schema{Ref: "#/components/schemas/" + _toySchema.Name},
						},
					},
					Required: true,
				},
			},
		},
		Extensions: _extensions,
	}
	// referenced path
	path := ogen.NewPathItem().
		SetDescription(ex.Paths[pathWithID].Description).
		AddParameters(_queryParam.AsLocalRef(), _headerParam.AsLocalRef(), _cookieParam.AsLocalRef()).
		SetGet(ogen.NewOperation().
			AddTags(ex.Paths[pathWithID].Get.Tags...).
			SetDescription(ex.Paths[pathWithID].Get.Description).
			SetOperationID(ex.Paths[pathWithID].Get.OperationID).
			AddParameters(
				ogen.NewParameter().
					InPath().
					SetName(ex.Paths[pathWithID].Get.Parameters[0].Name).
					SetDescription(ex.Paths[pathWithID].Get.Parameters[0].Description).
					SetRequired(true).
					SetSchema(ogen.Int32()),
			).
			AddNamedResponses(
				ex.RefResponse("error"),
				ogen.NewResponse().
					SetDescription(ex.Paths[pathWithID].Get.Responses["ok"].Description).
					SetJSONContent(ogen.NewSchema().
						SetDescription(ex.Paths[pathWithID].Get.Responses["ok"].Content[ir.EncodingJSON.String()].Schema.Description).
						AddOptionalProperties(
							ogen.Int32().ToProperty("prop1"),
							ogen.String().ToProperty("prop2"),
						),
					).
					ToNamed("ok"),
			),
		).
		ToNamed(pathWithID)
	// actual result
	ac := ogen.NewSpec().
		SetOpenAPI(ex.OpenAPI).
		SetInfo(ogen.NewInfo().
			SetTitle(ex.Info.Title).
			SetDescription(ex.Info.Description).
			SetTermsOfService(ex.Info.TermsOfService).
			SetContact(ogen.NewContact().
				SetName(ex.Info.Contact.Name).
				SetURL(ex.Info.Contact.URL).
				SetEmail(ex.Info.Contact.Email),
			).
			SetLicense(ogen.NewLicense().
				SetName(ex.Info.License.Name).
				SetURL(ex.Info.License.URL),
			).
			SetVersion(ex.Info.Version),
		).
		AddServers(
			&ex.Servers[0],
			ogen.NewServer().
				SetDescription(ex.Servers[1].Description).
				SetURL(ex.Servers[1].URL),
		).
		AddNamedPathItems(
			path,
			ogen.NewPathItem().
				SetDescription(ex.Paths[pathWithBody].Description).
				SetPost(ogen.NewOperation().
					AddTags(ex.Paths[pathWithBody].Post.Tags...).
					SetDescription(ex.Paths[pathWithBody].Post.Description).
					SetOperationID(ex.Paths[pathWithBody].Post.OperationID).
					AddParameters(_queryParam.AsLocalRef(), _headerParam.AsLocalRef(), _cookieParam.AsLocalRef()).
					AddNamedResponses(ex.RefResponse("error")).
					SetRequestBody(ex.RefRequestBody(pathWithBody).RequestBody),
				).
				ToNamed(pathWithBody),
		).
		AddPathItem(refPathWithID, path.AsLocalRef()).
		AddNamedParameters(_queryParam, _headerParam, _cookieParam).
		AddNamedSchemas(_petSchema, _toySchema).
		AddNamedResponses(
			ogen.NewResponse().
				SetDescription(ex.Components.Responses["error"].Description).
				SetJSONContent(ogen.NewSchema().
					SetDescription(ex.Components.Responses["error"].Content[ir.EncodingJSON.String()].Schema.Description).
					AddOptionalProperties(
						ogen.Int32().ToProperty("code"),
						ogen.String().ToProperty("status"),
					),
				).
				ToNamed("error"),
		).
		AddNamedRequestBodies(
			ogen.NewRequestBody().
				SetDescription(ex.Components.RequestBodies[pathWithBody].Description).
				SetJSONContent(ex.RefSchema(_toySchema.Name).Schema).
				SetRequired(true).
				ToNamed(pathWithBody),
		)
	assert.Equal(t, ex, ac)

	ex.SetServers(nil)
	assert.Nil(t, ex.Servers)

	ex.SetPaths(nil)
	assert.Nil(t, ex.Paths)

	ex.SetComponents(nil)
	assert.Nil(t, ex.Components)
	assert.Nil(t, ex.RefSchema(""))
	assert.Nil(t, ex.RefResponse(""))
	assert.Nil(t, ex.RefRequestBody(""))

	req := ogen.NewRequestBody().SetContent(map[string]ogen.Media{"key": {}})
	assert.Equal(t, req, &ogen.RequestBody{Content: map[string]ogen.Media{"key": {}}})

	pi := ogen.NewPathItem().
		SetPut(ogen.NewOperation().SetOperationID("put").SetTags([]string{"tag1", "tag2"})).
		SetDelete(ogen.NewOperation().SetOperationID("delete").SetSummary("summary")).
		SetOptions(ogen.NewOperation().SetOperationID("options").SetParameters([]*ogen.Parameter{_cookieParam.Parameter})).
		SetHead(ogen.NewOperation().SetOperationID("head").SetResponses(ogen.Responses{"resp": ogen.NewResponse()})).
		SetPatch(ogen.NewOperation().SetOperationID("patch").AddParameters(ogen.NewParameter().InHeader().SetDeprecated(true))).
		SetTrace(ogen.NewOperation().SetOperationID("trace")).
		SetQuery(ogen.NewOperation().SetOperationID("query")).
		SetAdditionalOperations(map[string]*ogen.Operation{"LINK": ogen.NewOperation().SetOperationID("link")}).
		SetAdditionalOperation("UNLINK", ogen.NewOperation().SetOperationID("unlink")).
		SetServers([]ogen.Server{{"url1", "desc1", nil, _common}}).
		AddServers(ogen.NewServer().SetDescription("desc2").SetURL("url2")).
		SetParameters([]*ogen.Parameter{_queryParam.Parameter})
	assert.Equal(t, &ogen.PathItem{
		Put:     &ogen.Operation{OperationID: "put", Tags: []string{"tag1", "tag2"}},
		Delete:  &ogen.Operation{OperationID: "delete", Summary: "summary"},
		Options: &ogen.Operation{OperationID: "options", Parameters: []*ogen.Parameter{_cookieParam.Parameter}},
		Head:    &ogen.Operation{OperationID: "head", Responses: ogen.Responses{"resp": &ogen.Response{}}},
		Patch:   &ogen.Operation{OperationID: "patch", Parameters: []*ogen.Parameter{{In: "header", Deprecated: true}}},
		Trace:   &ogen.Operation{OperationID: "trace"},
		Query:   &ogen.Operation{OperationID: "query"},
		AdditionalOperations: map[string]*ogen.Operation{
			"LINK":   {OperationID: "link"},
			"UNLINK": {OperationID: "unlink"},
		},
		Servers: []ogen.Server{
			{"url1", "desc1", nil, _common},
			{"url2", "desc2", nil, _common},
		},
		Parameters: []*ogen.Parameter{_queryParam.Parameter},
		Common:     _common,
	}, pi)

	mlt := uint64(1)
	mltStr := ogen.Num("1")
	maxn := int64(2)
	maxStr := ogen.Num("2")
	umax := uint64(maxn)
	assert.Equal(t, &ogen.Schema{
		Ref:         "ref",
		Description: "desc",
		Type:        "object",
		Format:      "",
		Properties:  []ogen.Property{{Name: "prop"}},
		Required:    []string{"prop"},
		Items: &ogen.Items{
			Item: ogen.String(),
		},
		Nullable:         true,
		AllOf:            []*ogen.Schema{ogen.NewSchema()},
		OneOf:            []*ogen.Schema{ogen.NewSchema()},
		AnyOf:            []*ogen.Schema{ogen.NewSchema()},
		Discriminator:    &ogen.Discriminator{PropertyName: "prop"},
		Enum:             []json.RawMessage{json.RawMessage("0"), json.RawMessage("1")},
		MultipleOf:       mltStr,
		Maximum:          maxStr,
		ExclusiveMaximum: true,
		Minimum:          maxStr,
		ExclusiveMinimum: true,
		MaxLength:        &umax,
		MinLength:        &umax,
		Pattern:          "",
		MaxItems:         &umax,
		MinItems:         &umax,
		UniqueItems:      true,
		MaxProperties:    &umax,
		MinProperties:    &umax,
		Default:          jsonschema.Default("0"),
	}, ogen.NewSchema().
		SetRef("ref").
		SetDescription("desc").
		SetType("string").
		SetFormat("").
		SetProperties(&ogen.Properties{{Name: "prop"}}).
		SetRequired([]string{"prop"}).
		SetItems(ogen.String()).
		SetNullable(true).
		SetAllOf([]*ogen.Schema{ogen.NewSchema()}).
		SetOneOf([]*ogen.Schema{ogen.NewSchema()}).
		SetAnyOf([]*ogen.Schema{ogen.NewSchema()}).
		SetDiscriminator(&ogen.Discriminator{PropertyName: "prop"}).
		SetEnum([]json.RawMessage{json.RawMessage("0"), json.RawMessage("1")}).
		SetMultipleOf(&mlt).
		SetMaximum(&maxn).
		SetExclusiveMaximum(true).
		SetMinimum(&maxn).
		SetExclusiveMinimum(true).
		SetMaxLength(&umax).
		SetMinLength(&umax).
		SetPattern("").
		SetMaxItems(&umax).
		SetMinItems(&umax).
		SetUniqueItems(true).
		SetMaxProperties(&umax).
		SetMinProperties(&umax).
		SetDefault(json.RawMessage("0")),
	)
}
