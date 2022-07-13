package parser

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseSecurityScheme(
	scheme *ogen.SecurityScheme,
	ctx *resolveCtx,
) (_ *ogen.SecurityScheme, rerr error) {
	if scheme == nil {
		return nil, errors.New("securityScheme is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(ctx.lastLoc(), scheme.Locator, rerr)
	}()

	if ref := scheme.Ref; ref != "" {
		sch, err := p.resolveSecurityScheme(ref, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "resolve security schema")
		}
		return sch, nil
	}

	if err := func() error {
		switch scheme.Type {
		case "apiKey":
			switch scheme.In {
			case "query", "header", "cookie":
			default:
				err := errors.Errorf(`invalid "in": %q`, scheme.In)
				return p.wrapField("in", ctx.lastLoc(), scheme.Locator, err)
			}
			if scheme.Name == "" {
				err := errors.New(`"name" is required and MUST be a non-empty string`)
				return p.wrapField("name", ctx.lastLoc(), scheme.Locator, err)
			}
			return nil
		case "http":
			// FIXME(tdakkota): spec is not clear about this, it says
			// 	`The values used SHOULD be registered in the IANA Authentication Scheme registry.`
			// 	Probably such validation is too strict.

			// Values from https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml.
			switch strings.ToLower(scheme.Scheme) {
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
				err := errors.Errorf(`invalid "scheme": %q`, scheme.Scheme)
				return p.wrapField("name", ctx.lastLoc(), scheme.Locator, err)
			}
			return nil
		case "mutualTLS":
			return nil
		case "oauth2":
			return p.validateOAuthFlows(scheme.Flows, ctx.lastLoc())
		case "openIdConnect":
			if _, err := url.ParseRequestURI(scheme.OpenIDConnectURL); err != nil {
				err := errors.Wrap(err, `"openIdConnectUrl" MUST be in the form of a URL`)
				return p.wrapField("openIdConnectUrl", ctx.lastLoc(), scheme.Locator, err)
			}
			return nil
		default:
			err := errors.Errorf("unknown security scheme type %q", scheme.Type)
			return p.wrapField("type", ctx.lastLoc(), scheme.Locator, err)
		}
	}(); err != nil {
		return nil, errors.Wrap(err, scheme.Type)
	}

	return scheme, nil
}

func forEachFlow(flows *ogen.OAuthFlows, cb func(flow *ogen.OAuthFlow, authURL, tokenURL bool) error) error {
	for flowName, v := range map[string]struct {
		flow              *ogen.OAuthFlow
		authURL, tokenURL bool
	}{
		"implicit":          {flows.Implicit, true, false},
		"password":          {flows.Password, false, true},
		"clientCredentials": {flows.ClientCredentials, false, true},
		"authorizationCode": {flows.AuthorizationCode, true, true},
	} {
		if v.flow == nil {
			continue
		}
		if err := cb(v.flow, v.authURL, v.tokenURL); err != nil {
			return errors.Wrapf(err, "flow %q", flowName)
		}
	}
	return nil
}

func (p *parser) validateOAuthFlows(flows *ogen.OAuthFlows, loc string) (rerr error) {
	if flows == nil {
		return errors.New("oAuthFlows is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(loc, flows.Locator, rerr)
	}()

	check := func(flow *ogen.OAuthFlow, authURL, tokenURL bool) (rerr error) {
		if flow == nil {
			return nil
		}
		defer func() {
			rerr = p.wrapLocation(loc, flow.Locator, rerr)
		}()

		checkURL := func(name, input string, check bool) error {
			if !check {
				return nil
			}
			if _, err := url.ParseRequestURI(input); err != nil {
				err = errors.Wrapf(err, `%q MUST be in the form of a URL`, name)
				return p.wrapField(name, loc, flow.Locator, err)
			}
			return nil
		}

		if err := checkURL("tokenUrl", flow.TokenURL, tokenURL); err != nil {
			return err
		}
		if err := checkURL("authorizationUrl", flow.AuthorizationURL, authURL); err != nil {
			return err
		}
		if err := checkURL("refreshUrl", flow.RefreshURL, flow.RefreshURL != ""); err != nil {
			return err
		}
		return nil
	}

	return forEachFlow(flows, check)
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
			Locator:          flow.Locator,
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
		Locator:           flows.Locator,
	}
}

func (p *parser) parseSecurityRequirements(
	requirements ogen.SecurityRequirements,
	ctx *resolveCtx,
) ([]openapi.SecurityRequirements, error) {
	result := make([]openapi.SecurityRequirements, 0, len(requirements))
	for _, req := range requirements {
		for requirementName, scopes := range req {
			v, ok := p.refs.securitySchemes[requirementName]
			if !ok {
				return nil, errors.Errorf("unknown security schema %q", requirementName)
			}

			spec, err := p.parseSecurityScheme(v, ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "parse security scheme %q", requirementName)
			}

			if len(scopes) > 0 {
				switch spec.Type {
				case "openIdConnect", "oauth2":
				default:
					return nil, errors.Errorf(`list of scopes MUST be empty for "type" %q`, spec.Type)
				}
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
					Locator:          spec.Locator,
				},
			})
		}
	}

	return result, nil
}
