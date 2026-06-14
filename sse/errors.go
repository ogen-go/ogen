package sse

import (
	"context"

	_ "net/http" // Used for doc link in comment.

	"github.com/go-faster/errors"
)

var (
	// ErrEventTooLarge reports that an SSE event exceeded the configured size limit.
	ErrEventTooLarge = errors.New("sse: event too large")
	// ErrStreamClosed reports that the SSE stream was closed by client.
	ErrStreamClosed = errors.New("sse: stream is closed")
	// ErrNoReconnect means the server explicitly requested no reconnect.
	//
	// Returned when the server responds with [http.StatusNoContent].
	ErrNoReconnect = errors.New("sse: no reconnect")
	// ErrMaxRetriesExceeded means the configured reconnect retry budget was exhausted.
	ErrMaxRetriesExceeded = errors.New("sse: max retries exceeded")
)

// IsReconnectableError reports whether err should trigger automatic SSE reconnect.
func IsReconnectableError(err error) bool {
	return err != nil &&
		!errors.Is(err, ErrStreamClosed) &&
		!errors.Is(err, ErrNoReconnect) &&
		!errors.Is(err, ErrMaxRetriesExceeded) &&
		!errors.Is(err, context.Canceled)
}
