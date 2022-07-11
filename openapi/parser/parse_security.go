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
	scopes []string,
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
				return errors.Errorf(`invalid "in": %q`, scheme.In)
			}
			if scheme.Name == "" {
				return errors.New(`"name" is required and MUST be a non-empty string`)
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
				return errors.Errorf(`invalid "scheme": %q`, scheme.Scheme)
			}
			return nil
		case "mutualTLS":
			return nil
		case "oauth2":
			return p.validateOAuthFlows(scopes, scheme.Flows, ctx.lastLoc())
		case "openIdConnect":
			if _, err := url.ParseRequestURI(scheme.OpenIDConnectURL); err != nil {
				return errors.Wrap(err, `"openIdConnectUrl" MUST be in the form of a URL`)
			}
			return nil
		default:
			return errors.Errorf("unknown security scheme type %q", scheme.Type)
		}
	}(); err != nil {
		return nil, errors.Wrap(err, scheme.Type)
	}

	return scheme, nil
}

func (p *parser) validateOAuthFlows(scopes []string, flows *ogen.OAuthFlows, loc string) (rerr error) {
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

			spec, err := p.parseSecurityScheme(v, scopes, ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "parse security scheme %q", requirementName)
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
