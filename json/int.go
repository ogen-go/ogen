package json

import (
	"strconv"

	"github.com/go-faster/jx"
	"golang.org/x/exp/constraints"
)

func encodeStringInt[T constraints.Integer](e *jx.Encoder, v T) {
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

// EncodeStringInt encodes string int to json.
func EncodeStringInt(e *jx.Encoder, v int) {
	encodeStringInt(e, v)
}

// DecodeStringInt decodes string int from json.
func DecodeStringInt(d *jx.Decoder) (int, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).Int()
}

// EncodeStringInt8 encodes string int8 to json.
func EncodeStringInt8(e *jx.Encoder, v int8) {
	encodeStringInt(e, v)
}

// DecodeStringInt8 decodes string int8 from json.
func DecodeStringInt8(d *jx.Decoder) (int8, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).Int8()
}

// EncodeStringInt16 encodes string int16 to json.
func EncodeStringInt16(e *jx.Encoder, v int16) {
	encodeStringInt(e, v)
}

// DecodeStringInt16 decodes string int16 from json.
func DecodeStringInt16(d *jx.Decoder) (int16, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).Int16()
}

// EncodeStringInt32 encodes string int32 to json.
func EncodeStringInt32(e *jx.Encoder, v int32) {
	encodeStringInt(e, v)
}

// DecodeStringInt32 decodes string int32 from json.
func DecodeStringInt32(d *jx.Decoder) (int32, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).Int32()
}

// EncodeStringInt64 encodes string int64 to json.
func EncodeStringInt64(e *jx.Encoder, v int64) {
	encodeStringInt(e, v)
}

// DecodeStringInt64 decodes string int64 from json.
func DecodeStringInt64(d *jx.Decoder) (int64, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).Int64()
}
