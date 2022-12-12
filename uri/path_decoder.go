package uri

import (
	"io"

	"github.com/go-faster/errors"
)

type PathDecoder struct {
	cur *cursor

	param   string
	style   PathStyle
	explode bool
}

type PathDecoderConfig struct {
	Param   string // Parameter name
	Value   string // chi.URLParam(r, "paramName")
	Style   PathStyle
	Explode bool
}

func NewPathDecoder(cfg PathDecoderConfig) *PathDecoder {
	return &PathDecoder{
		cur:     &cursor{src: cfg.Value},
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (d *PathDecoder) DecodeValue() (string, error) {
	switch d.style {
	case PathStyleSimple:
		return d.cur.readAll()

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return "", errors.New(`label style value must begin with "."`)
		}
		return d.cur.readAll()

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return "", errors.New(`label style value must begin with ";"`)
		}

		param, err := d.cur.readUntil('=')
		if err != nil {
			return "", err
		}

		if param != d.param {
			return "", errors.Errorf("invalid param name %q", param)
		}

		return d.cur.readAll()

	default:
		panic("unreachable")
	}
}

func (d *PathDecoder) DecodeArray(f func(d Decoder) error) error {
	switch d.style {
	case PathStyleSimple:
		return parseArray(d.cur, ',', f)

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return errors.New(`value must begin with "."`)
		}

		delim := byte(',')
		if d.explode {
			delim = '.'
		}
		return parseArray(d.cur, delim, f)

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return errors.New(`value must begin with "."`)
		}

		if !d.explode {
			param, hasNext, err := d.cur.readValue('=')
			if err != nil {
				return err
			}

			if param != d.param {
				return errors.Errorf("unexpected param name: %q", param)
			}

			if !hasNext {
				return io.EOF
			}

			return parseArray(d.cur, ',', f)
		}

		for {
			param, hasNext, err := d.cur.readValue('=')
			if err != nil {
				return err
			}

			if param != d.param {
				return errors.Errorf("unexpected param name: %q", param)
			}

			if !hasNext {
				return io.EOF
			}

			value, hasNext, err := d.cur.readValue(';')
			if err != nil {
				return err
			}

			if err := f(&constval{v: value}); err != nil {
				return err
			}

			if !hasNext {
				return nil
			}
		}

	default:
		panic("unreachable")
	}
}

func (d *PathDecoder) DecodeFields(f func(name string, d Decoder) error) error {
	adapter := func(k, v string) error { return f(k, &constval{v: v}) }
	switch d.style {
	case PathStyleSimple:
		if d.explode {
			const kvSep, fieldSep = '=', ','
			return decodeObject(d.cur, kvSep, fieldSep, adapter)
		}

		const kvSep, fieldSep = ',', ','
		return decodeObject(d.cur, kvSep, fieldSep, adapter)

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return errors.New(`value must begin with "."`)
		}

		if d.explode {
			const kvSep, fieldSep = '=', '.'
			return decodeObject(d.cur, kvSep, fieldSep, adapter)
		}

		const kvSep, fieldSep = ',', ','
		return decodeObject(d.cur, kvSep, fieldSep, adapter)

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return errors.New(`value must begin with ";"`)
		}

		if !d.explode {
			name, err := d.cur.readUntil('=')
			if err != nil {
				return err
			}

			if name != d.param {
				return errors.Errorf("expect param %q, got %q", d.param, name)
			}

			const kvSep, fieldSep = ',', ','
			return decodeObject(d.cur, kvSep, fieldSep, adapter)
		}

		const kvSep, fieldSep = '=', ';'
		return decodeObject(d.cur, kvSep, fieldSep, adapter)

	default:
		panic("unreachable")
	}
}

func parseArray(cur *cursor, delim byte, f func(d Decoder) error) error {
	for {
		value, hasNext, err := cur.readValue(delim)
		if err != nil {
			return err
		}

		if err := f(&constval{v: value}); err != nil {
			return err
		}

		if !hasNext {
			return nil
		}
	}
}
