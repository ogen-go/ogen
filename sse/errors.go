package sse

import (
	_ "net/http" // Used for doc link in comment.

	"github.com/go-faster/errors"
)

var (
	// ErrEventTooLarge reports that an SSE event exceeded the configured size limit.
	ErrEventTooLarge = errors.New("sse: event too large")
	// ErrStreamClosed reports use of a closed SSE stream.
	ErrStreamClosed = errors.New("sse: stream is closed")
	// ErrNoReconnect means the server explicitly requested no reconnect.
	//
	// Returned when the server responds with [http.StatusNoContent].
	ErrNoReconnect = errors.New("sse: no reconnect")
)
