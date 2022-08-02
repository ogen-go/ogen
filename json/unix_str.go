package json

import (
	"time"

	"github.com/go-faster/jx"
)

// DecodeStringUnixSeconds decodes unix-seconds from json string.
func DecodeStringUnixSeconds(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

// EncodeStringUnixSeconds encodes unix-seconds to json string.
func EncodeStringUnixSeconds(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.Unix())
}

// DecodeStringUnixNano decodes unix-nano from json string.
func DecodeStringUnixNano(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, val), nil
}

// EncodeStringUnixNano encodes unix-nano to json string.
func EncodeStringUnixNano(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.UnixNano())
}

// DecodeStringUnixMicro decodes unix-micro from json string.
func DecodeStringUnixMicro(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(val), nil
}

// EncodeStringUnixMicro encodes unix-micro to json string.
func EncodeStringUnixMicro(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.UnixMicro())
}

// DecodeStringUnixMilli decodes unix-milli from json string.
func DecodeStringUnixMilli(d *jx.Decoder) (time.Time, error) {
	val, err := DecodeStringInt64(d)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(val), nil
}

// EncodeStringUnixMilli encodes unix-milli to json string.
func EncodeStringUnixMilli(e *jx.Encoder, v time.Time) {
	EncodeStringInt64(e, v.UnixMilli())
}
