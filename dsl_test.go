package ogen_test

import (
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	ex := &ogen.Spec{
		OpenAPI: "3.1.0",
		Info: ogen.Info{
			Title:          "title",
			Description:    "description",
			TermsOfService: "terms of service",
			Contact: &ogen.Contact{
				Name:  "name",
				URL:   "url",
				Email: "email",
			},
			License: &ogen.License{
				Name: "name",
				URL:  "url",
			},
			Version: "0.1.0",
		},
		Servers: []ogen.Server{
			{"staging", "staging.api.com"},
			{"production", "api.com"},
		},
	}
	ac := ogen.NewSpec().
		SetOpenAPI(ex.OpenAPI).
		SetInfo(ogen.NewInfo().
			SetTitle(ex.Info.Title).
			SetDescription(ex.Info.Description).
			SetTermsOfService(ex.Info.TermsOfService).
			SetVersion(ex.Info.Version).
			SetContact(ogen.NewContact().
				SetName(ex.Info.Contact.Name).
				SetURL(ex.Info.Contact.URL).
				SetEmail(ex.Info.Contact.Email),
			).
			SetLicense(ogen.NewLicense().
				SetName(ex.Info.License.Name).
				SetURL(ex.Info.License.URL),
			),
		).
		AddServer(&ex.Servers[0]).
		AddServer(&ex.Servers[1])
	assert.Equal(t, ex, ac)
}
