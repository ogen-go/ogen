package gen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/bitset"
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
		// wantChecks is a list of "security:scopes" checks.
		wantChecks []string
		// wantCheckRequirements is a list of check indexes per requirement.
		wantCheckRequirements [][]int
	}{
		{
			name: "OrSameSecurity",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A")}},
				{Schemes: []openapi.SecurityScheme{bearer("B")}},
			},
			wantScopes:            map[string][]string{"bearer": {"A", "B"}},
			wantScopeSets:         map[string][][]string{"bearer": {{"A"}, {"B"}}},
			wantChecks:            []string{"Bearer:A", "Bearer:B"},
			wantCheckRequirements: [][]int{{0}, {1}},
		},
		{
			name: "AndSameSecurity",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A", "B")}},
			},
			wantScopes:            map[string][]string{"bearer": {"A", "B"}},
			wantScopeSets:         map[string][][]string{"bearer": {{"A", "B"}}},
			wantChecks:            []string{"Bearer:A,B"},
			wantCheckRequirements: [][]int{{0}},
		},
		{
			name: "EmptySet",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer()}},
				{Schemes: []openapi.SecurityScheme{bearer("A")}},
			},
			wantScopes:            map[string][]string{"bearer": {"A"}},
			wantScopeSets:         map[string][][]string{"bearer": {{}, {"A"}}},
			wantChecks:            []string{"Bearer:", "Bearer:A"},
			wantCheckRequirements: [][]int{{0}, {1}},
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
			wantChecks:            []string{"Bearer:A", "ApiKey:X", "Bearer:B"},
			wantCheckRequirements: [][]int{{0, 1}, {2}},
		},
		{
			name: "DuplicateSets",
			requirements: []openapi.SecurityRequirement{
				{Schemes: []openapi.SecurityScheme{bearer("A", "B"), apiKey()}},
				{Schemes: []openapi.SecurityScheme{bearer("B", "A")}},
				{Schemes: []openapi.SecurityScheme{apiKey()}},
			},
			wantScopes: map[string][]string{
				"bearer": {"A", "B", "B", "A"},
				"apiKey": nil,
			},
			wantScopeSets: map[string][][]string{
				"bearer": {{"A", "B"}, {"B", "A"}},
				"apiKey": {{}, {}},
			},
			// Same scope sets are checked only once.
			wantChecks:            []string{"Bearer:A,B", "ApiKey:"},
			wantCheckRequirements: [][]int{{0, 1}, {0}, {1}},
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
			wantScopeSets:         map[string][][]string{"bearer": {{"B"}}},
			wantChecks:            []string{"Bearer:B"},
			wantCheckRequirements: [][]int{{0}},
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
			require.Len(t, r.Requirements, len(tt.wantCheckRequirements))

			var checks []string
			for _, c := range r.Checks {
				checks = append(checks, c.Security.Type.Name+":"+strings.Join(c.Scopes, ","))
				// ScopeSetIndex must point to the same scope set.
				require.Equal(t, c.Scopes, c.Security.ScopeSets["testOp"][c.ScopeSetIndex])
			}
			require.Equal(t, tt.wantChecks, checks)

			var wantReqs []bitset.Bitset
			for _, idxs := range tt.wantCheckRequirements {
				var set bitset.Bitset
				for _, idx := range idxs {
					set.Set(idx, true)
				}
				wantReqs = append(wantReqs, set)
			}
			require.Equal(t, wantReqs, r.CheckRequirements)
			require.Equal(t, wantReqs, r.ServerRequirements(true))
			require.Equal(t, r.Requirements, r.ServerRequirements(false))
			require.Len(t, r.ServerChecks(false), len(r.Securities))
		})
	}
}
