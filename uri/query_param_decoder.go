package uri

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"
)

type queryParamDecoder struct {
	values       url.Values
	objectFields []QueryParameterObjectField

	paramName string
	style     QueryStyle // immutable
	explode   bool       // immutable
}

func (d *queryParamDecoder) DecodeValue() (string, error) {
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

func (d *queryParamDecoder) DecodeArray(f func(d Decoder) error) error {
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

		// do not decode `?param=` as `[""]` and leave the parameter as whatever zero value it has
		if values[0] == "" {
			return nil
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

func (d *queryParamDecoder) DecodeFields(f func(name string, d Decoder) error) error {
	adapter := func(name, value string) error { return f(name, &constval{value}) }
	switch d.style {
	case QueryStyleForm:
		if d.explode {
			for _, field := range d.objectFields {
				values, ok := d.values[field.Name]
				if !ok || len(values) == 0 {
					if !field.Required {
						continue
					}
					return errors.Errorf("query parameter %q field %q not set", d.paramName, field.Name)
				}

				if len(values) > 1 {
					return errors.Errorf("query parameter %q field %q multiple values", d.paramName, field.Name)
				}

				if err := adapter(field.Name, values[0]); err != nil {
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

		cur := &cursor{src: values[0]}
		return decodeObject(cur, ',', ',', adapter)

	case QueryStyleSpaceDelimited:
		panic("object cannot have spaceDelimited style")

	case QueryStylePipeDelimited:
		panic("object cannot have pipeDelimited style")

	case QueryStyleDeepObject:
		if !d.explode {
			panic("invalid deepObject style configuration")
		}

		for _, field := range d.objectFields {
			qparam := d.paramName + "[" + field.Name + "]"
			values, ok := d.values[qparam]
			if !ok || len(values) == 0 {
				if !field.Required {
					continue
				}
				return errors.Errorf("query parameter %q field %q not set", d.paramName, qparam)
			}

			if len(values) > 1 {
				return errors.Errorf("query parameter %q field %q multiple values", d.paramName, qparam)
			}

			if err := adapter(field.Name, values[0]); err != nil {
				return err
			}
		}
		return nil

	default:
		panic("unreachable")
	}
}
