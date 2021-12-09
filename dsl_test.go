package ogen_test

import (
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

const (
	pathWithID    = "/path/with/{id}"
	refPathWithID = "/ref/path/with/id"
)

var (
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
			).
			AddOptionalProperties(
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
			{"staging", "staging.api.com"},
			{"production", "api.com"},
		},
		Paths: map[string]*ogen.PathItem{
			pathWithID: {
				Description: "This is my first path",
				Get: &ogen.Operation{
					Tags:        []string{"default"},
					Description: "Description for my path",
					OperationID: "path-with-id",
					Parameters: []*ogen.Parameter{
						{
							Name:        "id",
							In:          "PathItem",
							Description: "ID Parameter in path",
							Required:    true,
							// TODO: Schema
							// TODO: Required
							// TODO: Deprecated
							// TODO: Content
							// TODO: Style
							// TODO: Explode
						},
						{Ref: "#/components/parameters/authInQuery"},
						{Ref: "#/components/parameters/authInHeader"},
						{Ref: "#/components/parameters/csrf"},
					},
					RequestBody: nil, // TODO
					Responses:   nil, // TODO
				},
			},
			refPathWithID: {
				Ref: "#/paths/~1path~1with~1{id}",
			},
		},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				_petSchema.Name: _petSchema.Schema,
				_toySchema.Name: _toySchema.Schema,
			},
			Responses: nil, // TODO
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
			RequestBodies: nil, // TODO
		},
	}
	// referenced path
	path := ogen.NewPathItem().
		SetDescription(ex.Paths[pathWithID].Description).
		SetGet(ogen.NewOperation().
			AddTags(ex.Paths[pathWithID].Get.Tags...).
			SetDescription(ex.Paths[pathWithID].Get.Description).
			SetOperationID(ex.Paths[pathWithID].Get.OperationID).
			AddParameters(
				ogen.NewParameter().
					InPath().
					SetName(ex.Paths[pathWithID].Get.Parameters[0].Name).
					SetDescription(ex.Paths[pathWithID].Get.Parameters[0].Description).
					SetRequired(true),
				_queryParam.AsLocalRef(),
				_headerParam.AsLocalRef(),
				_cookieParam.AsLocalRef(),
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
		AddNamedPathItems(path).
		AddPathItem(refPathWithID, path.AsLocalRef()).
		AddNamedParameters(_queryParam, _headerParam, _cookieParam).
		AddNamedSchemas(_petSchema, _toySchema)
	assert.Equal(t, ex, ac)
}
