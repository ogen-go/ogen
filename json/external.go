package json

import (
	"encoding"
	"encoding/base64"
	"encoding/json"

	"github.com/go-faster/jx"
)

type (
	marshaler[T any] interface {
		Marshaler
		*T
	}
	unmarshaler[T any] interface {
		Unmarshaler
		*T
	}
	textMarshaler[T any] interface {
		encoding.TextMarshaler
		*T
	}
	textUnmarshaler[T any] interface {
		encoding.TextUnmarshaler
		*T
	}
	binaryMarshaler[T any] interface {
		encoding.BinaryMarshaler
		*T
	}
	binaryUnmarshaler[T any] interface {
		encoding.BinaryUnmarshaler
		*T
	}
	jsonMarshaler[T any] interface {
		json.Marshaler
		*T
	}
	jsonUnmarshaler[T any] interface {
		json.Unmarshaler
		*T
	}
)

// EncodeNative encodes a value using [Marshaler] interface.
func EncodeNative[T any, P marshaler[T]](e *jx.Encoder, v T) {
	P(&v).Encode(e)
}

// DecodeNative decodes a value using [Unmarshaler] interface.
func DecodeNative[T any, P unmarshaler[T]](d *jx.Decoder) (T, error) {
	var v T
	err := P(&v).Decode(d)
	return v, err
}

// EncodeText encodes a value using [encoding.TextMarshaler] interface.
func EncodeText[T any, P textMarshaler[T]](e *jx.Encoder, v T) {
	b, _ := P(&v).MarshalText()
	e.Raw(b)
}

// DecodeText decodes a value using [encoding.TextUnmarshaler] interface.
func DecodeText[T any, P textUnmarshaler[T]](d *jx.Decoder) (T, error) {
	var v T
	b, err := d.Raw()
	if err != nil {
		return v, err
	}
	err = P(&v).UnmarshalText(b)
	return v, err
}

// EncodeStringText encodes a string value using [encoding.TextMarshaler] interface.
func EncodeStringText[T any, P textMarshaler[T]](e *jx.Encoder, v T) {
	b, _ := P(&v).MarshalText()
	e.ByteStr(b)
}

// DecodeStringText decodes a string value using [encoding.TextUnmarshaler] interface.
func DecodeStringText[T any, P textUnmarshaler[T]](d *jx.Decoder) (T, error) {
	var v T
	b, err := d.StrBytes()
	if err != nil {
		return v, err
	}
	err = P(&v).UnmarshalText(b)
	return v, err
}

// EncodeBinary encodes a value using [encoding.BinaryMarshaler] interface.
func EncodeBinary[T any, P binaryMarshaler[T]](e *jx.Encoder, v T) {
	raw, _ := P(&v).MarshalBinary()
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(encoded, raw)
	e.ByteStr(encoded)
}

// DecodeBinary decodes a value using [encoding.BinaryUnmarshaler] interface.
func DecodeBinary[T any, P binaryUnmarshaler[T]](d *jx.Decoder) (T, error) {
	var v T
	raw, err := d.StrBytes()
	if err != nil {
		return v, err
	}
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	n, err := base64.StdEncoding.Decode(decoded, raw)
	if err != nil {
		return v, err
	}
	err = P(&v).UnmarshalBinary(decoded[:n])
	return v, err
}

// EncodeJSON encodes a value using [json.Marshaler] interface.
func EncodeJSON[T any, P jsonMarshaler[T]](e *jx.Encoder, v T) {
	b, _ := P(&v).MarshalJSON()
	e.Raw(b)
}

// DecodeJSON decodes a value using [json.Marshaler] interface.
func DecodeJSON[T any, P jsonUnmarshaler[T]](d *jx.Decoder) (T, error) {
	var v T
	b, err := d.Raw()
	if err != nil {
		return v, err
	}
	err = P(&v).UnmarshalJSON(b)
	return v, err
}

// EncodeExternal encodes a value using [json.Marshal].
func EncodeExternal[T any](e *jx.Encoder, v T) {
	b, _ := json.Marshal(v)
	e.Raw(b)
}

// DecodeExternal decodes a value using [json.Unmarshal].
func DecodeExternal[T any](d *jx.Decoder) (T, error) {
	var v T
	b, err := d.Raw()
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(b, &v)
	return v, err
}
