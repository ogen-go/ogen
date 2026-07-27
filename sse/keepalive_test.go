package sse

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func identityEncode(e Event) (Event, error) { return e, nil }

// TestKeepAliveEvery verifies that keep-alive comments are written until the
// context is canceled.
func TestKeepAliveEvery(t *testing.T) {
	var buf syncBuffer
	sender := NewSender(NewEncoder(&buf), identityEncode)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- KeepAliveEvery(ctx, sender, time.Millisecond, "ka")
	}()

	// Wait until at least one keep-alive comment has been written.
	deadline := time.After(2 * time.Second)
	for !strings.Contains(buf.String(), ":ka\n") {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for keep-alive comment")
		case <-time.After(time.Millisecond):
		}
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("KeepAliveEvery did not stop after cancel")
	}

	// No comments must be written after the loop returns.
	before := buf.String()
	time.Sleep(20 * time.Millisecond)
	if after := buf.String(); after != before {
		t.Fatalf("keep-alive kept writing after stop: %q -> %q", before, after)
	}
}

// TestKeepAliveEveryStopsOnError verifies that a write error stops the loop and
// is returned.
func TestKeepAliveEveryStopsOnError(t *testing.T) {
	wantErr := errors.New("boom")
	sender := NewSender(NewEncoder(errWriter{err: wantErr}), identityEncode)

	err := KeepAliveEvery(context.Background(), sender, time.Millisecond, "ka")
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}

// TestSenderConcurrent verifies that concurrent Send and Comment calls produce
// well-formed, non-interleaved frames. Run with -race to catch data races on
// the shared encoder buffer.
func TestSenderConcurrent(t *testing.T) {
	var buf syncBuffer
	sender := NewSender(NewEncoder(&buf), identityEncode)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(2)

	const n = 500
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			if err := sender.Send(ctx, Event{Data: "event"}); err != nil {
				t.Errorf("send: %v", err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			if err := sender.Comment(ctx, "keep-alive"); err != nil {
				t.Errorf("comment: %v", err)
				return
			}
		}
	}()
	wg.Wait()

	// Every line must be a well-formed data field, comment, or blank
	// separator. Interleaving would produce garbage lines.
	dataCount, commentCount := 0, 0
	for _, line := range strings.Split(buf.String(), "\n") {
		switch line {
		case "":
		case "data: event":
			dataCount++
		case ":keep-alive":
			commentCount++
		default:
			t.Fatalf("unexpected line %q", line)
		}
	}
	if dataCount != n {
		t.Fatalf("got %d data frames, want %d", dataCount, n)
	}
	if commentCount != n {
		t.Fatalf("got %d comment frames, want %d", commentCount, n)
	}
}

// errWriter fails every write.
type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

// syncBuffer is a concurrency-safe bytes.Buffer for the race test.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
