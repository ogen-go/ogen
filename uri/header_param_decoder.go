package uri

import (
	"net/http"
	"strings"

	"github.com/go-faster/errors"
)

type headerParamDecoder struct {
	paramName string
	explode   bool
	header    http.Header
}

func (d *headerParamDecoder) DecodeValue() (string, error) {
	if len(d.header.Values(d.paramName)) == 0 {
		return "", errors.Errorf("header parameter %q not set", d.paramName)
	}

	return d.header.Get(d.paramName), nil
}

func (d *headerParamDecoder) DecodeArray(f func(Decoder) error) error {
	val, err := d.DecodeValue()
	if err != nil {
		return err
	}

	for _, v := range strings.Split(val, ",") {
		if err := f(constval{v}); err != nil {
			return err
		}
	}
	return nil
}

func (d *headerParamDecoder) DecodeFields(f func(field string, d Decoder) error) error {
	val, err := d.DecodeValue()
	if err != nil {
		return err
	}

	pushField := func(field, value string) error { return f(field, constval{value}) }
	var kvSep, fieldSep byte = ',', ','
	if d.explode {
		kvSep = '='
	}
	return decodeObject(
		(&cursor{src: val}),
		kvSep, fieldSep, pushField,
	)
}
