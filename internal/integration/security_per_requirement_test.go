package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_security_per_requirement"
	"github.com/ogen-go/ogen/ogenerrors"
)

type securityCall struct {
	Security string
	Roles    []string
}

// testSecurityPerRequirement grants roles per token.
type testSecurityPerRequirement struct {
	tokens    map[string][]string
	headerKey string
	calls     []securityCall
}

func (t *testSecurityPerRequirement) OptionalSecurity(ctx context.Context) error     { return nil }
func (t *testSecurityPerRequirement) DisjointSecurity(ctx context.Context) error     { return nil }
func (t *testSecurityPerRequirement) IntersectSecurity(ctx context.Context) error    { return nil }
func (t *testSecurityPerRequirement) OrScopesSecurity(ctx context.Context) error     { return nil }
func (t *testSecurityPerRequirement) MixedScopesSecurity(ctx context.Context) error  { return nil }
func (t *testSecurityPerRequirement) Oauth2ScopesSecurity(ctx context.Context) error { return nil }
func (t *testSecurityPerRequirement) CustomSecurity(ctx context.Context) error       { return nil }

// checkRoles returns an error if token is unknown and ErrSkipServerSecurity
// if token does not have all required roles.
func (t *testSecurityPerRequirement) checkRoles(security, token string, required []string) error {
	t.calls = append(t.calls, securityCall{Security: security, Roles: required})
	granted, ok := t.tokens[token]
	if !ok {
		return errors.Errorf("invalid token: %q", token)
	}
	for _, role := range required {
		if !slices.Contains(granted, role) {
			return ogenerrors.ErrSkipServerSecurity
		}
	}
	return nil
}

func (t *testSecurityPerRequirement) HandleBasicAuth(ctx context.Context, operationName string, v api.BasicAuth) (context.Context, error) {
	return ctx, t.checkRoles("BasicAuth", v.Username, v.Roles)
}

func (t *testSecurityPerRequirement) HandleBearerToken(ctx context.Context, operationName string, v api.BearerToken) (context.Context, error) {
	return ctx, t.checkRoles("BearerToken", v.Token, v.Roles)
}

func (t *testSecurityPerRequirement) HandleOAuth2(ctx context.Context, operationName string, v api.OAuth2) (context.Context, error) {
	return ctx, t.checkRoles("OAuth2", v.Token, v.Scopes)
}

func (t *testSecurityPerRequirement) HandleHeaderKey(ctx context.Context, operationName string, v api.HeaderKey) (context.Context, error) {
	t.calls = append(t.calls, securityCall{Security: "HeaderKey", Roles: v.Roles})
	if v.APIKey != t.headerKey {
		return nil, errors.Errorf("invalid api key: %q", v.APIKey)
	}
	return ctx, nil
}

func (t *testSecurityPerRequirement) HandleQueryKey(ctx context.Context, operationName string, v api.QueryKey) (context.Context, error) {
	return ctx, ogenerrors.ErrSkipServerSecurity
}

func (t *testSecurityPerRequirement) HandleCookieKey(ctx context.Context, operationName string, v api.CookieKey) (context.Context, error) {
	return ctx, ogenerrors.ErrSkipServerSecurity
}

func (t *testSecurityPerRequirement) HandleCustom(ctx context.Context, operationName string, v api.Custom) (context.Context, error) {
	return ctx, ogenerrors.ErrSkipServerSecurity
}

func TestSecurityPerRequirement(t *testing.T) {
	const headerKey = "HeaderKey"

	type request struct {
		token     string
		headerKey string
	}
	type testCase struct {
		name      string
		req       request
		wantCode  int
		wantCalls []securityCall
	}
	run := func(path string, tests []testCase) func(t *testing.T) {
		return func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					h := &testSecurityPerRequirement{
						tokens: map[string][]string{
							"none":   {},
							"a":      {"A"},
							"b":      {"B"},
							"bc":     {"B", "C"},
							"read":   {"read"},
							"rw":     {"read", "write"},
							"admin":  {"admin"},
							"abcrwa": {"A", "B", "C", "read", "write", "admin"},
						},
						headerKey: headerKey,
					}
					srv, err := api.NewServer(h, h)
					require.NoError(t, err)

					req := httptest.NewRequest(http.MethodGet, path, nil)
					if tt.req.token != "" {
						req.Header.Set("Authorization", "Bearer "+tt.req.token)
					}
					if tt.req.headerKey != "" {
						req.Header.Set("X-Api-Key", tt.req.headerKey)
					}
					w := httptest.NewRecorder()
					srv.ServeHTTP(w, req)

					require.Equal(t, tt.wantCode, w.Code)
					if tt.wantCalls != nil {
						require.Equal(t, tt.wantCalls, h.calls)
					}
				})
			}
		}
	}

	t.Run("OrScopesSecurity", run("/orScopesSecurity", []testCase{
		{name: "NoAuth", wantCode: http.StatusUnauthorized},
		{name: "InvalidToken", req: request{token: "invalid"}, wantCode: http.StatusUnauthorized},
		{name: "NoRoles", req: request{token: "none"}, wantCode: http.StatusUnauthorized},
		{
			name:     "FirstSet",
			req:      request{token: "a"},
			wantCode: http.StatusOK,
			wantCalls: []securityCall{
				{"BearerToken", []string{"A"}},
				{"BearerToken", []string{"B", "C"}},
			},
		},
		{name: "SecondSet", req: request{token: "bc"}, wantCode: http.StatusOK},
		// B alone is not enough, since second requirement is "B AND C".
		{name: "PartialSet", req: request{token: "b"}, wantCode: http.StatusUnauthorized},
	}))
	t.Run("MixedScopesSecurity", run("/mixedScopesSecurity", []testCase{
		{name: "NoAuth", wantCode: http.StatusUnauthorized},
		// Token satisfies the Bearer part of the first requirement only,
		// it must not satisfy the second requirement.
		{
			name:     "FirstSetWithoutHeaderKey",
			req:      request{token: "a"},
			wantCode: http.StatusUnauthorized,
			wantCalls: []securityCall{
				{"BearerToken", []string{"A"}},
				{"BearerToken", []string{"B"}},
			},
		},
		{
			name:     "FirstSetWithHeaderKey",
			req:      request{token: "a", headerKey: headerKey},
			wantCode: http.StatusOK,
			wantCalls: []securityCall{
				{"BearerToken", []string{"A"}},
				{"HeaderKey", []string{}},
				{"BearerToken", []string{"B"}},
			},
		},
		{name: "SecondSet", req: request{token: "b"}, wantCode: http.StatusOK},
		{name: "HeaderKeyOnly", req: request{headerKey: headerKey}, wantCode: http.StatusUnauthorized},
	}))
	t.Run("Oauth2ScopesSecurity", run("/oauth2ScopesSecurity", []testCase{
		{name: "NoAuth", wantCode: http.StatusUnauthorized},
		{name: "PartialSet", req: request{token: "read"}, wantCode: http.StatusUnauthorized},
		{
			name:     "FirstSet",
			req:      request{token: "rw"},
			wantCode: http.StatusOK,
			wantCalls: []securityCall{
				{"OAuth2", []string{"read", "write"}},
				{"OAuth2", []string{"admin"}},
			},
		},
		{name: "SecondSet", req: request{token: "admin"}, wantCode: http.StatusOK},
	}))
	t.Run("IntersectSecurity", run("/intersectSecurity", []testCase{
		// Empty scope set: any valid token is enough.
		{name: "BearerNoRoles", req: request{token: "none", headerKey: headerKey}, wantCode: http.StatusOK},
		{name: "BearerWithoutHeaderKey", req: request{token: "abcrwa"}, wantCode: http.StatusUnauthorized},
	}))
}

func TestSecurityPerRequirementRoleSets(t *testing.T) {
	// Merged roles and role sets are still available.
	require.Equal(t,
		[]string{"A", "B", "C"},
		api.GetRolesForBearerToken(api.OrScopesSecurityOperation),
	)
	require.Equal(t,
		[][]string{{"A"}, {"B", "C"}},
		api.GetRoleSetsForBearerToken(api.OrScopesSecurityOperation),
	)
}
