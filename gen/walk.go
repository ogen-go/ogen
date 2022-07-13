package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
)

func walkResponseTypes(r *ir.Responses, walkFn func(name string, t *ir.Type) (*ir.Type, error)) error {
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
		for contentType, media := range r.Contents {
			typ, err := do(statusText(code), media.Type, contentType)
			if err != nil {
				return errors.Wrapf(err, "%d: %q", code, contentType)
			}
			r.Contents[contentType] = ir.Media{
				Encoding: media.Encoding,
				Type:     typ,
			}
		}
		if r.NoContent != nil {
			typ, err := do(statusText(code), r.NoContent, "")
			if err != nil {
				return errors.Wrapf(err, "%d: no content", code)
			}
			r.NoContent = typ
		}
	}

	if def := r.Default; def != nil {
		for contentType, media := range def.Contents {
			typ, err := do("Default", media.Type, contentType)
			if err != nil {
				return errors.Wrapf(err, "default: %q", contentType)
			}
			def.Contents[contentType] = ir.Media{
				Encoding: media.Encoding,
				Type:     typ,
			}
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

func walkOpTypes(ops []*ir.Operation, walk func(*ir.Type) error) (err error) {
	seen := make(map[*ir.Type]struct{})
	visit := func(t *ir.Type) {
		_, skip := seen[t]
		if err != nil || t == nil || skip {
			return
		}

		seen[t] = struct{}{}
		err = walk(t)
	}

	visitContents := func(c map[ir.ContentType]ir.Media) {
		for _, media := range c {
			visit(media.Type)
		}
	}

	visitResponse := func(r *ir.Response) {
		if r == nil {
			return
		}

		visit(r.NoContent)
		visitContents(r.Contents)
		for _, p := range r.Headers {
			visit(p.Type)
		}
	}

	for _, op := range ops {
		for _, p := range op.Params {
			visit(p.Type)
		}
		if op.Request != nil {
			visit(op.Request.Type)
			visitContents(op.Request.Contents)
		}
		visit(op.Responses.Type)
		for _, r := range op.Responses.StatusCode {
			visitResponse(r)
		}
		visitResponse(op.Responses.Default)
	}

	return err
}
