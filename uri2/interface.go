package uri

type Encoder interface {
	Value(v string) error
	Array(f func(e Encoder) error) error
	Field(name string, f func(e Encoder) error) error
}

type Decoder interface {
	Value() (string, error)
	Array(f func(d Decoder) error) error
	Fields(f func(field string, d Decoder) error) error
}
