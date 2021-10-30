package uri

import "fmt"

type PathStyle string

const (
	PathStyleSimple PathStyle = "simple"
	PathStyleLabel  PathStyle = "label"
	PathStyleMatrix PathStyle = "matrix"
)

func (s PathStyle) String() string { return string(s) }

type PathDecoder struct {
	cur *cursor

	param   string    // immutable
	style   PathStyle // immutable
	explode bool      // immutable
}

type PathDecoderConfig struct {
	Param   string // Parameter name
	Value   string // chi.URLParam(r, "paramName")
	Style   PathStyle
	Explode bool
}

func NewPathDecoder(cfg PathDecoderConfig) PathDecoder {
	return PathDecoder{
		cur:     &cursor{src: []rune(cfg.Value)},
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (d PathDecoder) DecodeString() (string, error) {
	switch d.style {
	case PathStyleSimple:
		return d.cur.readAll()

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return "", fmt.Errorf("label style value must begin with '.'")
		}
		return d.cur.readAll()

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return "", fmt.Errorf("label style value must begin with ';'")
		}

		param, err := d.cur.readAt('=')
		if err != nil {
			return "", err
		}

		if param != d.param {
			return "", fmt.Errorf("invalid param name '%s'", param)
		}

		return d.cur.readAll()

	default:
		panic("unreachable")
	}
}

func (d PathDecoder) DecodeStrings() ([]string, error) {
	var values []string
	switch d.style {
	case PathStyleSimple:
		for {
			v, hasNext, err := d.cur.readValue(',')
			if err != nil {
				return nil, err
			}

			values = append(values, v)
			if !hasNext {
				return values, nil
			}
		}

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return nil, fmt.Errorf("value must begin with '.'")
		}

		delim := ','
		if d.explode {
			delim = '.'
		}

		for {
			v, hasNext, err := d.cur.readValue(delim)
			if err != nil {
				return nil, err
			}

			values = append(values, v)
			if !hasNext {
				return values, nil
			}
		}

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return nil, fmt.Errorf("value must begin with ';'")
		}

		if !d.explode {
			param, err := d.cur.readAt('=')
			if err != nil {
				return nil, err
			}

			if param != d.param {
				return nil, fmt.Errorf("invalid param name: '%s'", param)
			}

			for {
				v, hasNext, err := d.cur.readValue(',')
				if err != nil {
					return nil, err
				}

				values = append(values, v)
				if !hasNext {
					return values, nil
				}
			}
		}

		for {
			param, err := d.cur.readAt('=')
			if err != nil {
				return nil, err
			}

			if param != d.param {
				return nil, fmt.Errorf("invalid param name '%s'", param)
			}

			v, hasNext, err := d.cur.readValue(';')
			if err != nil {
				return nil, err
			}

			values = append(values, v)
			if !hasNext {
				return values, nil
			}
		}

	default:
		panic("unreachable")
	}
}

func (d PathDecoder) DecodeObject(f func(field, value string) error) error {
	switch d.style {
	case PathStyleSimple:
		if d.explode {
			const kvSep, fieldSep = '=', ','
			return d.decodeObject(kvSep, fieldSep, f)
		}

		const kvSep, fieldSep = ',', ','
		return d.decodeObject(kvSep, fieldSep, f)

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return fmt.Errorf("value must begin with '.'")
		}

		if d.explode {
			const kvSep, fieldSep = '=', '.'
			return d.decodeObject(kvSep, fieldSep, f)
		}

		const kvSep, fieldSep = ',', ','
		return d.decodeObject(kvSep, fieldSep, f)

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return fmt.Errorf("value must begin with ';'")
		}

		if !d.explode {
			name, err := d.cur.readAt('=')
			if err != nil {
				return err
			}

			if name != d.param {
				return fmt.Errorf("expect param '%s', got '%s'", d.param, name)
			}

			const kvSep, fieldSep = ',', ','
			return d.decodeObject(kvSep, fieldSep, f)
		}

		const kvSep, fieldSep = '=', ';'
		return d.decodeObject(kvSep, fieldSep, f)

	default:
		panic("unreachable")
	}
}

func (d PathDecoder) decodeObject(kvSep, fieldSep rune, f func(field, value string) error) error {
	var (
		fname string
		field = true
	)

	for {
		until := fieldSep
		if field {
			until = kvSep
		}

		v, hasNext, err := d.cur.readValue(until)
		if err != nil {
			return err
		}

		if field {
			fname = v
			field = false
			continue
		}

		field = true
		if err := f(fname, v); err != nil {
			return err
		}

		if !hasNext {
			return nil
		}
	}
}
