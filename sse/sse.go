// Package sse provides support for HTML Server-Sent Events.
//
// See the Server-Sent Events specification:
// https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import (
	"context"
	"iter"
	"sync"
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
//
// A Sender is safe for concurrent use, so events and keep-alive comments may be
// written from different goroutines.
type Sender[E any] interface {
	// Send encodes event and writes it to the stream, flushing it.
	Send(ctx context.Context, event E) error
	// Comment writes a comment to the stream, e.g. to keep it alive.
	Comment(ctx context.Context, text string) error
}

// NewSender returns a [Sender] writing events to e, encoding them with encode.
func NewSender[E any](e *Encoder, encode func(event E) (Event, error)) Sender[E] {
	return &sender[E]{enc: e, encode: encode}
}

type sender[E any] struct {
	// mu serializes writes to enc, which is not safe for concurrent use.
	mu     sync.Mutex
	enc    *Encoder
	encode func(event E) (Event, error)
}

func (s *sender[E]) Send(ctx context.Context, event E) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := s.encode(event)
	if err != nil {
		return errors.Wrap(err, "encode event")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enc.Encode(raw)
}

func (s *sender[E]) Comment(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enc.Comment(text)
}

// DefaultKeepAliveInterval is the default interval between keep-alive comments,
// as recommended by the SSE specification authoring notes.
//
// [authoring notes]: https://html.spec.whatwg.org/multipage/server-sent-events.html#authoring-notes
const DefaultKeepAliveInterval = 15 * time.Second

// defaultKeepAliveComment is the comment text sent by [KeepAlive].
const defaultKeepAliveComment = "keep-alive"

// KeepAlive writes a keep-alive comment to s every [DefaultKeepAliveInterval]
// until ctx is canceled, then returns nil.
//
// It is used as the default value of the generated KeepAlive field, keeping
// idle connections open. Use [KeepAliveEvery] to configure the interval and
// comment.
func KeepAlive[E any](ctx context.Context, s Sender[E]) error {
	return KeepAliveEvery(ctx, s, DefaultKeepAliveInterval, defaultKeepAliveComment)
}

// KeepAliveEvery writes comment to s every interval until ctx is canceled,
// then returns nil. A write error is returned as-is and stops the loop.
//
// A non-positive interval uses [DefaultKeepAliveInterval].
func KeepAliveEvery[E any](ctx context.Context, s Sender[E], interval time.Duration, comment string) error {
	if interval <= 0 {
		interval = DefaultKeepAliveInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.Comment(ctx, comment); err != nil {
				return err
			}
		}
	}
}
