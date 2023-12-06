package gen

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

func unreachable(v any) string {
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

func isStream(s *jsonschema.Schema) bool {
	// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#considerations-for-file-uploads
	//
	// The Spec says:
	//
	//  Content transferred in binary (octet-stream) MAY omit schema.
	//
	if s == nil {
		return true
	}

	switch s.Type {
	case jsonschema.Empty, jsonschema.String:
	default:
		return false
	}

	// Allow format to be empty, stream body often defined as just string.
	switch s.Format {
	case "", "binary", "byte", "base64":
	default:
		return false
	}
	// TODO(tdakkota): check ContentEncoding field
	return true
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

type position interface {
	Position() (location.Position, bool)
	File() location.File
}

func zapPosition(l position) zap.Field {
	if l == nil {
		return zap.Skip()
	}
	loc, ok := l.Position()
	if !ok {
		return zap.Skip()
	}
	file := l.File()
	return zap.String("at", loc.WithFilename(file.Name))
}
