package uri

import "net/http"

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
	p := &cookieParamEncoder{
		receiver:  newReceiver(),
		paramName: cfg.Name,
		explode:   cfg.Explode,
		req:       e.req,
	}

	if err := f(p); err != nil {
		return err
	}

	p.serialize()
	return nil
}
