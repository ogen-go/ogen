package uri

import (
	"net/http"

	"github.com/go-faster/errors"
)

type cookieParamDecoder struct {
	paramName string
	explode   bool
	req       *http.Request
}

func (d *cookieParamDecoder) DecodeValue() (string, error) {
	c, err := d.req.Cookie(d.paramName)
	if err != nil {
		return "", errors.Wrapf(err, "get cookie %q", d.paramName)
	}

	val, ok := unescapeCookie(c.Value)
	if !ok {
		return "", errors.Errorf("invalid cookie escaping %q", c.Value)
	}
	return val, nil
}

func (d *cookieParamDecoder) DecodeArray(f func(Decoder) error) error {
	panic("cookie with array values is not implemented")
}

func (d *cookieParamDecoder) DecodeFields(f func(field string, d Decoder) error) error {
	panic("cookie with object values is not implemented")
}
