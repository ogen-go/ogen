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

func (d HeaderDecoder) DecodeString() (string, error) {
	return d.value, nil
}

func (d HeaderDecoder) DecodeStrings() ([]string, error) {
	return strings.Split(d.value, ","), nil
}

func (d HeaderDecoder) DecodeObject(f func(field, value string) error) error {
	kvSep, fieldSep := ',', ','
	if d.explode {
		kvSep = '='
	}
	return decodeObject(
		(&cursor{src: []rune(d.value)}),
		kvSep, fieldSep, f,
	)
}
