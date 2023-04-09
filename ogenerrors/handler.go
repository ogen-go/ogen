package ogenerrors

import (
	"context"
	"net/http"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/validate"
)

// ErrorHandler is an error handler.
type ErrorHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error)

// ErrorCode returns HTTP code for given error.
//
// The default code is http.StatusInternalServerError.
func ErrorCode(err error) int {
	var (
		ctError *validate.InvalidContentTypeError
		ogenErr Error
	)
	switch {
	case errors.Is(err, ht.ErrNotImplemented):
		return http.StatusNotImplemented
	case errors.As(err, &ctError):
		// Takes precedence over Error.
		return http.StatusUnsupportedMediaType
	case errors.As(err, &ogenErr):
		return ogenErr.Code()
	default:
		return http.StatusInternalServerError
	}
}

// DefaultErrorHandler is the default error handler.
func DefaultErrorHandler(_ context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	code := ErrorCode(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	e := jx.GetEncoder()
	e.ObjStart()
	e.FieldStart("error_message")
	e.StrEscape(err.Error())
	e.ObjEnd()

	if _, writeErr := w.Write(e.Bytes()); writeErr != nil {
		http.Error(w, writeErr.Error(), http.StatusInternalServerError)
	}
}
