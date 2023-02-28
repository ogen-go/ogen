package uri

import "net/http"

type HeaderEncoder struct {
	header http.Header
}

func NewHeaderEncoder(header http.Header) *HeaderEncoder {
	return &HeaderEncoder{
		header: header,
	}
}

type HeaderParameterEncodingConfig struct {
	Name    string
	Explode bool
}

func (e *HeaderEncoder) EncodeParam(cfg HeaderParameterEncodingConfig, f func(Encoder) error) error {
	p := &headerParamEncoder{
		receiver:  newReceiver(),
		paramName: cfg.Name,
		explode:   cfg.Explode,
		header:    e.header,
	}

	if err := f(p); err != nil {
		return err
	}

	// FIXME(tdakkota): probable we should return the error during encoding
	return p.serialize()
}

func (e *HeaderEncoder) Header() http.Header {
	return e.header
}
