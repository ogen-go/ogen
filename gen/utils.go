package gen

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/jsonschema"
)

func unreachable(v interface{}) string {
	return fmt.Sprintf("unreachable: %v", v)
}

func isBinary(s *jsonschema.Schema) bool {
	if s == nil {
		return false
	}

	switch s.Type {
	case jsonschema.Empty, jsonschema.String:
		return s.Format == "binary"
	default:
		return false
	}
}

// isMultipartFile tries to map field to multipart file.
//
// Returns nil type if field is not a file parameter.
func isMultipartFile(ctx *genctx, t *ir.Type, p *jsonschema.Property) (*ir.Type, error) {
	if p == nil || p.Schema == nil {
		return nil, nil
	}
	file := ir.Primitive(ir.File, p.Schema)
	switch {
	case t.IsGeneric():
		v := t.GenericVariant
		if !isBinary(p.Schema) || !v.OnlyOptional() {
			return nil, nil
		}

		r := ir.Generic("MultipartFile", file, v)
		if err := ctx.saveType(r); err != nil {
			return nil, err
		}
		return r, nil
	case t.IsArray():
		if !isBinary(p.Schema.Item) {
			return nil, nil
		}

		r := ir.Array(file, ir.NilNull, p.Schema)
		r.Validators = ir.Validators{
			Array: t.Validators.Array,
		}
		return r, nil
	case t.IsPrimitive():
		if !isBinary(p.Schema) {
			return nil, nil
		}
		return file, nil
	}
	return nil, nil
}

func statusText(code int) string {
	r := http.StatusText(code)
	if r != "" {
		return r
	}
	return fmt.Sprintf("Code%d", code)
}

func (g *Generator) zapLocation(l interface {
	Location() (location.Location, bool)
}) zap.Field {
	loc, ok := l.Location()
	if !ok {
		return zap.Skip()
	}
	return zap.String("at", loc.WithFilename(g.opt.Filename))
}

func (g *schemaGen) zapLocation(l interface {
	Location() (location.Location, bool)
}) zap.Field {
	loc, ok := l.Location()
	if !ok {
		return zap.Skip()
	}
	return zap.String("at", loc.WithFilename(g.filename))
}
