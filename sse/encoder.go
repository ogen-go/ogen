package sse

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
)

// Encoder writes an SSE stream as defined by the HTML Standard.
//
// Encoder is not safe for concurrent use.
type Encoder struct {
	w     io.Writer
	flush func() error
	buf   []byte
}

// NewEncoder creates an encoder writing an SSE event stream to w.
//
// Every written event is flushed, so that it is delivered without waiting for
// the underlying buffer to fill. If w is an [http.ResponseWriter], flushing is
// done via [http.ResponseController], unwrapping the writer if needed.
// Otherwise, w is flushed if it implements `Flush() error` or [http.Flusher].
func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{w: w}
	switch v := w.(type) {
	case http.ResponseWriter:
		rc := http.NewResponseController(v)
		e.flush = func() error {
			// Writer may not support flushing, e.g. if it is wrapped by a
			// middleware that does not implement Unwrap. Such stream is still
			// valid, it is just delivered in chunks.
			if err := rc.Flush(); err != nil && !errors.Is(err, http.ErrNotSupported) {
				return err
			}
			return nil
		}
	case interface{ Flush() error }:
		e.flush = v.Flush
	case http.Flusher:
		e.flush = func() error {
			v.Flush()
			return nil
		}
	}
	return e
}

// Encode writes event to the stream and flushes it.
//
// A `data` field is always written, even if [Event.Data] is empty, otherwise
// the event would not be dispatched by the client. Empty [Event.ID] and
// [Event.Type] fields are omitted, making the client use the stream defaults.
func (e *Encoder) Encode(event Event) error {
	if err := checkFieldValue(event.ID); err != nil {
		return errors.Wrap(err, "id")
	}
	if strings.ContainsRune(event.ID, 0) {
		// Such id is ignored by the client, see the SSE specification.
		return errors.New("id must not contain NUL")
	}
	if err := checkFieldValue(event.Type); err != nil {
		return errors.Wrap(err, "event")
	}

	e.buf = e.buf[:0]
	if event.ID != "" {
		e.appendField(fieldID, event.ID)
	}
	if event.Type != "" {
		e.appendField(fieldEvent, event.Type)
	}
	if event.Retry != nil {
		ms := event.Retry.Milliseconds()
		if ms < 0 {
			return errors.Errorf("retry must not be negative, got %v", *event.Retry)
		}
		e.appendField(fieldRetry, strconv.FormatInt(ms, 10))
	}
	// Data is written last, so that all metadata fields are already known to
	// the client when the event is dispatched.
	e.appendData(event.Data)
	// Blank line dispatches the event.
	e.buf = append(e.buf, '\n')

	return e.write()
}

// Comment writes an SSE comment, e.g. to keep the connection alive, and
// flushes the stream.
func (e *Encoder) Comment(text string) error {
	e.buf = e.buf[:0]
	rest := normalizeNewlines(text)
	for {
		line, tail, found := strings.Cut(rest, "\n")
		e.buf = append(e.buf, ':')
		e.buf = append(e.buf, line...)
		e.buf = append(e.buf, '\n')
		if !found {
			break
		}
		rest = tail
	}
	return e.write()
}

// Flush flushes the underlying writer.
func (e *Encoder) Flush() error {
	if e.flush == nil {
		return nil
	}
	return e.flush()
}

func (e *Encoder) write() error {
	if _, err := e.w.Write(e.buf); err != nil {
		return err
	}
	return e.Flush()
}

func (e *Encoder) appendField(name []byte, value string) {
	e.buf = append(e.buf, name...)
	if value == "" {
		e.buf = append(e.buf, ':', '\n')
		return
	}
	// Single space after the colon is stripped by the decoder.
	e.buf = append(e.buf, ':', ' ')
	e.buf = append(e.buf, value...)
	e.buf = append(e.buf, '\n')
}

// appendData writes data as one or more `data` fields, one per line.
func (e *Encoder) appendData(data string) {
	rest := normalizeNewlines(data)
	for {
		line, tail, found := strings.Cut(rest, "\n")
		e.appendField(fieldData, line)
		if !found {
			return
		}
		rest = tail
	}
}

// normalizeNewlines converts CRLF and bare CR line endings to LF, so that
// multi-line values are split into fields consistently.
func normalizeNewlines(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

// checkFieldValue reports whether value can be written as a single-line field.
func checkFieldValue(value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return errors.New("value must not contain newlines")
	}
	return nil
}
