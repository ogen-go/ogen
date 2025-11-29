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

	external                     jsonschema.ExternalResolver
	rootURL                      *url.URL
	schemas                      map[string]resolver
	depthLimit                   int
	authenticationSchemes        []string
	disallowDuplicateMethodPaths bool
	rootFile                     location.File // optional, used for error messages

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
		depthLimit:                   s.DepthLimit,
		authenticationSchemes:        s.AuthenticationSchemes,
		disallowDuplicateMethodPaths: s.DisallowDuplicateMethodPaths,
		rootFile:                     s.File,
		schemaParser: jsonschema.NewParser(jsonschema.Settings{
			External: s.External,
			Resolver: componentsResolver{
				components: spec.Components.Schemas,
				root:       jsonschema.NewRootResolver(spec.Raw),
			},
			File:                      s.File,
			InferTypes:                s.InferTypes,
			AllowCrossTypeConstraints: s.AllowCrossTypeConstraints,
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

// pathEntry tracks a path's location and the HTTP methods defined on it.
type pathEntry struct {
	ptr     location.Pointer
	methods map[string]struct{}
}

// getPathMethods returns the set of HTTP methods defined on a PathItem.
func getPathMethods(item *ogen.PathItem) map[string]struct{} {
	methods := make(map[string]struct{})
	if item == nil {
		return methods
	}
	if item.Get != nil {
		methods["get"] = struct{}{}
	}
	if item.Put != nil {
		methods["put"] = struct{}{}
	}
	if item.Post != nil {
		methods["post"] = struct{}{}
	}
	if item.Delete != nil {
		methods["delete"] = struct{}{}
	}
	if item.Options != nil {
		methods["options"] = struct{}{}
	}
	if item.Head != nil {
		methods["head"] = struct{}{}
	}
	if item.Patch != nil {
		methods["patch"] = struct{}{}
	}
	if item.Trace != nil {
		methods["trace"] = struct{}{}
	}
	return methods
}

// methodsOverlap checks if two method sets have any common methods.
func methodsOverlap(a, b map[string]struct{}) (string, bool) {
	for method := range a {
		if _, ok := b[method]; ok {
			return method, true
		}
	}
	return "", false
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
		// However, when DisallowDuplicateMethodPaths is false (default),
		// we allow duplicate paths if they have different HTTP methods.
		paths = make(map[string]pathEntry, len(p.spec.Paths))

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
			methods := getPathMethods(item)

			if existing, ok := paths[id]; ok {
				// Always error if strict mode is enabled
				if p.disallowDuplicateMethodPaths {
					msg := fmt.Sprintf("duplicate path: %q", path)
					if normalized != path {
						msg += fmt.Sprintf(" (normalized: %q)", normalized)
					}

					me := new(location.MultiError)
					me.ReportPtr(existing.ptr, msg)
					me.ReportPtr(ptr, "")
					return me
				}

				// Check if methods overlap - if so, it's a true conflict
				if method, overlap := methodsOverlap(existing.methods, methods); overlap {
					msg := fmt.Sprintf("duplicate path: %q (method %s conflicts)", path, method)
					if normalized != path {
						msg += fmt.Sprintf(" (normalized: %q)", normalized)
					}

					me := new(location.MultiError)
					me.ReportPtr(existing.ptr, msg)
					me.ReportPtr(ptr, "")
					return me
				}

				// Methods don't overlap - merge the method sets
				for method := range methods {
					existing.methods[method] = struct{}{}
				}
				paths[id] = existing
			} else {
				paths[id] = pathEntry{ptr: ptr, methods: methods}
			}

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
