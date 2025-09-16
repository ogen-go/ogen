package openapi

import "github.com/ogen-go/ogen/jsonschema"

type (
	// Extensions is a map of OpenAPI extensions.
	//
	// See https://spec.openapis.org/oas/v3.1.0#specification-extensions.
	Extensions = jsonschema.Extensions
)

// Info provides metadata about the API.
//
// The metadata MAY be used by the clients if needed,
// and MAY be presented in editing or documentation generation tools for convenience.
//
// See https://spec.openapis.org/oas/v3.1.0#info-object.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title" yaml:"title"`
	// A short summary of the API.
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	// A short description of the API.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	// The license information for the exposed API.
	License *License `json:"license,omitempty" yaml:"license,omitempty"`
	// REQUIRED. The version of the OpenAPI document.
	Version string `json:"version" yaml:"version"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}

// Contact information for the exposed API.
//
// See https://spec.openapis.org/oas/v3.1.0#contact-object.
type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// The URL pointing to the contact information.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// The email address of the contact person/organization.
	Email string `json:"email,omitempty" yaml:"email,omitempty"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}

// License information for the exposed API.
//
// See https://spec.openapis.org/oas/v3.1.0#license-object.
type License struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name" yaml:"name"`
	// An SPDX license expression for the API.
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	// A URL to the license used for the API.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}
