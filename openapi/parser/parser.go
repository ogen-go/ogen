// Package parser contains the parser for OpenAPI Spec.
package parser

import (
	"sort"

	"github.com/go-faster/errors"
	yaml "github.com/go-faster/yamlx"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

type parser struct {
	// api spec, immutable.
	spec *ogen.Spec
	// root location of the spec, immutable.
	rootLoc location.Locator
	// parsed operations.
	operations []*openapi.Operation
	// refs contains lazy-initialized referenced components.
	refs struct {
		requestBodies   map[string]*openapi.RequestBody
		responses       map[string]*openapi.Response
		parameters      map[string]*openapi.Parameter
		headers         map[string]*openapi.Header
		examples        map[string]*openapi.Example
		securitySchemes map[string]*ogen.SecurityScheme
		pathItems       map[string]pathItem
	}

	external   jsonschema.ExternalResolver
	schemas    map[string]*yaml.Node
	depthLimit int
	filename   string // optional, used for error messages

	schemaParser *jsonschema.Parser
}

// Parse parses raw Spec into
func Parse(spec *ogen.Spec, s Settings) (*openapi.API, error) {
	if spec == nil {
		return nil, errors.New("spec is nil")
	}
	spec.Init()

	s.setDefaults()
	p := &parser{
		spec:       spec,
		operations: nil,
		refs: struct {
			requestBodies   map[string]*openapi.RequestBody
			responses       map[string]*openapi.Response
			parameters      map[string]*openapi.Parameter
			headers         map[string]*openapi.Header
			examples        map[string]*openapi.Example
			securitySchemes map[string]*ogen.SecurityScheme
			pathItems       map[string]pathItem
		}{
			requestBodies:   map[string]*openapi.RequestBody{},
			responses:       map[string]*openapi.Response{},
			parameters:      map[string]*openapi.Parameter{},
			headers:         map[string]*openapi.Header{},
			examples:        map[string]*openapi.Example{},
			securitySchemes: map[string]*ogen.SecurityScheme{},
			pathItems:       map[string]pathItem{},
		},
		external: s.External,
		schemas: map[string]*yaml.Node{
			"": spec.Raw,
		},
		depthLimit: s.DepthLimit,
		filename:   s.Filename,
		schemaParser: jsonschema.NewParser(jsonschema.Settings{
			External: s.External,
			Resolver: componentsResolver{
				components: spec.Components.Schemas,
				root:       jsonschema.NewRootResolver(spec.Raw),
			},
			Filename:   s.Filename,
			InferTypes: s.InferTypes,
		}),
	}
	if spec.Raw != nil {
		var loc location.Location
		loc.FromNode(spec.Raw)
		p.rootLoc.SetLocation(loc)
	}

	for name, s := range spec.Components.SecuritySchemes {
		p.refs.securitySchemes[name] = s
	}

	components, err := p.parseComponents(spec.Components)
	if err != nil {
		return nil, errors.Wrap(err, "parse components")
	}

	if err := p.parsePathItems(); err != nil {
		return nil, errors.Wrap(err, "parse operations")
	}

	return &openapi.API{
		Operations: p.operations,
		Components: components,
	}, nil
}

func (p *parser) parsePathItems() error {
	var (
		// Maps to check for duplicate operationIds and paths.
		operationIDs = map[string]struct{}{}
		// paths contains simple paths, e.g. "/users/{id}" -> "/users/{}".
		//
		// OpenAPI spec says:
		//
		// 	Templated paths with the same hierarchy but different templated
		//	names MUST NOT exist as they are identical.
		//
		paths = map[string]struct{}{}

		pathsLoc = p.rootLoc.Field("paths")
		keys     = make([]string, 0, len(p.spec.Paths))
	)
	for k := range p.spec.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, path := range keys {
		item := p.spec.Paths[path]
		if err := func() error {
			id, err := pathID(path)
			if err != nil {
				return err
			}

			if _, ok := paths[id]; ok {
				return errors.Errorf("duplicate path: %q", path)
			}
			paths[id] = struct{}{}

			return nil
		}(); err != nil {
			return p.wrapLocation("", pathsLoc.Key(path), err)
		}

		ops, err := p.parsePathItem(path, item, operationIDs, jsonpointer.NewResolveCtx(p.depthLimit))
		if err != nil {
			err := errors.Wrapf(err, "path %q", path)
			return p.wrapLocation("", pathsLoc.Field(path), err)
		}
		p.operations = append(p.operations, ops...)
	}
	return nil
}
