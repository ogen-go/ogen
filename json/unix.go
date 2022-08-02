package json

import (
	"time"

	"github.com/go-faster/jx"
)

// DecodeUnixSeconds decodes unix-seconds from json string.
func DecodeUnixSeconds(d *jx.Decoder) (time.Time, error) {
	val, err := d.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

// EncodeUnixSeconds encodes unix-seconds to json string.
func EncodeUnixSeconds(e *jx.Encoder, v time.Time) {
	e.Int64(v.Unix())
}

// DecodeUnixNano decodes unix-nano from json string.
func DecodeUnixNano(d *jx.Decoder) (time.Time, error) {
	val, err := d.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, val), nil
}

// EncodeUnixNano encodes unix-nano to json string.
func EncodeUnixNano(e *jx.Encoder, v time.Time) {
	e.Int64(v.UnixNano())
}

// DecodeUnixMicro decodes unix-micro from json string.
func DecodeUnixMicro(d *jx.Decoder) (time.Time, error) {
	val, err := d.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(val), nil
}

// EncodeUnixMicro encodes unix-micro to json string.
func EncodeUnixMicro(e *jx.Encoder, v time.Time) {
	e.Int64(v.UnixMicro())
}

// DecodeUnixMilli decodes unix-milli from json string.
func DecodeUnixMilli(d *jx.Decoder) (time.Time, error) {
	val, err := d.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(val), nil
}

// EncodeUnixMilli encodes unix-milli to json string.
func EncodeUnixMilli(e *jx.Encoder, v time.Time) {
	e.Int64(v.UnixMilli())
}
