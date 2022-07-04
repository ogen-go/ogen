package parser

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseSecuritySchema(s *ogen.SecuritySchema, scopes []string, ctx *resolveCtx) (*ogen.SecuritySchema, error) {
	if s == nil {
		return nil, errors.New("securitySchema object is empty or null")
	}

	if ref := s.Ref; ref != "" {
		sch, err := p.resolveSecuritySchema(ref, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "resolve security schema")
		}
		return sch, nil
	}

	if err := func() error {
		switch s.Type {
		case "apiKey":
			switch s.In {
			case "query", "header", "cookie":
			default:
				return errors.Errorf(`invalid "in": %q`, s.In)
			}
			if s.Name == "" {
				return errors.New(`"name" is required and MUST be a non-empty string`)
			}
			return nil
		case "http":
			// FIXME(tdakkota): spec is not clear about this, it says
			// 	`The values used SHOULD be registered in the IANA Authentication Scheme registry.`
			// 	Probably such validation is too strict.

			// Values from https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml.
			switch strings.ToLower(s.Scheme) {
			case "basic",
				"bearer",
				"digest",
				"hoba",
				"mutual",
				"negotiate",
				"oauth",
				"scram-sha-1",
				"scram-sha-256",
				"vapid":
			default:
				return errors.Errorf(`invalid "scheme": %q`, s.Scheme)
			}
			return nil
		case "mutualTLS":
			return nil
		case "oauth2":
			return validateOAuthFlows(scopes, s.Flows)
		case "openIdConnect":
			if _, err := url.ParseRequestURI(s.OpenIDConnectURL); err != nil {
				return errors.Wrap(err, `"openIdConnectUrl" MUST be in the form of a URL`)
			}
			return nil
		default:
			return errors.Errorf("unknown security scheme type %q", s.Type)
		}
	}(); err != nil {
		return nil, errors.Wrap(err, s.Type)
	}

	return s, nil
}

func validateOAuthFlows(scopes []string, flows *ogen.OAuthFlows) error {
	if flows == nil {
		return errors.New("oAuthFlows is empty or null")
	}

	check := func(flow *ogen.OAuthFlow, authURL, tokenURL bool) error {
		if flow == nil {
			return nil
		}

		if tokenURL {
			if _, err := url.ParseRequestURI(flow.TokenURL); err != nil {
				return errors.Wrap(err, `"tokenUrl" MUST be in the form of a URL`)
			}
		}
		if authURL {
			if _, err := url.ParseRequestURI(flow.AuthorizationURL); err != nil {
				return errors.Wrap(err, `"authorizationUrl" MUST be in the form of a URL`)
			}
		}
		if flow.RefreshURL != "" {
			if _, err := url.ParseRequestURI(flow.RefreshURL); err != nil {
				return errors.Wrap(err, `"refreshUrl" MUST be in the form of a URL`)
			}
		}

		for _, scope := range scopes {
			if _, ok := flow.Scopes[scope]; !ok {
				return errors.Errorf("unknown scope %q", scope)
			}
		}
		return nil
	}

	for flowName, v := range map[string]struct {
		flow              *ogen.OAuthFlow
		authURL, tokenURL bool
	}{
		"implicit":          {flows.Implicit, true, false},
		"password":          {flows.Password, false, true},
		"clientCredentials": {flows.ClientCredentials, false, true},
		"authorizationCode": {flows.AuthorizationCode, true, true},
	} {
		if err := check(v.flow, v.authURL, v.tokenURL); err != nil {
			return errors.Wrapf(err, "flow %q", flowName)
		}
	}

	return nil
}

func cloneOAuthFlows(flows ogen.OAuthFlows) (r openapi.OAuthFlows) {
	cloneFlow := func(flow *ogen.OAuthFlow) *openapi.OAuthFlow {
		if flow == nil {
			return nil
		}
		r := &openapi.OAuthFlow{
			AuthorizationURL: flow.AuthorizationURL,
			TokenURL:         flow.TokenURL,
			RefreshURL:       flow.RefreshURL,
			Scopes:           make(map[string]string, len(flow.Scopes)),
		}
		for k, v := range flow.Scopes {
			r.Scopes[k] = v
		}
		return r
	}

	return openapi.OAuthFlows{
		Implicit:          cloneFlow(flows.Implicit),
		Password:          cloneFlow(flows.Password),
		ClientCredentials: cloneFlow(flows.ClientCredentials),
		AuthorizationCode: cloneFlow(flows.AuthorizationCode),
	}
}

func (p *parser) parseSecurityRequirements(requirements ogen.SecurityRequirements) ([]openapi.SecurityRequirements, error) {
	result := make([]openapi.SecurityRequirements, 0, len(requirements))
	for _, req := range requirements {
		for requirementName, scopes := range req {
			v, ok := p.refs.securitySchemes[requirementName]
			if !ok {
				return nil, errors.Errorf("unknown security schema %q", requirementName)
			}

			spec, err := p.parseSecuritySchema(v, newResolveCtx(p.depthLimit))
			if err != nil {
				return nil, errors.Wrapf(err, "resolve %q", requirementName)
			}

			var flows ogen.OAuthFlows
			if spec.Flows != nil {
				flows = *spec.Flows
			}

			result = append(result, openapi.SecurityRequirements{
				Scopes: scopes,
				Name:   requirementName,
				Security: openapi.Security{
					Type:             spec.Type,
					Description:      spec.Description,
					Name:             spec.Name,
					In:               spec.In,
					Scheme:           spec.Scheme,
					BearerFormat:     spec.BearerFormat,
					Flows:            cloneOAuthFlows(flows),
					OpenIDConnectURL: spec.OpenIDConnectURL,
				},
			})
		}
	}

	return result, nil
}
