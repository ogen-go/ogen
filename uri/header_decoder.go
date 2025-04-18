package uri

import (
	"net/http"

	"github.com/ogen-go/ogen/validate"
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
		return &validate.Error{
			Fields: []validate.FieldError{
				{
					Name:  cfg.Name,
					Error: validate.ErrFieldRequired,
				},
			},
		}
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
