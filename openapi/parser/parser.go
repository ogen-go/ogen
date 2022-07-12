package parser

import (
	"github.com/go-faster/errors"
	"gopkg.in/yaml.v3"

	"github.com/ogen-go/ogen"
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
		}{
			requestBodies:   map[string]*openapi.RequestBody{},
			responses:       map[string]*openapi.Response{},
			parameters:      map[string]*openapi.Parameter{},
			headers:         map[string]*openapi.Header{},
			examples:        map[string]*openapi.Example{},
			securitySchemes: map[string]*ogen.SecurityScheme{},
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
			DepthLimit: s.DepthLimit,
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
	)
	for path, item := range p.spec.Paths {
		if err := func() error {
			id, err := pathID(path)
			if err != nil {
				return err
			}

			if _, ok := paths[id]; ok {
				pathLoc := p.rootLoc.Field("paths").Key(path)
				err := errors.Errorf("duplicate path: %q", path)
				return p.wrapLocation("", pathLoc, err)
			}
			paths[id] = struct{}{}

			return p.parsePathItem(path, item, operationIDs, newResolveCtx(p.depthLimit))
		}(); err != nil {
			return errors.Wrapf(err, "path %q", path)
		}
	}
	return nil
}

func (p *parser) parsePathItem(
	path string,
	item *ogen.PathItem,
	operationIDs map[string]struct{},
	ctx *resolveCtx,
) (rerr error) {
	if item == nil {
		return errors.New("pathItem object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(ctx.lastLoc(), item.Locator, rerr)
	}()
	if item.Ref != "" {
		return errors.New("referenced pathItem not supported")
	}

	itemParams, err := p.parseParams(item.Parameters, ctx)
	if err != nil {
		return errors.Wrap(err, "parameters")
	}

	return forEachOps(item, func(method string, op ogen.Operation) error {
		if id := op.OperationID; id != "" {
			if _, ok := operationIDs[id]; ok {
				return errors.Errorf("duplicate operationId: %q", id)
			}
			operationIDs[id] = struct{}{}
		}

		parsedOp, err := p.parseOp(path, method, op, itemParams, ctx)
		if err != nil {
			if op.OperationID != "" {
				return errors.Wrapf(err, "operation %q", op.OperationID)
			}
			return err
		}

		p.operations = append(p.operations, parsedOp)
		return nil
	})
}

func (p *parser) parseOp(
	path, httpMethod string,
	spec ogen.Operation,
	itemParams []*openapi.Parameter,
	ctx *resolveCtx,
) (_ *openapi.Operation, err error) {
	defer func() {
		err = p.wrapLocation(ctx.lastLoc(), spec.Locator, err)
	}()

	op := &openapi.Operation{
		OperationID: spec.OperationID,
		Summary:     spec.Summary,
		Description: spec.Description,
		Deprecated:  spec.Deprecated,
		HTTPMethod:  httpMethod,
		Locator:     spec.Locator,
	}

	opParams, err := p.parseParams(spec.Parameters, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Merge operation parameters with pathItem parameters.
	op.Parameters = mergeParams(opParams, itemParams)

	op.Path, err = parsePath(path, op.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "parse path %q", path)
	}

	if spec.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	op.Responses, err = p.parseResponses(spec.Responses, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "responses")
	}

	if spec.Security != nil {
		// Use operation level security.
		op.Security, err = p.parseSecurityRequirements(spec.Security, ctx)
		if err != nil {
			err = errors.Wrap(err, "security")
			return nil, p.wrapField("security", ctx.lastLoc(), spec.Locator, err)
		}
	} else {
		// Use root level security.
		op.Security, err = p.parseSecurityRequirements(p.spec.Security, ctx)
		if err != nil {
			loc := p.rootLoc.Field("security")
			err = errors.Wrap(err, "security")
			return nil, p.wrapField("security", ctx.lastLoc(), loc, err)
		}
	}

	return op, nil
}

func forEachOps(item *ogen.PathItem, f func(method string, op ogen.Operation) error) error {
	var err error
	handle := func(method string, op *ogen.Operation) {
		if err != nil || op == nil {
			return
		}

		err = f(method, *op)
		if err != nil {
			err = errors.Wrap(err, method)
		}
	}

	handle("get", item.Get)
	handle("put", item.Put)
	handle("post", item.Post)
	handle("delete", item.Delete)
	handle("options", item.Options)
	handle("head", item.Head)
	handle("patch", item.Patch)
	handle("trace", item.Trace)
	return err
}
