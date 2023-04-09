package uri

import (
	"net/http"
	"strings"

	"github.com/go-faster/errors"
)

type cookieParamDecoder struct {
	paramName string
	explode   bool
	req       *http.Request
}

func (d *cookieParamDecoder) DecodeValue() (string, error) {
	c, err := d.req.Cookie(d.paramName)
	switch {
	case err == nil: // if NO error
	case errors.Is(err, http.ErrNoCookie):
		return "", errors.Errorf("cookie parameter %q not set", d.paramName)
	default:
		return "", errors.Wrapf(err, "get cookie %q", d.paramName)
	}

	val, ok := unescapeCookie(c.Value)
	if !ok {
		return "", errors.Errorf("invalid cookie escaping %q", c.Value)
	}
	return val, nil
}

func (d *cookieParamDecoder) DecodeArray(f func(Decoder) error) error {
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

func (d *cookieParamDecoder) DecodeFields(f func(field string, d Decoder) error) error {
	val, err := d.DecodeValue()
	if err != nil {
		return err
	}

	pushField := func(field, value string) error { return f(field, constval{value}) }
	const kvSep, fieldSep = ',', ','
	return decodeObject(
		(&cursor{src: val}),
		kvSep, fieldSep, pushField,
	)
}
