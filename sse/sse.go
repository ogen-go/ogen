// Package sse provides support for HTML Server-Sent Events.
//
// See the Server-Sent Events specification:
// https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import (
	"context"
	"errors"
	"iter"
	"time"
)

// State is the SSE stream client state.
type State int

const (
	StateConnecting State = iota
	StateOpen
	StateClosed
)

// RetryErrorHandler is called after a retry reconnect attempt fails.
type RetryErrorHandler func(ctx context.Context, connectErr error)

// ClientOptions configures SSE client behavior.
type ClientOptions struct {
	LastEventID       string
	Retry             *time.Duration
	MaxRetries        int
	InitialBufferCap  int
	MaxEventSize      int
	RetryErrorHandler RetryErrorHandler
}

// Client represents SSE client.
type Client[E any] interface {
	Next(ctx context.Context) (E, error)
	All(ctx context.Context) iter.Seq2[E, error]
	State() (state State, connErr error)
	Close() error
}

// IsReconnectable reports whether err should trigger automatic SSE reconnect.
func IsReconnectable(err error) bool {
	return err != nil &&
		!errors.Is(err, ErrEventTooLarge) &&
		!errors.Is(err, ErrStreamClosed) &&
		!errors.Is(err, ErrNoReconnect) &&
		!errors.Is(err, context.Canceled)
}
