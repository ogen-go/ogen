package ir

import (
	"github.com/ogen-go/ogen/internal/bitset"
)

// SecurityKind defines security kind.
type SecurityKind string

const (
	// QuerySecurity is URL query security parameter. Matches "apiKey" type with "in" = "query".
	QuerySecurity SecurityKind = "query"
	// HeaderSecurity is HTTP header security parameter. Matches some "http" schemes and "apiKey" with "in" = "header".
	HeaderSecurity SecurityKind = "header"
	// CookieSecurity is HTTP cookie security parameter. Matches some "http" schemes and "apiKey" with "in" = "cookie".
	CookieSecurity SecurityKind = "cookie"
	// OAuth2Security is special type for OAuth2-based authentication. Matches "oauth2" and "openIdConnect".
	OAuth2Security SecurityKind = "oauth2"
)

// IsQuery whether s is QuerySecurity.
func (s SecurityKind) IsQuery() bool {
	return s == QuerySecurity
}

// IsHeader whether s is HeaderSecurity.
func (s SecurityKind) IsHeader() bool {
	return s == HeaderSecurity
}

// IsCookie whether s is CookieSecurity.
func (s SecurityKind) IsCookie() bool {
	return s == CookieSecurity
}

// IsOAuth2 whether s is OAuth2Security.
func (s SecurityKind) IsOAuth2() bool {
	return s == OAuth2Security
}

// SecurityFormat defines security parameter format.
type SecurityFormat string

const (
	// APIKeySecurityFormat is plain value format.
	APIKeySecurityFormat SecurityFormat = "apiKey"
	// BearerSecurityFormat is Bearer authentication (RFC 6750) format.
	//
	// Unsupported yet.
	BearerSecurityFormat SecurityFormat = "bearer"
	// BasicHTTPSecurityFormat is Basic HTTP authentication (RFC 7617) format.
	BasicHTTPSecurityFormat SecurityFormat = "basic"
	// DigestHTTPSecurityFormat is Digest HTTP authentication (RFC 7616) format.
	//
	// Unsupported yet.
	DigestHTTPSecurityFormat SecurityFormat = "digest"

	// Oauth2SecurityFormat is Oauth2 security format.
	Oauth2SecurityFormat SecurityFormat = "oauth2"

	// CustomSecurityFormat is a user-defined security format.
	CustomSecurityFormat = "x-ogen-custom-security"
)

// IsAPIKeySecurity whether s is APIKeySecurityFormat.
func (s SecurityFormat) IsAPIKeySecurity() bool {
	return s == APIKeySecurityFormat
}

// IsBearerSecurity whether s is BearerSecurityFormat.
func (s SecurityFormat) IsBearerSecurity() bool {
	return s == BearerSecurityFormat
}

// IsBasicHTTPSecurity whether s is BasicHTTPSecurityFormat.
func (s SecurityFormat) IsBasicHTTPSecurity() bool {
	return s == BasicHTTPSecurityFormat
}

// IsDigestHTTPSecurity whether s is DigestHTTPSecurityFormat.
func (s SecurityFormat) IsDigestHTTPSecurity() bool {
	return s == DigestHTTPSecurityFormat
}

// IsOAuth2Security whether s is Oauth2SecurityFormat.
func (s SecurityFormat) IsOAuth2Security() bool {
	return s == Oauth2SecurityFormat
}

// IsCustomSecurity whether s is CustomSecurityFormat.
func (s SecurityFormat) IsCustomSecurity() bool {
	return s == CustomSecurityFormat
}

type Security struct {
	Kind          SecurityKind
	Format        SecurityFormat
	ParameterName string
	Description   string
	Type          *Type
	Scopes        map[string][]string
}

func (s *Security) GoDoc() []string {
	return prettyDoc(s.Description, "")
}

type SecurityRequirements struct {
	Securities   []*Security
	Requirements []bitset.Bitset
}

// BitArrayLen returns the length for bitset's underlying array.
func (s SecurityRequirements) BitArrayLen() (r int) {
	for _, req := range s.Requirements {
		if len(req) > r {
			r = len(req)
		}
	}
	return r
}
