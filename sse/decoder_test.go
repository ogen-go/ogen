package sse

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func newDuration(d time.Duration) *time.Duration {
	return &d
}

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Event
		wantErr bool
	}{
		{
			name:  "ok for lf",
			input: "id: 10\nevent: update\ndata: one\ndata: two\n\n",
			want: Event{
				ID:   "10",
				Type: "update",
				Data: "one\ntwo",
			},
		},
		{
			name:  "ok for crlf",
			input: "data: ok\r\n\r\n",
			want: Event{
				Type: DefaultEventType,
				Data: "ok",
			},
		},
		{
			name:  "ok for cr",
			input: "data: ok\r\r",
			want: Event{
				Type: DefaultEventType,
				Data: "ok",
			},
		},
		{
			name:  "ok for comments and unknown fields",
			input: ": comment\ndata: ok\nunknown: ignored\nsomething\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "ok",
			},
		},
		{
			name:  "ok for bom",
			input: "\uFEFFdata: ok\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "ok",
			},
		},
		{
			name:  "bom in field value is preserved",
			input: "data: \uFEFFok\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "\uFEFFok",
			},
		},
		{
			name:  "empty data field is dispatched",
			input: "data\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "",
			},
		},
		{
			name:  "only one leading value space is stripped",
			input: "data:  ok\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: " ok",
			},
		},
		{
			name:  "no-colon event field defaults event type",
			input: "event\ndata\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "",
			},
		},
		{
			name:  "invalid id is ignored",
			input: "id: keep\nid: bad\x00id\ndata: ok\n\n",
			want: Event{
				ID:   "keep",
				Type: DefaultEventType,
				Data: "ok",
			},
		},
		{
			name:  "ok for invalid utf8",
			input: "data: \xff\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "\uFFFD",
			},
		},
		{
			name:  "invalid retry is ignored",
			input: "retry: 25\nretry: invalid\ndata: ok\n\n",
			want: Event{
				Type:  DefaultEventType,
				Data:  "ok",
				Retry: newDuration(25 * time.Millisecond),
			},
		},
		{
			name:    "EOF on incomplete event",
			input:   "data: incomplete",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(strings.NewReader(tt.input), 0, 0, "", nil)
			got, gotErr := d.Decode()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Decode() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Decode() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecoder_Decode_MaxEventSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Event
		wantErr bool
	}{
		{
			name:    "error on long data line",
			input:   "data: too-long-line-fr\n\n",
			wantErr: true,
		},
		{
			name:  "ok for accumulated data",
			input: "data: not\ndata: longy\n\n",
			want: Event{
				Type: DefaultEventType,
				Data: "not\nlongy",
			},
		},
		{
			name:    "error on long accumulated data",
			input:   "data: too-long\ndata: too-long\n\n",
			wantErr: true,
		},
		{
			name:    "error on long unknown line",
			input:   "unknown: but-this-is-too-long\n\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const maxEventSize = 20
			d := NewDecoder(strings.NewReader(tt.input), 0, maxEventSize, "", nil)
			got, gotErr := d.Decode()
			if tt.wantErr {
				if !errors.Is(gotErr, ErrEventTooLarge) {
					t.Fatalf("Decode() error = %v, want ErrEventTooLarge", gotErr)
				}
				return
			}
			if gotErr != nil {
				t.Fatalf("Decode() failed: %v", gotErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecoder_Decode_EmptyIDResetsLastEventID(t *testing.T) {
	r := strings.NewReader("id:\n\n")
	const lastEventID = "prev"
	d := NewDecoder(r, 0, 0, lastEventID, nil)
	event, err := d.Decode()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("Decode() error = %v, want io.EOF", err)
	}
	if event != (Event{}) {
		t.Fatalf("Decode() event = %v, want zero Event", event)
	}
	if got, want := d.LastEventID(), ""; got != want {
		t.Fatalf("LastEventID() = %q, want %q", got, want)
	}
}

func TestDecoder_Decode_MaxEventSizeReturnsPartialEvent(t *testing.T) {
	r := strings.NewReader("id: next\nevent: update\nunknown: too-long\n\n")
	const (
		maxEventSize = 25
		lastEventID  = "prev"
	)
	d := NewDecoder(r, 0, maxEventSize, lastEventID, nil)
	event, err := d.Decode()
	if !errors.Is(err, ErrEventTooLarge) {
		t.Fatalf("Decode() error = %v, want ErrEventTooLarge", err)
	}
	if got, want := event.ID, "next"; got != want {
		t.Fatalf("Decode().ID = %q, want %q", got, want)
	}
	if got, want := event.Type, "update"; got != want {
		t.Fatalf("Decode().Type = %q, want %q", got, want)
	}
	if got, want := d.LastEventID(), "next"; got != want {
		t.Fatalf("LastEventID() = %q, want %q", got, want)
	}
}

func TestDecoder_Decode_ContinuesAfterMaxEventSize(t *testing.T) {
	r := strings.NewReader("id: too-big\ndata: oversized\n\ndata: next\n\n")
	const maxEventSize = 20
	d := NewDecoder(r, 0, maxEventSize, "", nil)
	event, err := d.Decode()
	if !errors.Is(err, ErrEventTooLarge) {
		t.Fatalf("Decode() error = %v, want ErrEventTooLarge", err)
	}
	if got, want := event.ID, "too-big"; got != want {
		t.Fatalf("Decode().ID = %q, want %q", got, want)
	}

	event, err = d.Decode()
	if err != nil {
		t.Fatalf("Decode() second error = %v", err)
	}
	if got, want := event.Data, "next"; got != want {
		t.Fatalf("Decode() second Data = %q, want %q", got, want)
	}
	if got, want := event.ID, ""; got != want {
		t.Fatalf("Decode() second ID = %q, want %q", got, want)
	}
}

func TestDecoder_Decode_DrainEventAfterLimitAtLineEnd(t *testing.T) {
	r := strings.NewReader("data: oversized\ndata: still-same-event\n\ndata: next\n\n")
	const maxEventSize = len("data: oversized")
	d := NewDecoder(r, 0, maxEventSize, "", nil)
	event, err := d.Decode()
	if !errors.Is(err, ErrEventTooLarge) {
		t.Fatalf("Decode() error = %v, want ErrEventTooLarge", err)
	}
	if got, want := event.Data, "oversized"; got != want {
		t.Fatalf("Decode().Data = %q, want %q", got, want)
	}

	event, err = d.Decode()
	if err != nil {
		t.Fatalf("Decode() second error = %v", err)
	}
	if got, want := event.Data, "next"; got != want {
		t.Fatalf("Decode() second Data = %q, want %q", got, want)
	}
}

func TestDecoder_Decode_LongBufferedLine(t *testing.T) {
	data := strings.Repeat("a", defaultLineBufferCap+1)
	r := strings.NewReader("data: " + data + "\n\n")
	d := NewDecoder(r, 0, 0, "", nil)
	event, err := d.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got, want := event.Data, data; got != want {
		t.Fatalf("Decode().Data length = %d, want %d", len(got), len(want))
	}
}

func TestDecoder_Decode_SkipsLongBufferedJunkLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "comment",
			input: ":" + strings.Repeat("x", defaultLineBufferCap+1) + "\ndata: ok\n\n",
		},
		{
			name:  "unknown field",
			input: "unknown: " + strings.Repeat("x", defaultLineBufferCap+1) + "\ndata: ok\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(strings.NewReader(tt.input), 0, 0, "", nil)
			event, err := d.Decode()
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got, want := event.Data, "ok"; got != want {
				t.Fatalf("Decode().Data = %q, want %q", got, want)
			}
		})
	}
}

func TestDecoder_Decode_ContinuesAfterMaxEventSizeLongLine(t *testing.T) {
	data := strings.Repeat("a", defaultLineBufferCap+1)
	r := strings.NewReader("data: " + data + "\n\ndata: next\n\n")
	const maxEventSize = 20
	d := NewDecoder(r, 0, maxEventSize, "", nil)
	if event, err := d.Decode(); !errors.Is(err, ErrEventTooLarge) {
		t.Fatalf("Decode() event = %v, error = %v, want ErrEventTooLarge", event, err)
	}

	event, err := d.Decode()
	if err != nil {
		t.Fatalf("Decode() second error = %v", err)
	}
	if got, want := event.Data, "next"; got != want {
		t.Fatalf("Decode() second Data = %q, want %q", got, want)
	}
}
