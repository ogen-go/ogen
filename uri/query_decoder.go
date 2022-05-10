package uri

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"
)

type QueryStyle string

const (
	QueryStyleForm           QueryStyle = "form"
	QueryStyleSpaceDelimited QueryStyle = "spaceDelimited"
	QueryStylePipeDelimited  QueryStyle = "pipeDelimited"
	QueryStyleDeepObject     QueryStyle = "deepObject"
)

type QueryDecoder struct {
	paramName    string
	style        QueryStyle // immutable
	explode      bool       // immutable
	values       url.Values
	objectFields []string
}

type QueryDecoderConfig struct {
	Param        string
	Values       url.Values
	Style        QueryStyle
	Explode      bool
	ObjectFields []string // only for object param
}

func NewQueryDecoder(cfg QueryDecoderConfig) *QueryDecoder {
	if len(cfg.Values) == 0 {
		panic("unreachable")
	}

	return &QueryDecoder{
		paramName:    cfg.Param,
		values:       cfg.Values,
		style:        cfg.Style,
		explode:      cfg.Explode,
		objectFields: cfg.ObjectFields,
	}
}

func (d *QueryDecoder) DecodeValue() (string, error) {
	switch d.style {
	case QueryStyleForm:
		values, ok := d.values[d.paramName]
		if !ok {
			return "", errors.Errorf("query parameter %q not set", d.paramName)
		}

		if len(values) != 1 {
			return "", errors.Errorf("query parameter %q multiple values", d.paramName)
		}

		return values[0], nil
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		return "", errors.Errorf("style %q cannot be used for primitive values", d.style)
	default:
		panic("unreachable")
	}
}

func (d *QueryDecoder) DecodeArray(f func(d Decoder) error) error {
	values, ok := d.values[d.paramName]
	if !ok {
		return errors.Errorf("query parameter %q not set", d.paramName)
	}

	switch d.style {
	case QueryStyleForm:
		if d.explode {
			for _, item := range values {
				if err := f(&constval{item}); err != nil {
					return err
				}
			}
			return nil
		}

		if len(values) != 1 {
			return errors.New("invalid value")
		}

		for _, item := range strings.Split(values[0], ",") {
			if err := f(&constval{item}); err != nil {
				return err
			}
		}

		return nil

	case QueryStyleSpaceDelimited:
		if d.explode {
			for _, item := range values {
				if err := f(&constval{item}); err != nil {
					return err
				}
			}
			return nil
		}

		if len(values) != 1 {
			return errors.New("invalid value")
		}

		return errors.New("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if d.explode {
			for _, item := range values {
				if err := f(&constval{item}); err != nil {
					return err
				}
			}
			return nil
		}

		if len(values) != 1 {
			return errors.New("invalid value")
		}

		for _, item := range strings.Split(values[0], "|") {
			if err := f(&constval{item}); err != nil {
				return err
			}
		}

		return nil

	case QueryStyleDeepObject:
		return errors.Errorf("style %q cannot be used for arrays", d.style)

	default:
		panic("unreachable")
	}
}

func (d *QueryDecoder) DecodeFields(f func(name string, d Decoder) error) error {
	adapter := func(name, value string) error { return f(name, &constval{value}) }
	switch d.style {
	case QueryStyleForm:
		if d.explode {
			for _, fname := range d.objectFields {
				values, ok := d.values[fname]
				if !ok || len(values) == 0 {
					return errors.Errorf("query parameter %q field %q not set", d.paramName, fname)
				}

				if len(values) > 1 {
					return errors.Errorf("query parameter %q field %q multiple values", d.paramName, fname)
				}

				if err := adapter(fname, values[0]); err != nil {
					return err
				}
			}

			return nil
		}

		values, ok := d.values[d.paramName]
		if !ok {
			return errors.Errorf("query parameter %q not set", d.paramName)
		}

		if len(values) > 1 {
			return errors.Errorf("query parameter %q multiple values", d.paramName)
		}

		cur := &cursor{src: []rune(values[0])}
		return decodeObject(cur, ',', ',', adapter)

	case QueryStyleSpaceDelimited:
		panic("object cannot have spaceDelimited style")

	case QueryStylePipeDelimited:
		panic("object cannot have pipeDelimited style")

	case QueryStyleDeepObject:
		if !d.explode {
			panic("invalid deepObject style configuration")
		}

		for _, fname := range d.objectFields {
			qparam := d.paramName + "[" + fname + "]"
			values, ok := d.values[qparam]
			if !ok || len(values) == 0 {
				return errors.Errorf("query parameter %q field %q not set", d.paramName, qparam)
			}

			if len(values) > 1 {
				return errors.Errorf("query parameter %q field %q multiple values", d.paramName, qparam)
			}

			if err := adapter(fname, values[0]); err != nil {
				return err
			}
		}
		return nil

	default:
		panic("unreachable")
	}
}
