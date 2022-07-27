package ogenerrors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-faster/errors"

	ht "github.com/ogen-go/ogen/http"
)

// ErrorHandler is an error handler.
type ErrorHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error)

// ErrorCode returns HTTP code for given error.
//
// The default code is http.StatusInternalServerError.
func ErrorCode(err error) (code int) {
	code = http.StatusInternalServerError

	var ogenErr Error
	switch {
	case errors.Is(err, ht.ErrNotImplemented):
		code = http.StatusNotImplemented
	case errors.As(err, &ogenErr):
		code = ogenErr.Code()
	}

	return code
}

// DefaultErrorHandler is the default error handler.
func DefaultErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	code := ErrorCode(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, writeErr := json.Marshal(struct {
		ErrorMessage string `json:"error_message"`
	}{
		ErrorMessage: err.Error(),
	})
	if writeErr == nil {
		w.Write(data)
	}
}
