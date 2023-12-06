package openapi

import "github.com/ogen-go/ogen/location"

// SecurityScheme defines one of security schemes used in the security requirement.
type SecurityScheme struct {
	Name     string
	Scopes   []string
	Security Security
}

// SecurityRequirement is parsed security requirement.
type SecurityRequirement struct {
	// Each element needs to be satisfied to authorize the request.
	Schemes []SecurityScheme

	location.Pointer `json:"-" yaml:"-"`
}

// SecurityRequirements are parsed security requirements.
//
// Only one element needs to be satisfied to authorize the request.
type SecurityRequirements []SecurityRequirement

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

	XOgenCustomSecurity bool

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
