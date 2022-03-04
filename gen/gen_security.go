package gen

import (
	"net/url"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateSecurityAPIKey(s *ir.Security, spec oas.SecurityRequirements) (*ir.Security, error) {
	security := spec.Security
	if name := security.Name; name == "" {
		return nil, errors.Errorf(`invalid "apiKey" name %q`, name)
	}
	s.Format = ir.APIKeySecurityFormat
	s.ParameterName = security.Name
	s.Type.Fields = append(s.Type.Fields, &ir.Field{
		Name: "APIKey",
		Type: ir.Primitive(ir.String, nil),
	})

	switch in := security.In; in {
	case "query":
		s.Kind = ir.QuerySecurity
	case "header":
		s.Kind = ir.HeaderSecurity
	case "cookie":
		return nil, &ErrNotImplemented{Name: "cookie security"}
	default:
		return nil, errors.Errorf(`unknown "in" value %q`, in)
	}
	return s, nil
}

func (g *Generator) generateSecurityHTTP(s *ir.Security, spec oas.SecurityRequirements) (*ir.Security, error) {
	security := spec.Security
	s.Kind = ir.HeaderSecurity
	s.Type.Fields = append(s.Type.Fields,
		&ir.Field{
			Name: "Username",
			Type: ir.Primitive(ir.String, nil),
		},
		&ir.Field{
			Name: "Password",
			Type: ir.Primitive(ir.String, nil),
		},
	)

	switch scheme := security.Scheme; scheme {
	case "basic":
		s.Format = ir.BasicHTTPSecurityFormat
	default:
		return nil, errors.Wrapf(&ErrNotImplemented{Name: "http security scheme"}, "unsupported scheme %q", scheme)
	}
	return s, nil
}

func (g *Generator) generateSecurity(ctx *genctx, spec oas.SecurityRequirements) (r *ir.Security, rErr error) {
	if sec, ok := g.securities[spec.Name]; ok {
		return sec, nil
	}
	security := spec.Security
	flows := security.Flows

	typeName, err := pascalNonEmpty(spec.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "security name %q", spec.Name)
	}

	t := &ir.Type{
		Kind: ir.KindStruct,
		Name: typeName,
	}
	s := &ir.Security{
		Type:        t,
		Description: security.Description,
	}
	defer func() {
		if rErr == nil {
			if err := ctx.saveType(t); err != nil {
				rErr = err
			}
		}
	}()

	switch typ := security.Type; typ {
	case "openIdConnect":
		if _, err := url.Parse(security.OpenIDConnectURL); err != nil {
			return nil, errors.Wrap(err, "invalid openIdConnectUrl")
		}
		fallthrough
	case "oauth2":
		checkScopes := func(flow *oas.OAuthFlow) error {
			if flow == nil {
				return nil
			}
			for _, scope := range spec.Scopes {
				if _, ok := flow.Scopes[scope]; !ok {
					return errors.Errorf("unknown scope %q", scope)
				}
			}
			return nil
		}

		for flowName, flow := range map[string]*oas.OAuthFlow{
			"implicit":          flows.Implicit,
			"password":          flows.Password,
			"clientCredentials": flows.ClientCredentials,
			"authorizationCode": flows.AuthorizationCode,
		} {
			if err := checkScopes(flow); err != nil {
				return nil, errors.Wrapf(err, "flow %q", flowName)
			}
		}

		return nil, &ErrNotImplemented{Name: "oauth2 security"}
	case "apiKey":
		return g.generateSecurityAPIKey(s, spec)
	case "http":
		return g.generateSecurityHTTP(s, spec)
	case "mutualTLS":
		return nil, &ErrNotImplemented{Name: "mutualTLS security"}
	default:
		return nil, errors.Errorf("unknown security type %q", typ)
	}
}

func (g *Generator) generateSecurities(ctx *genctx, spec []oas.SecurityRequirements) (r []*ir.SecurityRequirement, _ error) {
	for idx, sr := range spec {
		s, err := g.generateSecurity(ctx, sr)
		if err != nil {
			return nil, errors.Wrapf(err, "security %q (index %d)", sr.Name, idx)
		}
		g.securities[sr.Name] = s

		r = append(r, &ir.SecurityRequirement{
			Security: s,
			Spec:     sr,
		})
	}
	return r, nil

}
