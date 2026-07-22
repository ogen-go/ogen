// Package sse provides support for HTML Server-Sent Events.
//
// See the Server-Sent Events specification:
// https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import (
	"context"
	"iter"
	"time"

	"github.com/go-faster/errors"
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
	State() (state State, latestErr error)
	Close() error
}

// Sender sends events to an SSE stream.
type Sender[E any] interface {
	// Send encodes event and writes it to the stream, flushing it.
	Send(ctx context.Context, event E) error
	// Comment writes a comment to the stream, e.g. to keep it alive.
	Comment(ctx context.Context, text string) error
}

// NewSender returns a [Sender] writing events to e, encoding them with encode.
func NewSender[E any](e *Encoder, encode func(event E) (Event, error)) Sender[E] {
	return sender[E]{enc: e, encode: encode}
}

type sender[E any] struct {
	enc    *Encoder
	encode func(event E) (Event, error)
}

func (s sender[E]) Send(ctx context.Context, event E) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := s.encode(event)
	if err != nil {
		return errors.Wrap(err, "encode event")
	}
	return s.enc.Encode(raw)
}

func (s sender[E]) Comment(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.enc.Comment(text)
}
