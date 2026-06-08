package sse

import "github.com/go-faster/errors"

// ErrEventTooLarge reports that an SSE event exceeded the configured size limit.
var ErrEventTooLarge = errors.New("sse event too large")
