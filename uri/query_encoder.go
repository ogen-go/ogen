package uri

import (
	"mime/multipart"
	"net/url"

	"github.com/go-faster/errors"
)

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

	// FIXME(tdakkota): probable we should return the error during encoding
	return p.serialize()
}

func (e *QueryEncoder) Values() url.Values {
	return e.values
}

func (e *QueryEncoder) WriteMultipart(w *multipart.Writer) error {
	for k, values := range e.values {
		for i, value := range values {
			if err := w.WriteField(k, value); err != nil {
				return errors.Wrapf(err, "write %q: [%d]", k, i)
			}
		}
	}
	return nil
}
