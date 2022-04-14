package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseSecuritySchema(s *ogen.SecuritySchema, ctx resolveCtx) (*ogen.SecuritySchema, error) {
	if s == nil {
		return nil, errors.New("security schema must not be null")
	}
	if ref := s.Ref; ref != "" {
		sch, err := p.resolveSecuritySchema(ref, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "resolve security schema")
		}
		return sch, nil
	}
	return s, nil
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

			spec, err := p.parseSecuritySchema(v, resolveCtx{})
			if err != nil {
				return nil, errors.Wrapf(err, "resolve %q", requirementName)
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
					Flows:            cloneOAuthFlows(spec.Flows),
					OpenIDConnectURL: spec.OpenIDConnectURL,
				},
			})
		}
	}

	return result, nil
}
