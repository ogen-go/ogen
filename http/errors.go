package http

import "github.com/go-faster/errors"

// ErrNotImplemented reports that handler is not implemented.
var ErrNotImplemented = errors.New("not implemented")

// ErrInternalServerErrorResponse reports that response was a internal server error type.
var ErrInternalServerErrorResponse = errors.New("internal server error response")

// ErrResponseSent reports that the response was already sent to the client, so
// it can't be replaced by an error response.
//
// It is returned when writing of a streamed response, like Server-Sent Events,
// fails after the response headers are sent.
var ErrResponseSent = errors.New("response is already sent")
