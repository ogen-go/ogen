package uri

import (
	"net/url"

	"github.com/go-faster/errors"
)

type QueryDecoder struct {
	values url.Values
}

func NewQueryDecoder(values url.Values) *QueryDecoder {
	return &QueryDecoder{
		values: values,
	}
}

type QueryParameterDecodingConfig struct {
	Name    string
	Style   QueryStyle
	Explode bool
	Fields  []QueryParameterObjectField // Only for object param.
}

type QueryParameterObjectField struct {
	Name     string
	Required bool
}

func (d *QueryDecoder) HasParam(cfg QueryParameterDecodingConfig) error {
	if len(cfg.Fields) > 0 {
		// https://swagger.io/docs/specification/serialization/
		var (
			case1 = cfg.Style == QueryStyleForm && cfg.Explode
			case2 = cfg.Style == QueryStyleDeepObject && cfg.Explode
		)

		if case1 || case2 {
			found := false
			for _, field := range cfg.Fields {
				qparam := field.Name
				if case2 {
					qparam = cfg.Name + "[" + field.Name + "]"
				}

				if _, ok := d.values[qparam]; ok {
					found = true
					continue
				}

				if field.Required {
					return errors.Errorf("query parameter %q not set", qparam)
				}
			}

			if !found {
				return errors.Errorf("none of parameters (%+v) are set", cfg.Fields)
			}

			return nil
		}
	}

	if _, ok := d.values[cfg.Name]; !ok {
		return errors.Errorf("query parameter %q not set", cfg.Name)
	}
	return nil
}

func (d *QueryDecoder) DecodeParam(cfg QueryParameterDecodingConfig, f func(Decoder) error) error {
	p := &queryParamDecoder{
		values:       d.values,
		objectFields: cfg.Fields,

		paramName: cfg.Name,
		style:     cfg.Style,
		explode:   cfg.Explode,
	}

	return f(p)
}
