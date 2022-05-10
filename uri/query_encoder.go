package uri

import "net/url"

type QueryEncoder struct {
	values url.Values
}

func NewQueryEncoder() *QueryEncoder {
	return &QueryEncoder{
		values: make(url.Values),
	}
}

type QueryParameterEncodingConfig struct {
	Name    string
	Style   QueryStyle
	Explode bool
}

func (e *QueryEncoder) EncodeParam(cfg QueryParameterEncodingConfig, f func(Encoder) error) error {
	p := &queryParamEncoder{
		receiver: newReceiver(),
		values:   e.values,

		paramName: cfg.Name,
		style:     cfg.Style,
		explode:   cfg.Explode,
	}

	if err := f(p); err != nil {
		return err
	}

	p.serialize()
	return nil
}

func (e *QueryEncoder) Values() url.Values {
	return e.values
}
