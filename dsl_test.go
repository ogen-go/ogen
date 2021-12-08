package ogen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	ex := &Spec{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:          "title",
			Description:    "description",
			TermsOfService: "terms of service",
			Contact: &Contact{
				Name:  "name",
				URL:   "url",
				Email: "email",
			},
			License: &License{
				Name: "name",
				URL:  "url",
			},
			Version: "0.1.0",
		},
	}
	ac := NewSpec().
		SetOpenAPI(ex.OpenAPI).
		SetInfo(NewInfo().
			SetTitle(ex.Info.Title).
			SetDescription(ex.Info.Description).
			SetTermsOfService(ex.Info.TermsOfService).
			SetVersion(ex.Info.Version).
			SetContact(NewContact().
				SetName(ex.Info.Contact.Name).
				SetURL(ex.Info.Contact.URL).
				SetEmail(ex.Info.Contact.Email),
			).
			SetLicense(NewLicense().
				SetName(ex.Info.License.Name).
				SetURL(ex.Info.License.URL),
			),
		)
	assert.Equal(t, ex, ac)
}
