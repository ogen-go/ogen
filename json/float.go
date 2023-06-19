package json

import (
	"strconv"

	"github.com/go-faster/jx"
	"golang.org/x/exp/constraints"
)

func encodeStringFloat[T constraints.Float](e *jx.Encoder, v T, bitSize int) {
	var (
		buf [32]byte
		n   int
	)
	// Write first quote
	buf[n] = '"'
	n++
	// Write float
	n += len(strconv.AppendFloat(buf[n:n], float64(v), 'g', -1, bitSize))
	// Write second quote
	buf[n] = '"'
	n++
	e.Raw(buf[:n])
}

// EncodeStringFloat32 encodes string float32 to json.
func EncodeStringFloat32(e *jx.Encoder, v float32) {
	encodeStringFloat(e, v, 32)
}

// DecodeStringFloat32 decodes string float32 from json.
func DecodeStringFloat32(d *jx.Decoder) (float32, error) {
	s, err := d.Str()
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

// EncodeStringFloat64 encodes string float64 to json.
func EncodeStringFloat64(e *jx.Encoder, v float64) {
	encodeStringFloat(e, v, 64)
}

// DecodeStringFloat64 decodes string float64 from json.
func DecodeStringFloat64(d *jx.Decoder) (float64, error) {
	s, err := d.Str()
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err
}
