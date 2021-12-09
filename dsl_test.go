package ogen_test

import (
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

const (
	pathWithID    = "/PathItem/with/{id}"
	refPathWithID = "/ref/PathItem/with/id"
)

func TestBuilder(t *testing.T) {
	// referenced query Parameter
	authQ := ogen.NewParameter().
		InQuery().
		SetName("auth").
		SetDescription("Optional bearer token").
		ToNamed("authInQuery")
	// referenced header Parameter
	authH := ogen.NewNamedParameter(
		"authInHeader",
		ogen.NewParameter().
			SetIn("header").
			SetName("Authorization").
			SetDescription("Optional bearer token"),
	)
	// referenced cookie Parameter
	csrf := ogen.NewParameter().
		InCookie().
		SetName("csrf").
		SetDescription("CSRF token").
		ToNamed("csrf")
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
		Paths: map[string]ogen.PathItem{
			pathWithID: {
				Description: "This is my first PathItem",
				Get: &ogen.Operation{
					Tags:        []string{"default"},
					Description: "Description for my PathItem",
					OperationID: "PathItem-with-id",
					Parameters: []ogen.Parameter{
						{
							Name:        "id",
							In:          "PathItem",
							Description: "ID Parameter in PathItem",
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
			Schemas:   nil, // TODO
			Responses: nil, // TODO
			Parameters: map[string]ogen.Parameter{
				authQ.Name: *authQ.Parameter,
				authH.Name: *authH.Parameter,
				csrf.Name:  *csrf.Parameter,
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
				authQ.AsLocalRef(),
				authH.AsLocalRef(),
				csrf.AsLocalRef(),
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
		AddNamedPaths(path).
		AddPathItem(refPathWithID, path.AsLocalRef()).
		AddNamedParameters(authQ, authH, csrf)
	assert.Equal(t, ex, ac)
}
