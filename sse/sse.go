// Package sse provides support for HTML Server-Sent Events.
//
// See the Server-Sent Events specification:
// https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import (
	"context"
	"iter"
	"time"
)

// State is the SSE stream client state.
type State int

const (
	// StateConnecting indicates that the SSE connection is being established.
	// It can indicate that the stream is reconnecting and waiting for the
	// retry period.
	StateConnecting State = iota
	// StateOpen indicates that the SSE connection is active and receiving events.
	StateOpen
	// StateClosed indicates that the SSE connection has been closed by either
	// the client or the server, and no further events will be received.
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
