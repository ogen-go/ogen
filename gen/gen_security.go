package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/bitset"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateSecurityAPIKey(
	s *ir.Security,
	operationName string,
	spec openapi.SecurityScheme,
) (*ir.Security, error) {
	security := spec.Security
	if name := security.Name; name == "" {
		return nil, errors.Errorf(`invalid "apiKey" name %q`, name)
	}
	s.Format = ir.APIKeySecurityFormat
	s.ParameterName = security.Name
	s.Scopes = map[string][]string{
		operationName: spec.Scopes,
	}

	s.Type.Fields = append(s.Type.Fields,
		&ir.Field{
			Name: "APIKey",
			Type: ir.Primitive(ir.String, nil),
		},
		&ir.Field{
			Name: "Roles",
			Type: ir.Array(ir.Primitive(ir.String, nil), ir.NilOptional, nil),
		},
	)

	switch in := security.In; in {
	case "query":
		s.Kind = ir.QuerySecurity
	case "header":
		s.Kind = ir.HeaderSecurity
		vetHeaderParameterName(g.log, s.ParameterName, spec.Security)
	case "cookie":
		s.Kind = ir.CookieSecurity
	default:
		return nil, errors.Errorf(`unknown "in" value %q`, in)
	}
	return s, nil
}

func (g *Generator) generateSecurityOauth2(
	s *ir.Security,
	operationName string,
	spec openapi.SecurityScheme,
) *ir.Security {
	s.Format = ir.Oauth2SecurityFormat
	s.Kind = ir.HeaderSecurity
	s.Scopes = map[string][]string{
		operationName: spec.Scopes,
	}

	s.Type.Fields = append(s.Type.Fields,
		&ir.Field{
			Name: "Token",
			Type: ir.Primitive(ir.String, nil),
		},
		&ir.Field{
			Name: "Scopes",
			Type: ir.Array(ir.Primitive(ir.String, nil), ir.NilOptional, nil),
		},
	)
	return s
}

func (g *Generator) generateSecurityHTTP(
	s *ir.Security,
	operationName string,
	spec openapi.SecurityScheme,
) (*ir.Security, error) {
	security := spec.Security
	s.Kind = ir.HeaderSecurity
	s.Scopes = map[string][]string{
		operationName: spec.Scopes,
	}

	switch scheme := strings.ToLower(security.Scheme); scheme {
	case "basic":
		s.Format = ir.BasicHTTPSecurityFormat
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
	case "bearer":
		s.Format = ir.BearerSecurityFormat
		s.Type.Fields = append(s.Type.Fields,
			&ir.Field{
				Name: "Token",
				Type: ir.Primitive(ir.String, nil),
			},
		)
	default:
		return nil, errors.Wrapf(&ErrNotImplemented{Name: "http security scheme"}, "unsupported scheme %q", scheme)
	}

	s.Type.Fields = append(s.Type.Fields, &ir.Field{
		Name: "Roles",
		Type: ir.Array(ir.Primitive(ir.String, nil), ir.NilOptional, nil),
	})

	return s, nil
}

func (g *Generator) generateCustomSecurity(
	s *ir.Security,
	operationName string,
	spec openapi.SecurityScheme,
) *ir.Security {
	s.Format = ir.CustomSecurityFormat
	s.Scopes = map[string][]string{
		operationName: spec.Scopes,
	}

	s.Type.Fields = append(s.Type.Fields,
		&ir.Field{
			Name: "Request",
			Type: ir.Pointer(&ir.Type{
				Kind: ir.KindStruct,
				Name: "http.Request",
			}, ir.NilInvalid),
		},
		&ir.Field{
			Name: "Roles",
			Type: ir.Array(ir.Primitive(ir.String, nil), ir.NilOptional, nil),
		},
	)

	return s
}

func (g *Generator) generateSecurity(ctx *genctx, operationName string, spec openapi.SecurityScheme) (r *ir.Security, rErr error) {
	if sec, ok := g.securities[spec.Name]; ok {
		sec.Scopes[operationName] = append(sec.Scopes[operationName], spec.Scopes...)
		return sec, nil
	}
	security := spec.Security

	typeName, err := pascalNonEmpty(spec.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "security name: %q", spec.Name)
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

	if security.XOgenCustomSecurity {
		return g.generateCustomSecurity(s, operationName, spec), nil
	}

	switch typ := security.Type; typ {
	case "apiKey":
		return g.generateSecurityAPIKey(s, operationName, spec)
	case "http":
		return g.generateSecurityHTTP(s, operationName, spec)
	case "oauth2":
		return g.generateSecurityOauth2(s, operationName, spec), nil
	case "openIdConnect", "mutualTLS":
		return nil, &ErrNotImplemented{Name: fmt.Sprintf("%s security", typ)}
	default:
		return nil, errors.Errorf("unknown security type %q", typ)
	}
}

func (g *Generator) generateSecurities(
	ctx *genctx,
	operationName string,
	spec openapi.SecurityRequirements,
) (r ir.SecurityRequirements, _ error) {
	indexes := map[string]int{}
	for idx, requirement := range spec {
		var set bitset.Bitset

		if err := func() error {
			for _, scheme := range requirement.Schemes {
				s, err := g.generateSecurity(ctx, operationName, scheme)
				if err != nil {
					return errors.Wrapf(err, "security scheme %q", scheme.Name)
				}
				g.securities[scheme.Name] = s

				idx, ok := indexes[scheme.Name]
				if !ok {
					r.Securities = append(r.Securities, s)
					idx = len(r.Securities) - 1
					indexes[scheme.Name] = idx
				}
				set.Set(idx, true)
			}
			return nil
		}(); err != nil {
			// Skip entire requirement if at least one security is not implemented.
			err = errors.Wrapf(err, "security requirement %d", idx)
			if err := g.trySkip(err, "Skipping security", requirement); err != nil {
				return r, err
			}
			continue
		}

		r.Requirements = append(r.Requirements, set)
	}
	return r, nil
}
