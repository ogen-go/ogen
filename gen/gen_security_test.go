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
		},
		{
			Kind:   ir.QuerySecurity,
			Format: ir.APIKeySecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope4"},
			},
		},
		{
			Kind:   ir.HeaderSecurity,
			Format: ir.BasicHTTPSecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope5"},
			},
		},
		{
			Kind:   ir.HeaderSecurity,
			Format: ir.BearerSecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope6"},
			},
		},
		{
			Kind:   "",
			Format: ir.CustomSecurityFormat,
			Scopes: map[string][]string{
				"testOp": {"scope7"},
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
	}
}

func TestGenerateSecuritiesRejectsIncompleteOnlyRequirement(t *testing.T) {
	ctx := &genctx{
		global: newTStorage(),
		local:  newTStorage(),
	}
	gen := &Generator{
		log:        zaptest.NewLogger(t),
		securities: make(map[string]*ir.Security),
		opt: GenerateOptions{
			IgnoreNotImplemented: []string{"mutualTLS security"},
		},
	}

	got, err := gen.generateSecurities(ctx, "testOp", []openapi.SecurityRequirement{
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name: "apiKey",
					Security: openapi.Security{
						Type: "apiKey",
						Name: "X-Api-Key",
						In:   "header",
					},
				},
				{
					Name:     "mutualTLS",
					Security: openapi.Security{Type: "mutualTLS"},
				},
			},
		},
	})
	require.ErrorContains(t, err, "no supported security requirements")
	require.ErrorContains(t, err, "mutualTLS security not implemented")
	require.Empty(t, got.Securities)
	require.Empty(t, got.Requirements)
	require.Empty(t, gen.securities)
	require.Empty(t, ctx.local.types)
}

func TestGenerateSecuritiesSkipsIncompleteAlternativeAtomically(t *testing.T) {
	ctx := &genctx{
		global: newTStorage(),
		local:  newTStorage(),
	}
	gen := &Generator{
		log:        zaptest.NewLogger(t),
		securities: make(map[string]*ir.Security),
		opt: GenerateOptions{
			IgnoreNotImplemented: []string{"mutualTLS security"},
		},
	}

	got, err := gen.generateSecurities(ctx, "testOp", []openapi.SecurityRequirement{
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:     "oauth2",
					Scopes:   []string{"read"},
					Security: openapi.Security{Type: "oauth2"},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name:     "oauth2",
					Scopes:   []string{"admin"},
					Security: openapi.Security{Type: "oauth2"},
				},
				{
					Name: "rolledBackAPIKey",
					Security: openapi.Security{
						Type: "apiKey",
						Name: "X-Rolled-Back",
						In:   "header",
					},
				},
				{
					Name:     "mutualTLS",
					Security: openapi.Security{Type: "mutualTLS"},
				},
			},
		},
		{
			Schemes: []openapi.SecurityScheme{
				{
					Name: "apiKey",
					Security: openapi.Security{
						Type: "apiKey",
						Name: "X-Api-Key",
						In:   "header",
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, got.Securities, 2)
	require.Len(t, got.Requirements, 2)
	require.Equal(t, []byte{0b00000001}, []byte(got.Requirements[0]))
	require.Equal(t, []byte{0b00000010}, []byte(got.Requirements[1]))
	require.Equal(t, []string{"read"}, gen.securities["oauth2"].Scopes["testOp"])
	require.Contains(t, gen.securities, "oauth2")
	require.Contains(t, gen.securities, "apiKey")
	require.NotContains(t, gen.securities, "rolledBackAPIKey")
	require.Contains(t, ctx.local.types, "OAuth2")
	require.Contains(t, ctx.local.types, "ApiKey")
	require.NotContains(t, ctx.local.types, "RolledBackAPIKey")
}

func TestMakeOpsSkipsOperationWithoutSupportedSecurityRequirement(t *testing.T) {
	var skipped []string
	gen := &Generator{
		log:        zaptest.NewLogger(t),
		securities: make(map[string]*ir.Security),
		tstorage:   newTStorage(),
		features:   make(FeatureSet),
		opt: GenerateOptions{
			ConvenientErrors:     -1,
			IgnoreNotImplemented: []string{"mutualTLS security"},
			NotImplementedHook: func(name string, _ error) {
				skipped = append(skipped, name)
			},
		},
	}

	response := func() openapi.Responses {
		return openapi.Responses{
			StatusCode: map[int]*openapi.Response{
				204: {},
			},
		}
	}
	err := gen.makeOps([]*openapi.Operation{
		{
			OperationID: "secured",
			HTTPMethod:  "GET",
			Responses:   response(),
			Security: []openapi.SecurityRequirement{
				{
					Schemes: []openapi.SecurityScheme{
						{
							Name: "apiKey",
							Security: openapi.Security{
								Type: "apiKey",
								Name: "X-Api-Key",
								In:   "header",
							},
						},
						{
							Name:     "mutualTLS",
							Security: openapi.Security{Type: "mutualTLS"},
						},
					},
				},
			},
		},
		{
			OperationID: "health",
			HTTPMethod:  "GET",
			Responses:   response(),
		},
	})
	require.NoError(t, err)
	require.Len(t, gen.operations, 1)
	require.Equal(t, "Health", gen.operations[0].Name)
	require.Empty(t, gen.securities)
	require.Equal(t, []string{"mutualTLS security"}, skipped)
	require.NotContains(t, gen.tstorage.types, "SecuredNoContent")
}
