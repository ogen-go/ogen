package uri

import (
	"fmt"
	"strconv"
)

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

func (d PathDecoder) DecodeInt64() (int64, error) {
	str, err := d.DecodeString()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(str, 10, 64)
}

func (d PathDecoder) DecodeInt32() (int32, error) {
	v, err := d.DecodeInt64()
	if err != nil {
		return 0, err
	}

	return int32(v), nil
}

func (d PathDecoder) DecodeInt() (int, error) {
	v, err := d.DecodeInt64()
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func (d PathDecoder) DecodeFloat64() (float64, error) {
	str, err := d.DecodeString()
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(str, 64)
}

func (d PathDecoder) DecodeFloat32() (float32, error) {
	v, err := d.DecodeFloat64()
	if err != nil {
		return 0, err
	}

	return float32(v), nil
}

func (d PathDecoder) DecodeBool() (bool, error) {
	str, err := d.DecodeString()
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(str)
}

func (d PathDecoder) DecodeStringArray() ([]string, error) {
	var values []string
	if err := d.decodeArray(func(s string) error {
		values = append(values, s)
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) DecodeBoolArray() ([]bool, error) {
	var values []bool
	if err := d.decodeArray(func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}

		values = append(values, b)
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) DecodeInt64Array() ([]int64, error) {
	var values []int64
	if err := d.decodeArray(func(s string) error {
		b, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		values = append(values, b)
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) DecodeInt32Array() ([]int32, error) {
	var values []int32
	if err := d.decodeArray(func(s string) error {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		values = append(values, int32(v))
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) DecodeIntArray() ([]int, error) {
	var values []int
	if err := d.decodeArray(func(s string) error {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		values = append(values, int(v))
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) DecodeFloat64Array() ([]float64, error) {
	var values []float64
	if err := d.decodeArray(func(s string) error {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}

		values = append(values, float64(v))
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) DecodeFloat32Array() ([]float32, error) {
	var values []float32
	if err := d.decodeArray(func(s string) error {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}

		values = append(values, float32(v))
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}

func (d PathDecoder) decodeArray(push func(string) error) error {
	switch d.style {
	case PathStyleSimple:
		for {
			v, hasNext, err := d.cur.readValue(',')
			if err != nil {
				return err
			}

			if err := push(v); err != nil {
				return err
			}
			if !hasNext {
				return nil
			}
		}

	case PathStyleLabel:
		if !d.cur.eat('.') {
			return fmt.Errorf("value must begin with '.'")
		}

		delim := ','
		if d.explode {
			delim = '.'
		}

		for {
			v, hasNext, err := d.cur.readValue(delim)
			if err != nil {
				return err
			}

			if err := push(v); err != nil {
				return err
			}
			if !hasNext {
				return nil
			}
		}

	case PathStyleMatrix:
		if !d.cur.eat(';') {
			return fmt.Errorf("value must begin with ';'")
		}

		if !d.explode {
			param, err := d.cur.readAt('=')
			if err != nil {
				return err
			}

			if param != d.param {
				return fmt.Errorf("invalid param name: '%s'", param)
			}

			for {
				v, hasNext, err := d.cur.readValue(',')
				if err != nil {
					return err
				}

				if err := push(v); err != nil {
					return err
				}
				if !hasNext {
					return nil
				}
			}
		}

		for {
			param, err := d.cur.readAt('=')
			if err != nil {
				return err
			}

			if param != d.param {
				return fmt.Errorf("invalid param name '%s'", param)
			}

			v, hasNext, err := d.cur.readValue(';')
			if err != nil {
				return err
			}

			if err := push(v); err != nil {
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
