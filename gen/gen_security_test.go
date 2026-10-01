package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func TestGenerateSecurities(t *testing.T) {
	logger := zaptest.NewLogger(t)

	ctx := &genctx{
		global: newTStorage(),
		local:  newTStorage(),
	}

	gen := &Generator{
		log:        logger,
		securities: map[string]*ir.Security{},
	}

	securityRequirements := []openapi.SecurityRequirement{
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:   "oauth2test",
					Scopes: []string{"scope1", "scope2"},
					Security: openapi.Security{
						Type: "oauth2",
					},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:   "oauth2test",
					Scopes: []string{"scope3"},
					Security: openapi.Security{
						Type: "oauth2",
					},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:   "apiKeyTest",
					Scopes: []string{"scope4"},
					Security: openapi.Security{
						Type: "apiKey",
						Name: "key",
						In:   "query",
					},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:   "basicTest",
					Scopes: []string{"scope5"},
					Security: openapi.Security{
						Type:   "http",
						Scheme: "basic",
					},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:   "bearerTest",
					Scopes: []string{"scope6"},
					Security: openapi.Security{
						Type:   "http",
						Scheme: "bearer",
					},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:   "customTest",
					Scopes: []string{"scope7"},
					Security: openapi.Security{
						Type:                "http",
						Scheme:              "myCustomScheme",
						XOgenCustomSecurity: true,
					},
				},
			},
		},
	}

	wantSecurities := []*ir.Security{
		{
			Kind:   ir.HeaderSecurity,
			Format: ir.Oauth2SecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope1", "scope2", "scope3"},
			},
			ScopeSets: map[string][][]string{
				"testOp": {{"scope1", "scope2"}, {"scope3"}},
			},
		},
		{
			Kind:   ir.QuerySecurity,
			Format: ir.APIKeySecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope4"},
			},
			ScopeSets: map[string][][]string{
				"testOp": {{"scope4"}},
			},
		},
		{
			Kind:   ir.HeaderSecurity,
			Format: ir.BasicHTTPSecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope5"},
			},
			ScopeSets: map[string][][]string{
				"testOp": {{"scope5"}},
			},
		},
		{
			Kind:   ir.HeaderSecurity,
			Format: ir.BearerSecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope6"},
			},
			ScopeSets: map[string][][]string{
				"testOp": {{"scope6"}},
			},
		},
		{
			Kind:   "",
			Format: ir.CustomSecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope7"},
			},
			ScopeSets: map[string][][]string{
				"testOp": {{"scope7"}},
			},
		},
	}

	secRequirements, err := gen.generateSecurities(
		ctx,
		"testOp",
		securityRequirements,
	)
	require.NoError(t, err)
	require.Len(t, secRequirements.Securities, len(wantSecurities))
	require.Len(t, secRequirements.Requirements, len(securityRequirements))

	for i, wantSec := range wantSecurities {
		sec := secRequirements.Securities[i]

		require.Equal(t, wantSec.Kind, sec.Kind)
		require.Equal(t, wantSec.Format, sec.Format)
		require.Contains(t, sec.Scopes, "testOp")
		require.Equal(t, wantSec.Scopes["testOp"], sec.Scopes["testOp"])
		require.Equal(t, wantSec.ScopeSets["testOp"], sec.ScopeSets["testOp"])
	}
}

func TestGenerateSecuritiesScopeSets(t *testing.T) {
	bearer := func(scopes ...string) openapi.SecurityScheme {
		return openapi.SecurityScheme{
			Name:   "bearer",
			Scopes: scopes,
			Security: openapi.Security{
				Type:   "http",
				Scheme: "bearer",
			},
		}
	}
	apiKey := func(scopes ...string) openapi.SecurityScheme {
		return openapi.SecurityScheme{
			Name:   "apiKey",
			Scopes: scopes,
			Security: openapi.Security{
				Type: "apiKey",
				Name: "key",
				In:   "header",
			},
		}
	}
	oidc := openapi.SecurityScheme{
		Name: "oidc",
		Security: openapi.Security{
			Type: "openIdConnect",
		},
	}

	tests := []struct {
		name          string
		requirements  []openapi.SecurityRequirement
		wantScopes    map[string][]string
		wantScopeSets map[string][][]string
	}{
		{
			name: "OrSameSecurity",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A")}},
				{Schemes: []openapi.SecurityScheme{bearer("B")}},
			},
			wantScopes:    map[string][]string{"bearer": {"A", "B"}},
			wantScopeSets: map[string][][]string{"bearer": {{"A"}, {"B"}}},
		},
		{
			name: "AndSameSecurity",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A", "B")}},
			},
			wantScopes:    map[string][]string{"bearer": {"A", "B"}},
			wantScopeSets: map[string][][]string{"bearer": {{"A", "B"}}},
		},
		{
			name: "EmptySet",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer()}},
				{Schemes: []openapi.SecurityScheme{bearer("A")}},
			},
			wantScopes:    map[string][]string{"bearer": {"A"}},
			wantScopeSets: map[string][][]string{"bearer": {{}, {"A"}}},
		},
		{
			name: "MixedSecurities",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A"), apiKey("X")}},
				{Schemes: []openapi.SecurityScheme{bearer("B")}},
			},
			wantScopes: map[string][]string{
				"bearer": {"A", "B"},
				"apiKey": {"X"},
			},
			wantScopeSets: map[string][][]string{
				"bearer": {{"A"}, {"B"}},
				"apiKey": {{"X"}},
			},
		},
		{
			name: "SkippedRequirement",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A"), oidc}},
				{Schemes: []openapi.SecurityScheme{bearer("B")}},
			},
			// Merged scopes keep the existing behavior.
			wantScopes: map[string][]string{"bearer": {"A", "B"}},
			// Skipped requirement is not an alternative.
			wantScopeSets: map[string][][]string{"bearer": {{"B"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &genctx{
				global: newTStorage(),
				local:  newTStorage(),
			}
			gen := &Generator{
				log:        zaptest.NewLogger(t),
				securities: map[string]*ir.Security{},
				opt: GenerateOptions{
					IgnoreNotImplemented: []string{"all"},
				},
			}

			r, err := gen.generateSecurities(ctx, "testOp", tt.requirements)
			require.NoError(t, err)

			for name, want := range tt.wantScopeSets {
				sec, ok := gen.securities[name]
				require.True(t, ok, name)
				require.Equal(t, tt.wantScopes[name], sec.Scopes["testOp"], name)
				require.Equal(t, want, sec.ScopeSets["testOp"], name)
			}
			require.Len(t, r.Requirements, len(tt.wantScopeSets["bearer"]))
		})
	}
}
