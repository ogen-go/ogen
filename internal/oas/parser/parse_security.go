package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
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

func cloneOAuthFlows(flows ogen.OAuthFlows) (r oas.OAuthFlows) {
	cloneFlow := func(flow *ogen.OAuthFlow) *oas.OAuthFlow {
		if flow == nil {
			return nil
		}
		r := &oas.OAuthFlow{
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

	return oas.OAuthFlows{
		Implicit:          cloneFlow(flows.Implicit),
		Password:          cloneFlow(flows.Password),
		ClientCredentials: cloneFlow(flows.ClientCredentials),
		AuthorizationCode: cloneFlow(flows.AuthorizationCode),
	}
}

func (p *parser) parseSecurityRequirements(security ogen.SecurityRequirements) (r []oas.SecurityRequirements, _ error) {
	collect := func(requirements ogen.SecurityRequirements) error {
		for _, req := range requirements {
			for requirementName, scopes := range req {
				v, ok := p.refs.securitySchemes[requirementName]
				if !ok {
					return errors.Errorf("unknown security schema %q", requirementName)
				}

				spec, err := p.parseSecuritySchema(v, resolveCtx{})
				if err != nil {
					return errors.Wrapf(err, "resolve %q", requirementName)
				}

				r = append(r, oas.SecurityRequirements{
					Scopes: scopes,
					Name:   requirementName,
					Security: oas.Security{
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
		return nil
	}

	if err := collect(p.spec.Security); err != nil {
		return nil, errors.Wrap(err, "parse global security")
	}

	if err := collect(security); err != nil {
		return nil, errors.Wrap(err, "parse operation security")
	}

	return r, nil
}
