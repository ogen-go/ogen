package json

import (
	"time"

	"github.com/go-faster/jx"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

func DecodeDate(i *jx.Decoder) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(dateLayout, s)
}

func EncodeDate(s *jx.Encoder, v time.Time) {
	b := make([]byte, 64)
	b = v.AppendFormat(b[:0], dateLayout)
	s.ByteStr(b)
}

func DecodeTime(i *jx.Decoder) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(timeLayout, s)
}

func EncodeTime(s *jx.Encoder, v time.Time) {
	b := make([]byte, 64)
	b = v.AppendFormat(b[:0], timeLayout)
	s.ByteStr(b)
}

func DecodeDateTime(i *jx.Decoder) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(time.RFC3339, s)
}

func EncodeDateTime(s *jx.Encoder, v time.Time) {
	b := make([]byte, 64)
	b = v.AppendFormat(b[:0], time.RFC3339)
	s.ByteStr(b)
}

func DecodeDuration(i *jx.Decoder) (v time.Duration, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.ParseDuration(s)
}

func EncodeDuration(s *jx.Encoder, v time.Duration) {
	s.Str(v.String())
}
