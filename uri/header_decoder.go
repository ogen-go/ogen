package uri

import (
	"net/http"

	"github.com/go-faster/errors"
)

type HeaderDecoder struct {
	header http.Header
}

func NewHeaderDecoder(header http.Header) *HeaderDecoder {
	return &HeaderDecoder{
		header: header,
	}
}

type HeaderParameterDecodingConfig struct {
	Name    string
	Explode bool
}

func (d *HeaderDecoder) HasParam(cfg HeaderParameterDecodingConfig) error {
	if len(d.header.Values(cfg.Name)) == 0 {
		return errors.Errorf("header parameter %q not set", cfg.Name)
	}
	return nil
}

func (d *HeaderDecoder) DecodeParam(cfg HeaderParameterDecodingConfig, f func(Decoder) error) error {
	p := &headerParamDecoder{
		paramName: cfg.Name,
		explode:   cfg.Explode,
		header:    d.header,
	}

	return f(p)
}
