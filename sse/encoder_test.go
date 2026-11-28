package sse

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEncoder_Encode(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		want    string
		wantErr bool
	}{
		{
			name:  "data only",
			event: Event{Data: "ok"},
			want:  "data: ok\n\n",
		},
		{
			name:  "all fields",
			event: Event{ID: "10", Type: "update", Data: "ok", Retry: newDuration(3 * time.Second)},
			want:  "id: 10\nevent: update\nretry: 3000\ndata: ok\n\n",
		},
		{
			name:  "multiline data",
			event: Event{Data: "one\ntwo"},
			want:  "data: one\ndata: two\n\n",
		},
		{
			name:  "data newlines are normalized",
			event: Event{Data: "one\r\ntwo\rthree"},
			want:  "data: one\ndata: two\ndata: three\n\n",
		},
		{
			// Only set fields are written, so an empty Data means no data
			// field and the frame is not dispatched by the client.
			name:  "id only frame",
			event: Event{ID: "1"},
			want:  "id: 1\n\n",
		},
		{
			name:  "retry only frame",
			event: Event{Retry: newDuration(3 * time.Second)},
			want:  "retry: 3000\n\n",
		},
		{
			name:  "id and retry only frame",
			event: Event{ID: "1", Retry: newDuration(time.Second)},
			want:  "id: 1\nretry: 1000\n\n",
		},
		{
			name:  "empty event is a blank frame",
			event: Event{},
			want:  "\n",
		},
		{
			name:  "trailing data newline",
			event: Event{Data: "ok\n"},
			want:  "data: ok\ndata:\n\n",
		},
		{
			name:  "data leading space is preserved",
			event: Event{Data: " ok"},
			want:  "data:  ok\n\n",
		},
		{
			name:    "id with newline",
			event:   Event{ID: "1\n2", Data: "ok"},
			wantErr: true,
		},
		{
			name:    "id with nul",
			event:   Event{ID: "1\x002", Data: "ok"},
			wantErr: true,
		},
		{
			name:    "type with newline",
			event:   Event{Type: "up\ndate", Data: "ok"},
			wantErr: true,
		},
		{
			name:    "negative retry",
			event:   Event{Data: "ok", Retry: newDuration(-time.Second)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := NewEncoder(&buf).Encode(tt.event)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", buf.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncoder_Comment(t *testing.T) {
	var buf bytes.Buffer
	e := NewEncoder(&buf)
	if err := e.Comment("keep\nalive"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), ":keep\n:alive\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	// Comment must not dispatch an event.
	if _, err := NewDecoder(strings.NewReader(buf.String()), 0, 0, "", nil).Decode(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestEncoderDecoderRoundTrip(t *testing.T) {
	// Only data-bearing events round-trip, since the decoder dispatches an
	// event only when a data field is present.
	events := []Event{
		{ID: "1", Type: "update", Data: "one\ntwo", Retry: newDuration(time.Second)},
		{Type: DefaultEventType, Data: `{"json":"value"}`},
		{ID: "2", Type: DefaultEventType, Data: ": not a comment"},
	}

	var buf bytes.Buffer
	e := NewEncoder(&buf)
	for _, event := range events {
		if err := e.Encode(event); err != nil {
			t.Fatalf("encode %+v: %v", event, err)
		}
	}

	d := NewDecoder(&buf, 0, 0, "", nil)
	for _, want := range events {
		got, err := d.Decode()
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	}
	if _, err := d.Decode(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

// TestEncoderMetadataOnlyFrame verifies that a retry-only frame updates the
// decoder stream state without being dispatched as an event, which is the
// point of not forcing a data field.
func TestEncoderMetadataOnlyFrame(t *testing.T) {
	var buf bytes.Buffer
	e := NewEncoder(&buf)

	// Retry-only and id-only frames must not carry a data field.
	if err := e.Encode(Event{Retry: newDuration(5 * time.Second)}); err != nil {
		t.Fatalf("encode retry: %v", err)
	}
	if err := e.Encode(Event{ID: "42"}); err != nil {
		t.Fatalf("encode id: %v", err)
	}
	if err := e.Encode(Event{Data: "hello"}); err != nil {
		t.Fatalf("encode data: %v", err)
	}
	if got, want := buf.String(), "retry: 5000\n\nid: 42\n\ndata: hello\n\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	d := NewDecoder(&buf, 0, 0, "", nil)

	// The metadata-only frames are not dispatched, only the data event is.
	event, err := d.Decode()
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if event.Data != "hello" {
		t.Fatalf("got data %q, want %q", event.Data, "hello")
	}

	// The retry and last event ID from the metadata frames were still applied.
	if got := d.Retry(); got != 5*time.Second {
		t.Fatalf("got retry %v, want %v", got, 5*time.Second)
	}
	if got := d.LastEventID(); got != "42" {
		t.Fatalf("got last event ID %q, want %q", got, "42")
	}

	if _, err := d.Decode(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

// TestEncoderBufferShrink verifies that a large event does not permanently
// retain buffer capacity.
func TestEncoderBufferShrink(t *testing.T) {
	var buf bytes.Buffer
	e := NewEncoder(&buf, WithEncoderBufferCap(64))

	large := strings.Repeat("x", 4096)
	if err := e.Encode(Event{Data: large}); err != nil {
		t.Fatalf("encode large: %v", err)
	}
	if cap(e.buf) <= 64 {
		t.Fatalf("expected buffer to grow past 64, got cap %d", cap(e.buf))
	}

	// A subsequent small event shrinks the retained buffer back.
	if err := e.Encode(Event{Data: "small"}); err != nil {
		t.Fatalf("encode small: %v", err)
	}
	if cap(e.buf) > 64 {
		t.Fatalf("expected buffer to shrink to 64, got cap %d", cap(e.buf))
	}
}

func TestCutLine(t *testing.T) {
	tests := []struct {
		in    string
		lines []string
	}{
		{"", []string{""}},
		{"a", []string{"a"}},
		{"a\nb", []string{"a", "b"}},
		{"a\r\nb", []string{"a", "b"}},
		{"a\rb", []string{"a", "b"}},
		{"a\r\nb\rc\nd", []string{"a", "b", "c", "d"}},
		{"a\n", []string{"a", ""}},
		{"a\r\n", []string{"a", ""}},
		{"\n", []string{"", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var got []string
			for rest := tt.in; ; {
				line, tail, more := cutLine(rest)
				got = append(got, line)
				if !more {
					break
				}
				rest = tail
			}
			if !reflect.DeepEqual(got, tt.lines) {
				t.Fatalf("got %q, want %q", got, tt.lines)
			}
		})
	}
}

// unflushableWriter hides the Flush method of the wrapped ResponseWriter.
type unflushableWriter struct {
	http.ResponseWriter
}

func TestEncoderFlush(t *testing.T) {
	t.Run("ResponseWriter", func(t *testing.T) {
		rec := httptest.NewRecorder()
		if err := NewEncoder(rec).Encode(Event{Data: "ok"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !rec.Flushed {
			t.Fatal("expected flush")
		}
	})
	t.Run("NotSupported", func(t *testing.T) {
		rec := httptest.NewRecorder()
		// Flushing is not supported and must not fail the write.
		if err := NewEncoder(unflushableWriter{rec}).Encode(Event{Data: "ok"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got, want := rec.Body.String(), "data: ok\n\n"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
	t.Run("Writer", func(t *testing.T) {
		var buf bytes.Buffer
		if err := NewEncoder(&buf).Encode(Event{Data: "ok"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got, want := buf.String(), "data: ok\n\n"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}
