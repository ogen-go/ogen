package json

import (
	"strconv"

	"github.com/go-faster/jx"
)

// DecodeStringInt32 decodes string int32 from json.
func DecodeStringInt32(d *jx.Decoder) (v int32, err error) {
	s, err := d.Str()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// EncodeStringInt32 encodes string int32 to json.
func EncodeStringInt32(e *jx.Encoder, v int32) {
	var (
		buf [32]byte
		n   int
	)
	// Write first quote
	buf[n] = '"'
	n++
	// Write integer
	n += len(strconv.AppendInt(buf[n:n], int64(v), 10))
	// Write second quote
	buf[n] = '"'
	n++
	e.Raw(buf[:n])
}

// DecodeStringInt64 decodes string int64 from json.
func DecodeStringInt64(d *jx.Decoder) (v int64, err error) {
	s, err := d.Str()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// EncodeStringInt64 encodes string int64 to json.
func EncodeStringInt64(e *jx.Encoder, v int64) {
	var (
		buf [64]byte
		n   int
	)
	// Write first quote
	buf[n] = '"'
	n++
	// Write integer
	n += len(strconv.AppendInt(buf[n:n], int64(v), 10))
	// Write second quote
	buf[n] = '"'
	n++
	e.Raw(buf[:n])
}
