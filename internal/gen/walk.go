package gen

import (
	"net/http"

	"github.com/ogen-go/ogen/internal/ir"
)

func walkResponseTypes(r *ir.Response, walkFn func(name string, t *ir.Type) *ir.Type) {
	for code, r := range r.StatusCode {
		for contentType, t := range r.Contents {
			r.Contents[contentType] = walkFn(pascal(http.StatusText(code), string(contentType)), t)
		}
		if r.NoContent != nil {
			r.NoContent = walkFn(pascal(http.StatusText(code)), r.NoContent)
		}
	}

	if def := r.Default; def != nil {
		for contentType, t := range def.Contents {
			def.Contents[contentType] = walkFn(pascal("Default", string(contentType)), t)
		}
		if def.NoContent != nil {
			def.NoContent = walkFn("Default", def.NoContent)
		}
	}
}
