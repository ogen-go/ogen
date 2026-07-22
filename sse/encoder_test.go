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
			name:  "empty data is dispatched",
			event: Event{ID: "1"},
			want:  "id: 1\ndata:\n\n",
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
	events := []Event{
		{ID: "1", Type: "update", Data: "one\ntwo", Retry: newDuration(time.Second)},
		{Type: DefaultEventType, Data: ""},
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
