package gen

import (
	"net/http"
	"reflect"

	"github.com/ogen-go/ogen/internal/ir"
)

func patchRequestTypes(r *ir.Request, patch func(name string, t *ir.Type) *ir.Type) {
	if r == nil {
		return
	}

	if len(r.Contents) == 1 {
		for ct, t := range r.Contents {
			if !reflect.DeepEqual(r.Type, t) {
				panic("unreachable")
			}

			t = patch(ct.String(), t)
			r.Contents[ct] = t
			r.Type = t
		}
		return
	}

	if !r.Type.IsInterface() {
		panic("unreachable")
	}
	for ct, t := range r.Contents {
		r.Contents[ct] = patch(ct.String(), t)
	}
}

func patchResponseTypes(r *ir.Response, patchFn func(name string, t *ir.Type) *ir.Type) {
	if r == nil {
		return
	}

	var (
		calls       int
		lastT       *ir.Type
		lastPatched *ir.Type
		patch       = func(name string, t *ir.Type) *ir.Type {
			calls++
			lastT = t
			lastPatched = patchFn(name, t)
			return lastPatched
		}
	)

	for code, r := range r.StatusCode {
		if r.Wrapped {
			panic("unreachable")
		}

		for ct, t := range r.Contents {
			name := pascal(http.StatusText(code), ct.String())
			r.Contents[ct] = patch(name, t)
		}
		if t := r.NoContent; t != nil {
			name := pascal(http.StatusText(code))
			r.NoContent = patch(name, t)
		}
	}

	if def := r.Default; def != nil {
		if def.Wrapped {
			for ct, t := range def.Contents {
				name := pascal("Default", ct.String())
				actualT := t.MustField("Response").Type
				patched := patch(name, actualT)
				t.SetFieldType("Response", patched)
				lastT, lastPatched = t, t
			}
			if t := def.NoContent; t != nil {
				name := "Default"
				actualT := t.MustField("Response").Type
				patched := patch(name, actualT)
				t.SetFieldType("Response", patched)
				lastT, lastPatched = t, t
			}
		} else {
			for ct, t := range def.Contents {
				name := pascal("Default", ct.String())
				def.Contents[ct] = patch(name, t)
			}
			if t := def.NoContent; t != nil {
				name := pascal("Default")
				def.NoContent = patch(name, t)
			}
		}
	}

	if calls == 1 {
		if !reflect.DeepEqual(lastT, r.Type) {
			panic("unreachable")
		}

		r.Type = lastPatched
		return
	}

	if !r.Type.IsInterface() {
		panic("unreachable")
	}
}
