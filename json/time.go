package json

import (
	"time"

	"github.com/go-faster/jx"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

// DecodeTimeFormat decodes date, time & date-time from json using a custom layout.
func DecodeTimeFormat(d *jx.Decoder, layout string) (v time.Time, err error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(layout, s)
}

// EncodeTimeFormat encodes date, time & date-time to json using a custom layout.
func EncodeTimeFormat(e *jx.Encoder, v time.Time, layout string) {
	const stackThreshold = 64

	var buf []byte
	if len(layout) > stackThreshold {
		buf = make([]byte, len(layout))
	} else {
		// Allocate buf on stack, if we can.
		buf = make([]byte, stackThreshold)
	}

	buf = v.AppendFormat(buf[:0], layout)
	e.ByteStr(buf)
}

// NewTimeDecoder returns a new time decoder using a custom layout.
func NewTimeDecoder(layout string) func(i *jx.Decoder) (time.Time, error) {
	return func(d *jx.Decoder) (time.Time, error) {
		return DecodeTimeFormat(d, layout)
	}
}

// NewTimeEncoder returns a new time encoder using a custom layout.
func NewTimeEncoder(layout string) func(e *jx.Encoder, v time.Time) {
	return func(e *jx.Encoder, v time.Time) {
		EncodeTimeFormat(e, v, layout)
	}
}

// DecodeDate decodes date from json.
func DecodeDate(d *jx.Decoder) (v time.Time, err error) {
	return DecodeTimeFormat(d, dateLayout)
}

// EncodeDate encodes date to json.
func EncodeDate(e *jx.Encoder, v time.Time) {
	EncodeTimeFormat(e, v, dateLayout)
}

// DecodeTime decodes time from json.
func DecodeTime(d *jx.Decoder) (v time.Time, err error) {
	return DecodeTimeFormat(d, timeLayout)
}

// EncodeTime encodes time to json.
func EncodeTime(e *jx.Encoder, v time.Time) {
	EncodeTimeFormat(e, v, timeLayout)
}

// DecodeDateTime decodes date-time from json.
func DecodeDateTime(d *jx.Decoder) (v time.Time, err error) {
	return DecodeTimeFormat(d, time.RFC3339)
}

// EncodeDateTime encodes date-time to json.
func EncodeDateTime(e *jx.Encoder, v time.Time) {
	EncodeTimeFormat(e, v, time.RFC3339)
}

// DecodeDuration decodes duration from json.
func DecodeDuration(d *jx.Decoder) (v time.Duration, err error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	return time.ParseDuration(s)
}

// EncodeDuration encodes duration to json.
func EncodeDuration(e *jx.Encoder, v time.Duration) {
	var buf [32]byte
	w := formatDuration(&buf, v)
	e.ByteStr(buf[w:])
}
