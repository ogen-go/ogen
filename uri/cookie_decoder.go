package uri

import (
	"net/http"

	"github.com/go-faster/errors"
)

type CookieDecoder struct {
	req *http.Request
}

func NewCookieDecoder(req *http.Request) *CookieDecoder {
	return &CookieDecoder{
		req: req,
	}
}

type CookieParameterDecodingConfig struct {
	Name    string
	Explode bool
}

func (d *CookieDecoder) HasParam(cfg CookieParameterDecodingConfig) error {
	_, err := d.req.Cookie(cfg.Name)
	if errors.Is(err, http.ErrNoCookie) {
		return errors.Errorf("cookie parameter %q not set", cfg.Name)
	}
	return err
}

func (d *CookieDecoder) DecodeParam(cfg CookieParameterDecodingConfig, f func(Decoder) error) error {
	p := &cookieParamDecoder{
		paramName: cfg.Name,
		explode:   cfg.Explode,
		req:       d.req,
	}

	return f(p)
}
