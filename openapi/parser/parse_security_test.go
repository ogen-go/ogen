package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func Test_validateOAuthFlows(t *testing.T) {
	implicit := func(flow ogen.OAuthFlow) *ogen.OAuthFlows {
		return &ogen.OAuthFlows{Implicit: &flow}
	}
	authCode := func(flow ogen.OAuthFlow) *ogen.OAuthFlows {
		return &ogen.OAuthFlows{AuthorizationCode: &flow}
	}
	password := func(flow ogen.OAuthFlow) *ogen.OAuthFlows {
		return &ogen.OAuthFlows{Password: &flow}
	}
	clientCreds := func(flow ogen.OAuthFlow) *ogen.OAuthFlows {
		return &ogen.OAuthFlows{ClientCredentials: &flow}
	}

	tests := []struct {
		flows   *ogen.OAuthFlows
		wantErr bool
	}{
		// OAuthFlows is required.
		{nil, true},
		{new(ogen.OAuthFlows), false},

		// `implicit` requires `authorizationUrl`.
		{implicit(ogen.OAuthFlow{}), true},
		// `authCode` requires `authorizationUrl` and `tokenUrl`.
		{authCode(ogen.OAuthFlow{}), true},
		// `password` requires `tokenUrl`.
		{password(ogen.OAuthFlow{}), true},
		// `clientCredentials` requires `tokenUrl`.
		{clientCreds(ogen.OAuthFlow{}), true},

		// `authorizationUrl` must be a valid URL.
		{implicit(ogen.OAuthFlow{
			AuthorizationURL: "-",
		}), true},
		// `tokenUrl` must be a valid URL.
		{password(ogen.OAuthFlow{
			TokenURL: "-",
		}), true},

		// `refreshUrl` must be a valid URL.
		{implicit(ogen.OAuthFlow{
			AuthorizationURL: "https://example.com/authorization",
			RefreshURL:       "-",
		}), true},

		// `unknown_scope` must be defined in `flows`.
		{implicit(ogen.OAuthFlow{
			AuthorizationURL: "https://example.com/authorization",
			RefreshURL:       "https://example.com/refresh",
		}), true},
		{implicit(ogen.OAuthFlow{
			AuthorizationURL: "https://example.com/authorization",
			RefreshURL:       "https://example.com/refresh",
			Scopes: map[string]string{
				"unknown_scope": "description",
			},
		}), false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			var p parser
			err := p.validateOAuthFlows(tt.flows, "")
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
		})
	}
}
