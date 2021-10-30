package uri

import (
	"fmt"
	"strings"
)

type QueryStyle string

const (
	QueryStyleForm           QueryStyle = "form"
	QueryStyleSpaceDelimited QueryStyle = "spaceDelimited"
	QueryStylePipeDelimited  QueryStyle = "pipeDelimited"
	QueryStyleDeepObject     QueryStyle = "deepObject"
)

type QueryDecoder struct {
	param string
	src   []string // r.URL.Query()["param"]

	style   QueryStyle // immutable
	explode bool       // immutable
}

type QueryDecoderConfig struct {
	Param   string
	Values  []string
	Style   QueryStyle
	Explode bool
}

func NewQueryDecoder(cfg QueryDecoderConfig) *QueryDecoder {
	if len(cfg.Values) == 0 {
		panic("unreachable")
	}

	return &QueryDecoder{
		param:   cfg.Param,
		src:     cfg.Values,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (d *QueryDecoder) DecodeValue() (string, error) {
	switch d.style {
	case QueryStyleForm:
		if len(d.src) != 1 {
			return "", fmt.Errorf("multiple params")
		}
		return d.src[0], nil
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		return "", fmt.Errorf("style '%s' cannot be used for primitive values", d.style)
	default:
		panic("unreachable")
	}
}

func (d *QueryDecoder) DecodeArray() ([]string, error) {
	if len(d.src) < 1 {
		return nil, fmt.Errorf("empty array")
	}

	switch d.style {
	case QueryStyleForm:
		if d.explode {
			return d.src, nil
		}

		if len(d.src) != 1 {
			return nil, fmt.Errorf("invalid value")
		}

		return strings.Split(d.src[0], ","), nil

	case QueryStyleSpaceDelimited:
		if d.explode {
			return d.src, nil
		}

		if len(d.src) != 1 {
			return nil, fmt.Errorf("invalid value")
		}

		return nil, fmt.Errorf("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if d.explode {
			return d.src, nil
		}

		if len(d.src) != 1 {
			return nil, fmt.Errorf("invalid value")
		}

		return strings.Split(d.src[0], "|"), nil

	case QueryStyleDeepObject:
		return nil, fmt.Errorf("style '%s' cannot be used for arrays", d.style)

	default:
		panic("unreachable")
	}
}

func (d *QueryDecoder) DecodeObject(f func(field, value string) error) error {
	switch d.style {
	case QueryStyleForm:
		if d.explode {
			for _, v := range d.src {
				if strings.Count(v, "=") != 1 {
					return fmt.Errorf("invalid value: %s", v)
				}

				s := strings.Split(v, "=")
				if err := f(s[0], s[1]); err != nil {
					return err
				}
			}
			return nil
		}

		if len(d.src) > 1 {
			return fmt.Errorf("multiple values passed")
		}

		cur := &cursor{src: []rune(d.src[0])}
		param, err := cur.readAt('=')
		if err != nil {
			return err
		}

		if param != d.param {
			return fmt.Errorf("invalid param name: '%s'", param)
		}

		return decodeObject(cur, ',', ',', f)

	case QueryStyleSpaceDelimited:
		panic("object cannot have spaceDelimited style")

	case QueryStylePipeDelimited:
		panic("object cannot have pipeDelimited style")

	case QueryStyleDeepObject:
		if !d.explode {
			panic("invalid deepObject style configuration")
		}

		cur := &cursor{}
		for _, v := range d.src {
			cur.src = []rune(v)
			cur.pos = 0

			pname, err := cur.readAt('[')
			if err != nil {
				return err
			}

			if pname != d.param {
				return fmt.Errorf("invalid param name: '%s'", pname)
			}

			key, err := cur.readAt(']')
			if err != nil {
				return err
			}

			if !cur.eat('=') {
				return fmt.Errorf("invalid value")
			}

			val, err := cur.readAll()
			if err != nil {
				return err
			}

			if err := f(key, val); err != nil {
				return err
			}
		}
		return nil

	default:
		panic("unreachable")
	}
}
