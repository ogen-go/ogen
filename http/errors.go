package http

import "github.com/go-faster/errors"

// ErrNotImplemented reports that handler is not implemented.
var ErrNotImplemented = errors.New("not implemented")

// ErrInternalServerErrorResponse reports that response was a internal server error type.
var ErrInternalServerErrorResponse = errors.New("internal server error response")
