package uri

import (
	"net/http"

	"github.com/go-faster/errors"
)

type CookieEncoder struct {
	req *http.Request
}

func NewCookieEncoder(req *http.Request) *CookieEncoder {
	return &CookieEncoder{
		req: req,
	}
}

type CookieParameterEncodingConfig struct {
	Name    string
	Explode bool
}

func (e *CookieEncoder) EncodeParam(cfg CookieParameterEncodingConfig, f func(Encoder) error) error {
	if !isCookieNameValid(cfg.Name) {
		return errors.Errorf("invalid cookie name %q", cfg.Name)
	}
	p := &cookieParamEncoder{
		receiver:  newReceiver(),
		paramName: cfg.Name,
		explode:   cfg.Explode,
		req:       e.req,
	}

	if err := f(p); err != nil {
		return err
	}

	return p.serialize()
}
