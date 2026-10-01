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
	// Scopes is a merged list of scopes required by each operation.
	//
	// Scopes from different security requirements are merged into a single list,
	// so the AND/OR relationship between them is lost. Use ScopeSets to get
	// the per-requirement scopes.
	Scopes map[string][]string
	// ScopeSets is a list of scope sets per operation, one per security
	// requirement that uses this security, in spec order.
	//
	// Any one set (logical OR) is sufficient, and all scopes within a set
	// (logical AND) are required.
	ScopeSets map[string][][]string
}

func (s *Security) GoDoc() []string {
	return prettyDoc(s.Description, "")
}

// SecurityCheck is a single security check with the given scope set.
type SecurityCheck struct {
	Security *Security
	// Scopes is a scope set to check.
	Scopes []string
	// ScopeSetIndex is an index of Scopes in Security.ScopeSets of the operation.
	ScopeSetIndex int
}

type SecurityRequirements struct {
	// Securities is a list of securities used by the operation.
	Securities []*Security
	// Requirements is a list of bitsets over Securities.
	Requirements []bitset.Bitset

	// Checks is a list of distinct (security, scope set) pairs used by the operation.
	Checks []SecurityCheck
	// CheckRequirements is a list of bitsets over Checks.
	CheckRequirements []bitset.Bitset
}

// BitArrayLen returns the length for bitset's underlying array.
func (s SecurityRequirements) BitArrayLen() int {
	return bitArrayLen(s.Requirements)
}

// ServerChecks returns a list of checks to perform on server side.
//
// If perRequirement is false, returns a check per security,
// otherwise returns Checks.
func (s SecurityRequirements) ServerChecks(perRequirement bool) []SecurityCheck {
	if perRequirement {
		return s.Checks
	}
	checks := make([]SecurityCheck, len(s.Securities))
	for i, sec := range s.Securities {
		checks[i] = SecurityCheck{Security: sec}
	}
	return checks
}

// ServerRequirements returns a list of bitsets over ServerChecks.
func (s SecurityRequirements) ServerRequirements(perRequirement bool) []bitset.Bitset {
	if perRequirement {
		return s.CheckRequirements
	}
	return s.Requirements
}

// ServerBitArrayLen returns the length for bitset's underlying array of ServerRequirements.
func (s SecurityRequirements) ServerBitArrayLen(perRequirement bool) int {
	return bitArrayLen(s.ServerRequirements(perRequirement))
}

func bitArrayLen(reqs []bitset.Bitset) (r int) {
	for _, req := range reqs {
		if len(req) > r {
			r = len(req)
		}
	}
	return r
}
