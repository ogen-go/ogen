package json

import (
	"time"

	"github.com/go-faster/jx"
)

// DecodeUnixSeconds decodes unix-seconds from json.
func DecodeUnixSeconds(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

// EncodeUnixSeconds encodes unix-seconds to json.
func EncodeUnixSeconds(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.Unix())
}

// DecodeUnixNano decodes unix-nano from json.
func DecodeUnixNano(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, val), nil
}

// EncodeUnixNano encodes unix-nano to json.
func EncodeUnixNano(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.UnixNano())
}

// DecodeUnixMicro decodes unix-micro from json.
func DecodeUnixMicro(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(val), nil
}

// EncodeUnixMicro encodes unix-micro to json.
func EncodeUnixMicro(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.UnixMicro())
}

// DecodeUnixMilli decodes unix-milli from json.
func DecodeUnixMilli(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(val), nil
}

// EncodeUnixMilli encodes unix-milli to json.
func EncodeUnixMilli(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.UnixMilli())
}
