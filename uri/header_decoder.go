package uri

import "strings"

type HeaderDecoder struct {
	value   string
	explode bool
}

type HeaderDecoderConfig struct {
	Value   string
	Explode bool
}

func NewHeaderDecoder(cfg HeaderDecoderConfig) *HeaderDecoder {
	return &HeaderDecoder{
		value:   cfg.Value,
		explode: cfg.Explode,
	}
}

func (d *HeaderDecoder) Value() (string, error) {
	return d.value, nil
}

func (d *HeaderDecoder) Array(f func(Decoder) error) error {
	for _, v := range strings.Split(d.value, ",") {
		if err := f(constval{v}); err != nil {
			return err
		}
	}
	return nil
}

func (d *HeaderDecoder) Fields(f func(field string, d Decoder) error) error {
	adapter := func(field, value string) error {
		return f(field, constval{value})
	}

	kvSep, fieldSep := ',', ','
	if d.explode {
		kvSep = '='
	}
	return decodeObject(
		(&cursor{src: []rune(d.value)}),
		kvSep, fieldSep, adapter,
	)
}
