package gen

import (
	"net/http"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
)

func walkResponseTypes(r *ir.Response, walkFn func(name string, t *ir.Type) (*ir.Type, error)) error {
	do := func(prefix string, t *ir.Type, contentType ir.ContentType) (*ir.Type, error) {
		respName, err := pascal(prefix, string(contentType))
		if err != nil {
			return nil, errors.Wrap(err, "generate name")
		}

		typ, err := walkFn(respName, t)
		if err != nil {
			return nil, errors.Wrap(err, "callback")
		}

		return typ, nil
	}

	for code, r := range r.StatusCode {
		for contentType, t := range r.Contents {
			typ, err := do(http.StatusText(code), t, contentType)
			if err != nil {
				return errors.Wrapf(err, "%d: %q", code, contentType)
			}
			r.Contents[contentType] = typ
		}
		if r.NoContent != nil {
			typ, err := do(http.StatusText(code), r.NoContent, "")
			if err != nil {
				return errors.Wrapf(err, "%d: no content", code)
			}
			r.NoContent = typ
		}
	}

	if def := r.Default; def != nil {
		for contentType, t := range def.Contents {
			typ, err := do("Default", t, contentType)
			if err != nil {
				return errors.Wrapf(err, "default: %q", contentType)
			}
			def.Contents[contentType] = typ
		}
		if def.NoContent != nil {
			typ, err := walkFn("Default", def.NoContent)
			if err != nil {
				return errors.Wrap(err, "callback")
			}
			def.NoContent = typ
		}
	}

	return nil
}
