package uri

// OpenAPI v3 allows to use only primitive arrays
// and objects with primitive fields as parameter types.
//
// Сurrent Encoder/Decoder interface design allows
// to represent type with any nesting level.
// This was done for simplicity of templates.
//
// The actual encoders/decoders (PathEncoder, PathDecoder, QueryEncoder, etc)
// does not support nested types and panic if you try to encode/decode them.
//
// To prevent these panics, gen checks that parameter type is satisfying
// for OpenAPI constraints (internal/gen/gen_parameters.go:isParamAllowed).
// This should protect from calling interface methods which can panic at runtime.
// But it still looks pretty dangerous, probably should be rewritten later.

type Encoder interface {
	EncodeValue(v string) error
	EncodeArray(f func(e Encoder) error) error
	EncodeField(name string, f func(e Encoder) error) error
}

type Decoder interface {
	DecodeValue() (string, error)
	DecodeArray(f func(d Decoder) error) error
	DecodeFields(f func(field string, d Decoder) error) error
}
