package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/location"
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

	securityRequirement1 := openapi.SecurityRequirement{
		Schemes: []openapi.SecurityScheme{
			{
				Name:   "oauth2test",
				Scopes: []string{"scope1", "scope2"},
				Security: openapi.Security{
					Type: "oauth2",
				},
			},
		},
		Pointer: location.Pointer{},
	}
	securityRequirement2 := openapi.SecurityRequirement{
		Schemes: []openapi.SecurityScheme{
			{
				Name:   "oauth2test",
				Scopes: []string{"scope3"},
				Security: openapi.Security{
					Type: "oauth2",
				},
			},
		},
		Pointer: location.Pointer{},
	}

	securityRequirements := []openapi.SecurityRequirement{securityRequirement1, securityRequirement2}

	secRequirements, err := gen.generateSecurities(
		ctx,
		"testOp",
		securityRequirements,
	)
	require.NoError(t, err)
	require.Equal(t, len(secRequirements.Securities), 1)
	require.Equal(t, len(secRequirements.Requirements), 2)
	sec := secRequirements.Securities[0]
	require.Equal(t, sec.Kind, ir.HeaderSecurity)
	require.Equal(t, sec.Format, ir.Oauth2SecurityFormat)
	operationScopes, ok := sec.Scopes["testOp"]
	require.True(t, ok)
	require.EqualValues(t, operationScopes, []string{"scope1", "scope2", "scope3"})
}
