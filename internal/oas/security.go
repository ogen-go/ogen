package oas

type SecurityRequirements struct {
	Scopes           []string
	Type             string
	Description      string
	Name             string
	In               string
	Scheme           string
	BearerFormat     string
	Flows            OAuthFlows
	OpenIDConnectURL string
}

// OAuthFlows allows configuration of the supported OAuth Flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow
	Password          *OAuthFlow
	ClientCredentials *OAuthFlow
	AuthorizationCode *OAuthFlow
}

// OAuthFlow is configuration details for a supported OAuth Flow.
type OAuthFlow struct {
	AuthorizationURL string
	TokenURL         string
	RefreshURL       string
	Scopes           map[string]string // name -> description
}
