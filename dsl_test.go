package ogen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	ex := Spec{
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
	ac := SpecBuilder().
		SetOpenAPI(ex.OpenAPI).
		SetInfo(InfoBuilder().
			SetTitle(ex.Info.Title).
			SetDescription(ex.Info.Description).
			SetTermsOfService(ex.Info.TermsOfService).
			SetVersion(ex.Info.Version).
			SetContact(ContactBuilder().
				SetName(ex.Info.Contact.Name).
				SetURL(ex.Info.Contact.URL).
				SetEmail(ex.Info.Contact.Email).
				Contact(),
			).
			SetLicense(LicenseBuilder().
				SetName(ex.Info.License.Name).
				SetURL(ex.Info.License.URL).
				License()).
			Info(),
		).
		Spec()
	assert.Equal(t, ex, ac)
}
