// Package parser contains the parser for OpenAPI Spec.
package parser

import (
	"fmt"
	"net/url"

	"github.com/go-faster/errors"
	"golang.org/x/exp/maps"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/uri"
)

type refKey = jsonpointer.RefKey

type parser struct {
	// api spec, immutable.
	spec *ogen.Spec
	// root location of the spec, immutable.
	rootLoc location.Locator
	// parsed version of the spec, immutable.
	version openapi.Version

	// parsed operations.
	operations []*openapi.Operation
	// refs contains lazy-initialized referenced components.
	refs struct {
		requestBodies   map[refKey]*openapi.RequestBody
		responses       map[refKey]*openapi.Response
		parameters      map[refKey]*openapi.Parameter
		headers         map[refKey]*openapi.Header
		examples        map[refKey]*openapi.Example
		securitySchemes map[refKey]*ogen.SecurityScheme
		pathItems       map[refKey]pathItem
	}
	// securitySchemes contains security schemes defined in the root spec.
	securitySchemes map[string]*ogen.SecurityScheme
	// operationIDs holds operation IDs of already parsed operations.
	//
	// Spec says:
	//
	// 	The id MUST be unique among all operations described in the API.
	//
	// Used to detect duplicates.
	operationIDs map[string]location.Pointer

	external              jsonschema.ExternalResolver
	rootURL               *url.URL
	schemas               map[string]resolver
	depthLimit            int
	authenticationSchemes []string
	rootFile              location.File // optional, used for error messages

	schemaParser *jsonschema.Parser
}

// Parse parses raw Spec into
func Parse(spec *ogen.Spec, s Settings) (_ *openapi.API, rerr error) {
	if spec == nil {
		return nil, errors.New("spec is nil")
	}
	spec.Init()

	s.setDefaults()
	p := &parser{
		spec: spec,
		refs: struct {
			requestBodies   map[refKey]*openapi.RequestBody
			responses       map[refKey]*openapi.Response
			parameters      map[refKey]*openapi.Parameter
			headers         map[refKey]*openapi.Header
			examples        map[refKey]*openapi.Example
			securitySchemes map[refKey]*ogen.SecurityScheme
			pathItems       map[refKey]pathItem
		}{
			requestBodies:   map[refKey]*openapi.RequestBody{},
			responses:       map[refKey]*openapi.Response{},
			parameters:      map[refKey]*openapi.Parameter{},
			headers:         map[refKey]*openapi.Header{},
			examples:        map[refKey]*openapi.Example{},
			securitySchemes: map[refKey]*ogen.SecurityScheme{},
			pathItems:       map[refKey]pathItem{},
		},
		securitySchemes: maps.Clone(spec.Components.SecuritySchemes),
		operationIDs:    map[string]location.Pointer{},
		external:        s.External,
		rootURL:         s.RootURL,
		schemas: map[string]resolver{
			"": {
				node: spec.Raw,
				file: s.File,
			},
		},
		depthLimit:            s.DepthLimit,
		authenticationSchemes: s.AuthenticationSchemes,
		rootFile:              s.File,
		schemaParser: jsonschema.NewParser(jsonschema.Settings{
			External: s.External,
			Resolver: componentsResolver{
				components: spec.Components.Schemas,
				root:       jsonschema.NewRootResolver(spec.Raw),
			},
			File:       s.File,
			InferTypes: s.InferTypes,
		}),
	}
	if spec.Raw != nil {
		var loc location.Position
		loc.FromNode(spec.Raw)
		p.rootLoc.SetPosition(loc)
		defer func() {
			rerr = p.wrapLocation(p.rootFile, p.rootLoc, rerr)
		}()
	}

	if err := p.parseVersion(); err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	components, err := p.parseComponents(spec.Components)
	if err != nil {
		return nil, errors.Wrap(err, "parse components")
	}

	if err := p.parsePathItems(); err != nil {
		return nil, errors.Wrap(err, "parse path items")
	}

	servers, err := p.parseServers(p.spec.Servers, p.resolveCtx())
	if err != nil {
		return nil, errors.Wrap(err, "parse servers")
	}

	webhooks, err := p.parseWebhooks(p.spec.Webhooks)
	if err != nil {
		return nil, errors.Wrap(err, "parse webhooks")
	}

	tags := make([]openapi.Tag, len(spec.Tags))
	for i, tag := range spec.Tags {
		tags[i] = openapi.Tag{
			Name:        tag.Name,
			Description: tag.Description,
		}
	}

	return &openapi.API{
		Tags:       tags,
		Version:    p.version,
		Servers:    servers,
		Operations: p.operations,
		Webhooks:   webhooks,
		Components: components,
		Info:       fromOgenInfo(p.spec.Info),
	}, nil
}

func (p *parser) parsePathItems() error {
	var (
		// paths contains simple paths, e.g. "/users/{id}" -> "/users/{}".
		//
		// OpenAPI spec says:
		//
		// 	Templated paths with the same hierarchy but different templated
		//	names MUST NOT exist as they are identical.
		//
		paths = make(map[string]location.Pointer, len(p.spec.Paths))

		file     = p.rootFile
		pathsLoc = p.rootLoc.Field("paths")
	)

	for _, path := range xmaps.SortedKeys(p.spec.Paths) {
		item := p.spec.Paths[path]
		if err := func() error {
			normalized, ok := uri.NormalizeEscapedPath(path)
			if !ok {
				normalized = path
			}

			id, err := pathID(normalized)
			if err != nil {
				return err
			}

			ptr := pathsLoc.Field(path).Pointer(file)
			if existing, ok := paths[id]; ok {
				msg := fmt.Sprintf("duplicate path: %q", path)
				if normalized != path {
					msg += fmt.Sprintf(" (normalized: %q)", normalized)
				}

				me := new(location.MultiError)
				me.ReportPtr(existing, msg)
				me.ReportPtr(ptr, "")
				return me
			}
			paths[id] = ptr

			return nil
		}(); err != nil {
			return p.wrapLocation(file, pathsLoc.Key(path), err)
		}

		up := unparsedPath{
			path: path,
			loc:  pathsLoc.Key(path),
			file: file,
		}
		ops, err := p.parsePathItem(up, item, p.resolveCtx())
		if err != nil {
			err := errors.Wrapf(err, "path %q", path)
			return p.wrapLocation(file, pathsLoc.Field(path), err)
		}
		p.operations = append(p.operations, ops...)
	}
	return nil
}
