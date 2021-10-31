package uri

import (
	"fmt"
	"io"
)

type PathDecoder struct {
	cur *cursor

	param   string
	style   PathStyle
	explode bool
}

func (d *PathDecoder) Value() (string, error) {
	return d.cur.readAll()
}

func (d *PathDecoder) Array(f func(d Decoder) error) error {
	switch d.style {
	case PathStyleSimple:
		return parseArray(d.cur, ',', f)

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return fmt.Errorf("value must begin with '.'")
		}

		delim := ','
		if d.explode {
			delim = '.'
		}
		return parseArray(d.cur, delim, f)

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return fmt.Errorf("value must begin with '.'")
		}

		if !d.explode {
			param, hasNext, err := d.cur.readValue('=')
			if err != nil {
				return err
			}

			if param != d.param {
				return fmt.Errorf("unexpected param name: '%s'", param)
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
				return fmt.Errorf("unexpected param name: '%s'", param)
			}

			if !hasNext {
				return io.EOF
			}

			value, hasNext, err := d.cur.readValue(';')
			if err != nil {
				return err
			}

			if err := f(&pathDecoderValue{
				v: value,
			}); err != nil {
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

func (d *PathDecoder) Fields(f func(name string, d Decoder) error) error {
	adapter := func(k, v string) error { return f(k, &pathDecoderValue{v: v}) }
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
			return fmt.Errorf("value must begin with '.'")
		}

		if d.explode {
			const kvSep, fieldSep = '=', '.'
			return decodeObject(d.cur, kvSep, fieldSep, adapter)
		}

		const kvSep, fieldSep = ',', ','
		return decodeObject(d.cur, kvSep, fieldSep, adapter)

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
			return decodeObject(d.cur, kvSep, fieldSep, adapter)
		}

		const kvSep, fieldSep = '=', ';'
		return decodeObject(d.cur, kvSep, fieldSep, adapter)

	default:
		panic("unreachable")
	}
}

func parseArray(cur *cursor, delim rune, f func(d Decoder) error) error {
	for {
		value, hasNext, err := cur.readValue(delim)
		if err != nil {
			return err
		}

		if err := f(&pathDecoderValue{
			v: value,
		}); err != nil {
			return err
		}

		if !hasNext {
			return nil
		}
	}
}
