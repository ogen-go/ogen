package gen

import (
	"reflect"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/internal/oas/parser"
)

type responses struct {
	types     map[string]*ir.Type
	responses map[string]*ir.StatusResponse
}

type Generator struct {
	opt        Options
	operations []*ir.Operation
	types      map[string]*ir.Type
	interfaces map[string]*ir.Type
	refs       responses
	wrapped    responses
	uriTypes   map[*ir.Type]struct{}
	errType    *ir.StatusResponse
}

type Options struct {
	SpecificMethodPath   string
	IgnoreNotImplemented []string
}

func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	operations, err := parser.Parse(spec)
	if err != nil {
		return nil, errors.Wrap(err, "parse")
	}

	g := &Generator{
		opt:        opts,
		types:      map[string]*ir.Type{},
		interfaces: map[string]*ir.Type{},
		refs: responses{
			types:     map[string]*ir.Type{},
			responses: map[string]*ir.StatusResponse{},
		},
		wrapped: responses{
			types:     map[string]*ir.Type{},
			responses: map[string]*ir.StatusResponse{},
		},
		uriTypes: map[*ir.Type]struct{}{},
	}

	if err := g.makeIR(operations); err != nil {
		return nil, errors.Wrap(err, "make ir")
	}
	for _, w := range g.wrapped.types {
		g.saveType(w)
	}
	g.reduce()
	g.wrapGenerics()
	return g, nil
}

func (g *Generator) reduceDefault(ops []*oas.Operation) error {
	if len(ops) < 2 {
		return nil
	}

	// Compare first default response to others.
	first := ops[0]
	if first.Responses == nil || first.Responses.Default == nil {
		return nil
	}
	d := first.Responses.Default
	if d.Ref == "" {
		// Not supported.
		return nil
	}

	for _, spec := range ops[1:] {
		if !reflect.DeepEqual(spec.Responses.Default, d) {
			return nil
		}
	}

	resp, err := g.responseToIR("ErrResp", "reduced default response", d)
	if err != nil {
		return errors.Wrap(err, "default")
	}
	if resp.NoContent != nil || len(resp.Contents) > 1 || resp.Contents[ir.ContentTypeJSON] == nil {
		return errors.Wrap(err, "too complicated to reduce default error")
	}

	g.errType = g.wrapResponseStatusCode(resp)

	return nil
}

func (g *Generator) makeIR(ops []*oas.Operation) error {
	if err := g.reduceDefault(ops); err != nil {
		return errors.Wrap(err, "reduce default")
	}

	for _, spec := range ops {
		op, err := g.generateOperation(spec)
		if err != nil {
			if g.shouldFail(err) {
				return errors.Wrapf(err, "%q: %s",
					spec.Path(),
					strings.ToLower(spec.HTTPMethod),
				)
			}

			continue
		}

		g.operations = append(g.operations, op)
	}

	sort.SliceStable(g.operations, func(i, j int) bool {
		a, b := g.operations[i], g.operations[j]
		return strings.Compare(a.Name, b.Name) < 0
	})

	return nil
}
