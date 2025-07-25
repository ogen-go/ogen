// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"

	ht "github.com/ogen-go/ogen/http"
)

// UnimplementedHandler is no-op Handler which returns http.ErrNotImplemented.
type UnimplementedHandler struct{}

var _ Handler = UnimplementedHandler{}

// Optional implements optional operation.
//
// GET /optional
func (UnimplementedHandler) Optional(ctx context.Context, params OptionalParams) (r *OptionalOK, _ error) {
	return r, ht.ErrNotImplemented
}

// Required implements required operation.
//
// GET /required
func (UnimplementedHandler) Required(ctx context.Context, params RequiredParams) (r *RequiredOK, _ error) {
	return r, ht.ErrNotImplemented
}
