package ogen

// SecurityScheme defines a security scheme that can be used by the operations.
type SecurityScheme struct {
	Ref string `json:"ref,omitempty"`
	// The type of the security scheme. Valid values are "apiKey", "http", "mutualTLS", "oauth2", "openIdConnect".
	Type string `json:"type"`
	// A description for security scheme. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// The name of the header, query or cookie parameter to be used.
	Name string `json:"name,omitempty"`
	// The location of the API key. Valid values are "query", "header" or "cookie".
	In string `json:"in,omitempty"`
	// The name of the HTTP Authorization scheme to be used in the Authorization header as defined in RFC7235.
	// The values used SHOULD be registered in the IANA Authentication Scheme registry.
	Scheme string `json:"scheme,omitempty"`
	// A hint to the client to identify how the bearer token is formatted. Bearer tokens are usually generated
	// by an authorization server, so this information is primarily for documentation purposes.
	BearerFormat string `json:"bearerFormat,omitempty"`
	// An object containing configuration information for the flow types supported.
	Flows *OAuthFlows `json:"flows,omitempty"`
	// OpenId Connect URL to discover OAuth2 configuration values.
	// This MUST be in the form of a URL. The OpenID Connect standard requires the use of TLS.
	OpenIDConnectURL string `json:"openIdConnectUrl,omitempty"`

	Locator Locator `json:"-" yaml:",inline"`
}

// OAuthFlows allows configuration of the supported OAuth Flows.
type OAuthFlows struct {
	// Configuration for the OAuth Implicit flow.
	Implicit *OAuthFlow `json:"implicit"`
	// Configuration for the OAuth Resource Owner Password flow.
	Password *OAuthFlow `json:"password"`
	// Configuration for the OAuth Client Credentials flow. Previously called application in OpenAPI 2.0.
	ClientCredentials *OAuthFlow `json:"clientCredentials"`
	// Configuration for the OAuth Authorization Code flow. Previously called accessCode in OpenAPI 2.0.
	AuthorizationCode *OAuthFlow `json:"authorizationCode"`

	Locator Locator `json:"-" yaml:",inline"`
}

// OAuthFlow is configuration details for a supported OAuth Flow.
type OAuthFlow struct {
	// The authorization URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	AuthorizationURL string `json:"authorizationUrl"`
	// The token URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
	// The URL to be used for obtaining refresh tokens.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitempty"`
	// The available scopes for the OAuth2 security scheme.
	// A map between the scope name and a short description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`

	Locator Locator `json:"-" yaml:",inline"`
}

// SecurityRequirements lists the required security schemes to execute this operation.
type SecurityRequirements []map[string][]string
