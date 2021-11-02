package uri

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
