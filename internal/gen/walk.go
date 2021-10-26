package gen

import (
	"net/http"

	"github.com/ogen-go/ogen/internal/ir"
)

func walkResponseTypes(r *ir.Response, walkFn func(name string, typ *ir.Type) *ir.Type) {
	for code, r := range r.StatusCode {
		for contentType, typ := range r.Contents {
			r.Contents[contentType] = walkFn(pascal(http.StatusText(code), string(contentType)), typ)
		}
		if r.NoContent != nil {
			r.NoContent = walkFn(pascal(http.StatusText(code)), r.NoContent)
		}
	}

	if def := r.Default; def != nil {
		for contentType, typ := range def.Contents {
			def.Contents[contentType] = walkFn(pascal("Default", string(contentType)), typ)
		}
		if def.NoContent != nil {
			def.NoContent = walkFn("Default", def.NoContent)
		}
	}
}
