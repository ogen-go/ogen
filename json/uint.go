package json

import (
	"strconv"

	"github.com/go-faster/jx"
	"golang.org/x/exp/constraints"
)

func encodeStringUint[T constraints.Unsigned](e *jx.Encoder, v T) {
	var (
		buf [32]byte
		n   int
	)
	// Write first quote
	buf[n] = '"'
	n++
	// Write integer
	n += len(strconv.AppendUint(buf[n:n], uint64(v), 10))
	// Write second quote
	buf[n] = '"'
	n++
	e.Raw(buf[:n])
}

// EncodeStringUint encodes string uint to json.
func EncodeStringUint(e *jx.Encoder, v uint) {
	encodeStringUint(e, v)
}

// DecodeStringUint decodes string int from json.
func DecodeStringUint(d *jx.Decoder) (uint, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).UInt()
}

// EncodeStringUint8 encodes string uint8 to json.
func EncodeStringUint8(e *jx.Encoder, v uint8) {
	encodeStringUint(e, v)
}

// DecodeStringUint8 decodes string int8 from json.
func DecodeStringUint8(d *jx.Decoder) (uint8, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).UInt8()
}

// EncodeStringUint16 encodes string uint16 to json.
func EncodeStringUint16(e *jx.Encoder, v uint16) {
	encodeStringUint(e, v)
}

// DecodeStringUint16 decodes string int16 from json.
func DecodeStringUint16(d *jx.Decoder) (uint16, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).UInt16()
}

// EncodeStringUint32 encodes string uint32 to json.
func EncodeStringUint32(e *jx.Encoder, v uint32) {
	encodeStringUint(e, v)
}

// DecodeStringUint32 decodes string int32 from json.
func DecodeStringUint32(d *jx.Decoder) (uint32, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).UInt32()
}

// EncodeStringUint64 encodes string uint64 to json.
func EncodeStringUint64(e *jx.Encoder, v uint64) {
	encodeStringUint(e, v)
}

// DecodeStringUint64 decodes string int64 from json.
func DecodeStringUint64(d *jx.Decoder) (uint64, error) {
	s, err := d.StrBytes()
	if err != nil {
		return 0, err
	}
	return jx.DecodeBytes(s).UInt64()
}
