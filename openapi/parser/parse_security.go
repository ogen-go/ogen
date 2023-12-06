package parser

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/exp/maps"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseSecurityScheme(
	scheme *ogen.SecurityScheme,
	ctx *jsonpointer.ResolveCtx,
) (_ *ogen.SecurityScheme, rerr error) {
	if scheme == nil {
		return nil, errors.New("securityScheme is empty or null")
	}
	locator := scheme.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()

	if ref := scheme.Ref; ref != "" {
		resolved, err := p.resolveSecurityScheme(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(p.file(ctx), locator, err)
		}
		return resolved, nil
	}

	if err := func() error {
		switch scheme.Type {
		case "apiKey":
			switch scheme.In {
			case "query", "header", "cookie":
			default:
				err := errors.Errorf(`invalid "in": %q`, scheme.In)
				return p.wrapField("in", p.file(ctx), locator, err)
			}
			if scheme.Name == "" {
				err := errors.New(`"name" is required and MUST be a non-empty string`)
				return p.wrapField("name", p.file(ctx), locator, err)
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
				return p.wrapField("scheme", p.file(ctx), locator, err)
			}
			return nil
		case "mutualTLS":
			if err := p.requireMinorVersion("mutualTLS security", 1); err != nil {
				return p.wrapField("type", p.file(ctx), locator, err)
			}
			return nil
		case "oauth2":
			err := p.validateOAuthFlows(scheme.Flows, p.file(ctx))
			return p.wrapField("flows", p.file(ctx), locator, err)
		case "openIdConnect":
			if _, err := url.ParseRequestURI(scheme.OpenIDConnectURL); err != nil {
				err := errors.Wrap(err, `"openIdConnectUrl" MUST be in the form of a URL`)
				return p.wrapField("openIdConnectUrl", p.file(ctx), locator, err)
			}
			return nil
		default:
			err := errors.Errorf("unknown security scheme type %q", scheme.Type)
			return p.wrapField("type", p.file(ctx), locator, err)
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

func (p *parser) validateOAuthFlows(flows *ogen.OAuthFlows, file location.File) (rerr error) {
	if flows == nil {
		return errors.New("oAuthFlows is empty or null")
	}
	locator := flows.Common.Locator
	defer func() {
		rerr = p.wrapLocation(file, locator, rerr)
	}()

	check := func(flow *ogen.OAuthFlow, authURL, tokenURL bool) (rerr error) {
		if flow == nil {
			return nil
		}
		locator := flow.Common.Locator
		defer func() {
			rerr = p.wrapLocation(file, locator, rerr)
		}()

		checkURL := func(name, input string, check bool) error {
			if !check {
				return nil
			}
			if _, err := url.ParseRequestURI(input); err != nil {
				err := errors.Wrapf(err, `%q MUST be in the form of a URL`, name)
				return p.wrapField(name, file, locator, err)
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

func cloneOAuthFlows(flows ogen.OAuthFlows, file location.File) (r openapi.OAuthFlows) {
	cloneFlow := func(flow *ogen.OAuthFlow) *openapi.OAuthFlow {
		if flow == nil {
			return nil
		}
		return &openapi.OAuthFlow{
			AuthorizationURL: flow.AuthorizationURL,
			TokenURL:         flow.TokenURL,
			RefreshURL:       flow.RefreshURL,
			Scopes:           maps.Clone(flow.Scopes),
			Pointer:          flow.Common.Locator.Pointer(file),
		}
	}

	return openapi.OAuthFlows{
		Implicit:          cloneFlow(flows.Implicit),
		Password:          cloneFlow(flows.Password),
		ClientCredentials: cloneFlow(flows.ClientCredentials),
		AuthorizationCode: cloneFlow(flows.AuthorizationCode),
		Pointer:           flows.Common.Locator.Pointer(file),
	}
}

func (p *parser) parseSecurityRequirementScheme(name string, scheme *ogen.SecurityScheme) (openapi.Security, error) {
	// Note that we use root context/file.
	ctx := p.resolveCtx()
	securitySchemesLoc := p.rootLoc.Field("components").Field("securitySchemes")

	spec, err := p.parseSecurityScheme(scheme, ctx)
	if err != nil {
		loc := securitySchemesLoc.Field(name)
		err := errors.Wrapf(err, "parse security scheme %q", name)
		return openapi.Security{}, p.wrapLocation(p.file(ctx), loc, err)
	}

	var flows ogen.OAuthFlows
	if f := spec.Flows; f != nil {
		flows = *f
	}

	var (
		custom  bool
		locator = spec.Common.Locator
	)
	{
		const extensionName = "x-ogen-custom-security"
		if ex, ok := scheme.Common.Extensions[extensionName]; ok {
			if err := ex.Decode(&custom); err != nil {
				err := errors.Wrap(err, "unmarshal value")
				return openapi.Security{}, p.wrapField(extensionName, p.file(ctx), locator, err)
			}
		}
	}

	return openapi.Security{
		Type:                spec.Type,
		Description:         spec.Description,
		Name:                spec.Name,
		In:                  spec.In,
		Scheme:              spec.Scheme,
		BearerFormat:        spec.BearerFormat,
		Flows:               cloneOAuthFlows(flows, p.file(ctx)),
		OpenIDConnectURL:    spec.OpenIDConnectURL,
		XOgenCustomSecurity: custom,
		Pointer:             locator.Pointer(p.file(ctx)),
	}, nil
}

func (p *parser) parseSecurityRequirements(
	requirements ogen.SecurityRequirements,
	locator location.Locator,
	ctx *jsonpointer.ResolveCtx,
) (openapi.SecurityRequirements, error) {
	result := make(openapi.SecurityRequirements, 0, len(requirements))

	for idx, req := range requirements {
		locator := locator.Index(idx)

		var schemes []openapi.SecurityScheme
		for _, name := range xmaps.SortedKeys(req) {
			scopes := req[name]

			v, ok := p.securitySchemes[name]
			if !ok {
				err := errors.Errorf("unknown security scheme %q", name)
				return nil, p.wrapLocation(p.file(ctx), locator.Key(name), err)
			}

			security, err := p.parseSecurityRequirementScheme(name, v)
			if err != nil {
				return nil, err
			}

			schemes = append(schemes, openapi.SecurityScheme{
				Scopes:   scopes,
				Name:     name,
				Security: security,
			})
		}
		result = append(result, openapi.SecurityRequirement{
			Schemes: schemes,
			Pointer: locator.Pointer(p.file(ctx)),
		})
	}

	return result, nil
}
