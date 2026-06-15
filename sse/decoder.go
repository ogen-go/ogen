package sse

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultRetry is the default retry interval used for SSE reconnection.
	DefaultRetry = 3 * time.Second
	// DefaultEventType is the default SSE event type.
	DefaultEventType = "message"
)

const defaultLineBufferCap = 16 << 10 // 16 KiB.

var (
	utf8BOM         = []byte("\uFEFF")
	utf8Replacement = []byte("\uFFFD")

	fieldID    = []byte("id")
	fieldEvent = []byte("event")
	fieldData  = []byte("data")
	fieldRetry = []byte("retry")

	fieldIDPrefix    = []byte("id:")
	fieldEventPrefix = []byte("event:")
	fieldDataPrefix  = []byte("data:")
	fieldRetryPrefix = []byte("retry:")
)

func isSSELinePrefix(line []byte, complete bool) bool {
	return isSSEFieldPrefix(line) ||
		// Complete no-colon field lines are allowed by the SSE standard.
		complete && isSSEField(line)
}

func isSSEField(line []byte) bool {
	return bytes.Equal(line, fieldID) ||
		bytes.Equal(line, fieldEvent) ||
		bytes.Equal(line, fieldData) ||
		bytes.Equal(line, fieldRetry)
}

func isSSEFieldPrefix(line []byte) bool {
	return bytes.HasPrefix(line, fieldIDPrefix) ||
		bytes.HasPrefix(line, fieldEventPrefix) ||
		bytes.HasPrefix(line, fieldDataPrefix) ||
		bytes.HasPrefix(line, fieldRetryPrefix)
}

// Decoder reads an SSE stream as defined by the HTML Standard.
type Decoder struct {
	r *bufio.Reader

	initialLineBufferCap int
	lineBuf              []byte

	// maxEventSize is the maximum number of bytes allowed for a single event.
	// Zero value sets no limit.
	maxEventSize int

	// bomChecked indicates whether the beginning of the stream has already
	// been inspected for a leading UTF-8 BOM.
	// This is required by [9.2.6 Interpreting an event stream].
	//
	// [9.2.6 Interpreting an event stream]: https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation
	bomChecked bool

	lastEventID string
	retry       time.Duration

	eventSize  int
	eventID    string
	eventType  string
	eventData  strings.Builder
	eventRetry *time.Duration
}

// NewDecoder creates a decoder for an SSE event stream.
//
// initialBufferCap controls the reusable line buffer size, zero uses the
// package default. maxEventSize limits the number of bytes accepted for one
// event, zero disables the limit. lastEventID and retry initialize reconnect
// state for a resumed stream.
func NewDecoder(r io.Reader,
	initialBufferCap, maxEventSize int,
	lastEventID string,
	retry *time.Duration,
) *Decoder {
	if initialBufferCap < 0 {
		panic("lineBufferSize must be non-negative")
	}
	if maxEventSize < 0 {
		panic("maxEventSize must be non-negative")
	}
	if initialBufferCap == 0 {
		initialBufferCap = defaultLineBufferCap
	}

	var retryValue = DefaultRetry
	if retry != nil {
		retryValue = *retry
	}

	return &Decoder{
		// r uses a fixed default buffer size.
		r: bufio.NewReaderSize(&newlineNormalizer{r: r},
			defaultLineBufferCap),
		initialLineBufferCap: initialBufferCap,
		lineBuf:              make([]byte, 0, initialBufferCap),
		maxEventSize:         maxEventSize,
		lastEventID:          lastEventID,
		retry:                retryValue,
	}
}

// LastEventID returns the current stream-level last event ID.
//
// It is updated when an id field is parsed and is intended to be sent as the
// Last-Event-ID header on reconnect. It is not the same as [Event.ID], which is
// set only from the id field of the returned event.
func (d *Decoder) LastEventID() string {
	return d.lastEventID
}

// Retry returns the current stream-level retry interval.
//
// It starts with configured default, and is updated when a valid retry field
// is parsed.
func (d *Decoder) Retry() time.Duration {
	return d.retry
}

// Decode reads and returns the next dispatched SSE event.
//
// Events without data are skipped. [Decoder.LastEventID] and [Decoder.Retry]
// are updated as valid id and retry fields are parsed, even before an event
// is returned.
//
// If the stream ends before an event is dispatched, Decode returns io.EOF and
// discards the incomplete event. If maxEventSize is exceeded, Decode returns
// the event fields parsed so far together with [ErrEventTooLarge] and drains
// the rest of that event.
func (d *Decoder) Decode() (Event, error) {
	for {
		line, isNewline, inLine, err := d.readLine()
		if err != nil {
			switch err {
			case io.EOF:
				return Event{}, io.EOF
			case ErrEventTooLarge:
				// On ErrEventTooLarge, Decode returns the event fields decoded
				// before the size limit was reached.
				_ = d.drainEvent(inLine)
				event, _ := d.parseEvent(true)
				return event, ErrEventTooLarge
			default:
				return Event{}, err
			}
		}
		if isNewline {
			if event, ok := d.parseEvent(false); ok {
				return event, nil
			}
			continue
		}
		if line == nil {
			continue
		}
		d.processLine(line)
	}
}

func (d *Decoder) readLine() (line []byte, isNewline, inLine bool, err error) {
	if cap(d.lineBuf) > d.initialLineBufferCap {
		d.lineBuf = make([]byte, 0, d.initialLineBufferCap)
	} else {
		d.lineBuf = d.lineBuf[:0]
	}

	var (
		keepLine bool
		skipLine bool
	)

	for {
		part, err := d.r.ReadSlice('\n')
		complete := err == nil
		switch err {
		case nil:
			part = part[:len(part)-1]
			if len(part) > 0 && part[len(part)-1] == '\r' {
				part = part[:len(part)-1]
			}
		case bufio.ErrBufferFull:
		case io.EOF:
			if err := d.addEventSize(len(part)); err != nil {
				return nil, false, true, err
			}
			return nil, false, false, io.EOF
		default:
			return nil, false, false, err
		}

		if !d.bomChecked {
			d.bomChecked = true
			part = bytes.TrimPrefix(part, utf8BOM)
		}

		if err := d.addEventSize(len(part)); err != nil {
			return nil, false, !complete, err
		}

		// If the line was already classified to be skipped, consume chunks
		// until the line ending without storing them.
		if skipLine {
			if complete {
				return nil, false, false, nil
			}
			continue
		}

		// Empty line that separates events.
		if !keepLine && len(d.lineBuf) == 0 && len(part) == 0 && complete {
			return nil, true, false, nil
		}

		// Comments start with ":" and are skipped.
		if !keepLine && len(d.lineBuf) == 0 && len(part) > 0 && part[0] == ':' {
			if complete {
				return nil, false, false, nil
			}
			skipLine = true
			continue
		}

		// Keep only standard SSE fields. Once the field name cannot be one of
		// the standard ones, skip the rest of the line.
		if !keepLine {
			keepLine = isSSELinePrefix(part, complete)
			if !keepLine {
				if complete {
					return nil, false, false, nil
				}
				skipLine = true
				continue
			}
		}

		if len(d.lineBuf) == 0 && complete {
			return part, false, false, nil
		}
		d.lineBuf = append(d.lineBuf, part...)
		if complete {
			return d.lineBuf, false, false, nil
		}
	}
}

func (d *Decoder) addEventSize(n int) error {
	if d.maxEventSize == 0 {
		return nil
	}
	if n > d.maxEventSize-d.eventSize {
		return ErrEventTooLarge
	}
	d.eventSize += n
	return nil
}

func (d *Decoder) drainEvent(inLine bool) error {
	lineEmpty := !inLine
	for {
		part, err := d.r.ReadSlice('\n')
		switch err {
		case nil:
			part = part[:len(part)-1]
			if len(part) > 0 && part[len(part)-1] == '\r' {
				part = part[:len(part)-1]
			}

			// Check for event separation.
			if lineEmpty && len(part) == 0 {
				return nil
			}
			lineEmpty = true
		case bufio.ErrBufferFull:
			lineEmpty = false
			continue
		case io.EOF:
			return io.EOF
		default:
			return err
		}
	}
}

func (d *Decoder) processLine(line []byte) {
	var (
		field = line
		value []byte
	)
	if before, after, ok := bytes.Cut(line, []byte{':'}); ok {
		field = before
		value = after
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}
	}

	switch {
	case bytes.Equal(field, fieldID):
		if bytes.IndexByte(value, 0) < 0 {
			d.eventID = string(bytes.ToValidUTF8(value, utf8Replacement))
			d.lastEventID = d.eventID
		}
	case bytes.Equal(field, fieldEvent):
		d.eventType = string(bytes.ToValidUTF8(value, utf8Replacement))
	case bytes.Equal(field, fieldData):
		value = bytes.ToValidUTF8(value, utf8Replacement)
		_, _ = d.eventData.Write(value)
		d.eventData.WriteByte('\n')
	case bytes.Equal(field, fieldRetry):
		if len(value) == 0 {
			return
		}
		for _, c := range value {
			if c < '0' || c > '9' {
				return
			}
		}
		ms, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return
		}
		retry := time.Duration(ms) * time.Millisecond
		d.retry = retry
		d.eventRetry = &retry
	}
}

func (d *Decoder) parseEvent(force bool) (event Event, ok bool) {
	defer d.resetEvent()

	if !force && d.eventData.Len() == 0 {
		return Event{}, false
	}

	data := d.eventData.String()
	data = strings.TrimSuffix(data, "\n")

	eventType := d.eventType
	if eventType == "" {
		eventType = DefaultEventType
	}

	return Event{
		ID:    d.eventID,
		Type:  eventType,
		Data:  data,
		Retry: d.eventRetry,
	}, true
}

func (d *Decoder) resetEvent() {
	d.eventSize = 0
	d.eventID = ""
	d.eventType = ""
	d.eventData.Reset()
	d.eventRetry = nil
}
