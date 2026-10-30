package sse

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
)

// defaultEncoderBufferCap is the default retained capacity of the encoder line
// buffer. A single larger event grows the buffer, which is shrunk back to this
// capacity afterwards, so one big event does not permanently retain memory.
const defaultEncoderBufferCap = 4 << 10 // 4 KiB.

// Encoder writes an SSE stream as defined by the HTML Standard.
//
// Encoder is not safe for concurrent use.
type Encoder struct {
	w     io.Writer
	flush func() error

	initialBufferCap int
	buf              []byte
}

// EncoderOption configures an [Encoder].
type EncoderOption func(*Encoder)

// WithEncoderBufferCap sets the retained capacity of the encoder line buffer.
//
// The buffer grows to fit a larger event and is shrunk back to this capacity
// afterwards, so one big event does not permanently retain memory. Zero uses
// the package default.
func WithEncoderBufferCap(bytes int) EncoderOption {
	return func(e *Encoder) {
		if bytes < 0 {
			panic("sse: encoder buffer cap must be non-negative")
		}
		if bytes == 0 {
			bytes = defaultEncoderBufferCap
		}
		e.initialBufferCap = bytes
	}
}

// NewEncoder creates an encoder writing an SSE event stream to w.
//
// Every written event is flushed, so that it is delivered without waiting for
// the underlying buffer to fill. If w is an [http.ResponseWriter], flushing is
// done via [http.ResponseController], unwrapping the writer if needed.
// Otherwise, w is flushed if it implements `Flush() error` or [http.Flusher].
func NewEncoder(w io.Writer, opts ...EncoderOption) *Encoder {
	e := &Encoder{w: w, initialBufferCap: defaultEncoderBufferCap}
	for _, opt := range opts {
		opt(e)
	}
	e.buf = make([]byte, 0, e.initialBufferCap)

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
// Only the fields that are set are written: a [time.Duration] pointer for
// Retry, and non-empty ID, Type and Data. This allows metadata-only frames,
// e.g. a retry-only frame that updates the client reconnect delay without
// dispatching an event. An empty Data means no data field is written, so such
// an event is not dispatched by the client.
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

	e.reset()
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
	if event.Data != "" {
		e.appendData(event.Data)
	}
	// Blank line dispatches the event.
	e.buf = append(e.buf, '\n')

	return e.write()
}

// Comment writes an SSE comment, e.g. to keep the connection alive, and
// flushes the stream.
func (e *Encoder) Comment(text string) error {
	e.reset()
	for rest := text; ; {
		line, tail, more := cutLine(rest)
		e.buf = append(e.buf, ':')
		e.buf = append(e.buf, line...)
		e.buf = append(e.buf, '\n')
		if !more {
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

// reset prepares the buffer for the next frame, shrinking it back to the
// initial capacity if a previous larger frame grew it.
func (e *Encoder) reset() {
	if cap(e.buf) > e.initialBufferCap {
		e.buf = make([]byte, 0, e.initialBufferCap)
		return
	}
	e.buf = e.buf[:0]
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

// appendData writes data as one or more `data` fields, one per line. Line
// endings are normalized to a single field boundary, so that the value is
// round-tripped by the decoder regardless of CRLF, CR or LF input.
func (e *Encoder) appendData(data string) {
	for rest := data; ; {
		line, tail, more := cutLine(rest)
		e.appendField(fieldData, line)
		if !more {
			return
		}
		rest = tail
	}
}

// cutLine returns the first line of s, the remainder after its line ending, and
// whether more lines follow. Lines are terminated by "\n", "\r\n" or "\r",
// matching the newline handling of [newlineNormalizer] in a single pass without
// allocating.
func cutLine(s string) (line, rest string, more bool) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\n':
			return s[:i], s[i+1:], true
		case '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				return s[:i], s[i+2:], true
			}
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}

// checkFieldValue reports whether value can be written as a single-line field.
func checkFieldValue(value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return errors.New("value must not contain newlines")
	}
	return nil
}
