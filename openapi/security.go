package openapi

import "github.com/ogen-go/ogen/internal/location"

// SecurityRequirement is parsed security requirements.
type SecurityRequirement struct {
	Scopes   []string
	Name     string
	Security Security
}

// Security is parsed security scheme.
type Security struct {
	Type             string
	Description      string
	Name             string
	In               string
	Scheme           string
	BearerFormat     string
	Flows            OAuthFlows
	OpenIDConnectURL string

	location.Pointer `json:"-" yaml:"-"`
}

// OAuthFlows allows configuration of the supported OAuth Flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow
	Password          *OAuthFlow
	ClientCredentials *OAuthFlow
	AuthorizationCode *OAuthFlow

	location.Pointer `json:"-" yaml:"-"`
}

// OAuthFlow is configuration details for a supported OAuth Flow.
type OAuthFlow struct {
	AuthorizationURL string
	TokenURL         string
	RefreshURL       string
	Scopes           map[string]string // name -> description

	location.Pointer `json:"-" yaml:"-"`
}
