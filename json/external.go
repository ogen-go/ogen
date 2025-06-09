package json

import (
	"encoding"
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

// EncodeStringNative encodes a string value using [Marshaler] interface.
func EncodeStringNative[T any, P marshaler[T]](e *jx.Encoder, v T) {
	EncodeNative[T, P](e, v)
}

// DecodeStringNative decodes a string value using [Unmarshaler] interface.
func DecodeStringNative[T any, P unmarshaler[T]](d *jx.Decoder) (T, error) {
	return DecodeNative[T, P](d)
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

// EncodeStringJSON encodes a string value using [json.Marshaler] interface.
func EncodeStringJSON[T any, P jsonMarshaler[T]](e *jx.Encoder, v T) {
	EncodeJSON[T, P](e, v)
}

// DecodeStringJSON decodes a string value using [json.Marshaler] interface.
func DecodeStringJSON[T any, P jsonUnmarshaler[T]](d *jx.Decoder) (T, error) {
	return DecodeJSON[T, P](d)
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

// EncodeStringExternal encodes a string value using [json.Marshal].
func EncodeStringExternal[T any](e *jx.Encoder, v T) {
	EncodeExternal(e, v)
}

// DecodeStringExternal decodes a string value using [json.Unmarshal].
func DecodeStringExternal[T any](d *jx.Decoder) (T, error) {
	return DecodeExternal[T](d)
}
