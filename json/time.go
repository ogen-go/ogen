package json

import (
	"time"

	"github.com/go-faster/jx"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

// DecodeDate decodes date from json.
func DecodeDate(i *jx.Decoder) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(dateLayout, s)
}

// EncodeDate encodes date to json.
func EncodeDate(s *jx.Encoder, v time.Time) {
	const (
		roundTo  = 8
		length   = len(dateLayout)
		allocate = ((length + roundTo - 1) / roundTo) * roundTo
	)
	b := make([]byte, allocate)
	b = v.AppendFormat(b[:0], dateLayout)
	s.ByteStr(b)
}

// DecodeTime decodes time from json.
func DecodeTime(i *jx.Decoder) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(timeLayout, s)
}

// EncodeTime encodes time to json.
func EncodeTime(s *jx.Encoder, v time.Time) {
	const (
		roundTo  = 8
		length   = len(timeLayout)
		allocate = ((length + roundTo - 1) / roundTo) * roundTo
	)
	b := make([]byte, allocate)
	b = v.AppendFormat(b[:0], timeLayout)
	s.ByteStr(b)
}

// DecodeDateTime decodes date-time from json.
func DecodeDateTime(i *jx.Decoder) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(time.RFC3339, s)
}

// EncodeDateTime encodes date-time to json.
func EncodeDateTime(s *jx.Encoder, v time.Time) {
	const (
		roundTo  = 8
		length   = len(time.RFC3339)
		allocate = ((length + roundTo - 1) / roundTo) * roundTo
	)
	b := make([]byte, allocate)
	b = v.AppendFormat(b[:0], time.RFC3339)
	s.ByteStr(b)
}

// DecodeDuration decodes duration from json.
func DecodeDuration(i *jx.Decoder) (v time.Duration, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.ParseDuration(s)
}

// EncodeDuration encodes duration to json.
func EncodeDuration(s *jx.Encoder, v time.Duration) {
	var buf [32]byte
	w := formatDuration(&buf, v)
	s.ByteStr(buf[w:])
}
