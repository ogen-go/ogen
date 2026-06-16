// Binary ogen-wasm exposes ogen's OpenAPI parser and validator to JavaScript
// via WebAssembly.
//
// It registers a global `ogenValidate(spec string)` function that parses and
// validates an OpenAPI v3 document using the same pipeline as the ogen code
// generator, returning a structured result object.
//
//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"strings"
	"syscall/js"

	"go.uber.org/zap"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/location"
)

func main() {
	js.Global().Set("ogenValidate", js.FuncOf(validate))
	// Signal to the host that the module is ready.
	if ready := js.Global().Get("onOgenReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	// Block forever so the exported function stays callable.
	select {}
}

// result builds the object returned to JavaScript.
func result(ok bool, message, summary string) map[string]any {
	return map[string]any{
		"ok":      ok,
		"error":   message,
		"summary": summary,
	}
}

// validate is the JS entrypoint: ogenValidate(spec) -> {ok, error, summary}.
func validate(this js.Value, args []js.Value) (res any) {
	// Never let a panic escape into the JS runtime.
	defer func() {
		if r := recover(); r != nil {
			res = result(false, fmt.Sprintf("internal error: %v", r), "")
		}
	}()

	if len(args) < 1 || args[0].Type() != js.TypeString {
		return result(false, "expected a single string argument with the OpenAPI document", "")
	}
	data := []byte(args[0].String())
	if len(bytes.TrimSpace(data)) == 0 {
		return result(false, "document is empty", "")
	}

	file := location.NewFile("openapi.yaml", "openapi.yaml", data)

	spec, err := ogen.Parse(data)
	if err != nil {
		return result(false, prettyError(file, err), "")
	}

	opts := gen.Options{
		Parser: gen.ParseOptions{
			// Match the spec as closely as possible without remote refs,
			// which are unavailable in the browser.
			InferSchemaType: true,
			File:            file,
		},
		Generator: gen.GenerateOptions{
			// Focus on schema validity rather than ogen feature coverage:
			// unsupported operations should not fail validation.
			IgnoreNotImplemented: []string{"all"},
		},
		Logger: zap.NewNop(),
	}

	if _, err := gen.NewGenerator(spec, opts); err != nil {
		return result(false, prettyError(file, err), "")
	}

	return result(true, "", summarize(spec))
}

// prettyError renders err with source listing when possible, falling back to
// the plain error string.
func prettyError(file location.File, err error) string {
	var buf bytes.Buffer
	if location.PrintPrettyError(&buf, false, err) {
		return strings.TrimRight(buf.String(), "\n")
	}
	return err.Error()
}

// summarize returns a short human-readable description of a valid spec.
func summarize(spec *ogen.Spec) string {
	var ops int
	for _, item := range spec.Paths {
		if item == nil {
			continue
		}
		for _, op := range []*ogen.Operation{
			item.Get, item.Put, item.Post, item.Delete,
			item.Options, item.Head, item.Patch, item.Trace,
		} {
			if op != nil {
				ops++
			}
		}
	}

	title := spec.Info.Title
	if title == "" {
		title = "(untitled)"
	}
	version := spec.OpenAPI
	if version == "" {
		version = "unknown"
	}
	return fmt.Sprintf("%s — OpenAPI %s, %d path(s), %d operation(s)",
		title, version, len(spec.Paths), ops)
}
