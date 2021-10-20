package uri

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type QueryStyle string

const (
	QueryStyleForm           QueryStyle = "form"
	QueryStyleSpaceDelimited QueryStyle = "spaceDelimited"
	QueryStylePipeDelimited  QueryStyle = "pipeDelimited"
	QueryStyleDeepObject     QueryStyle = "deepObject"
)

type QueryDecoder struct {
	src []string // r.URL.Query()["param"]

	style   QueryStyle // immutable
	explode bool       // immutable
}

type QueryDecoderConfig struct {
	Values  []string
	Style   QueryStyle
	Explode bool
}

func NewQueryDecoder(cfg QueryDecoderConfig) *QueryDecoder {
	return &QueryDecoder{
		src:     cfg.Values,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (d *QueryDecoder) DecodeString() (string, error) {
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

func (d *QueryDecoder) DecodeInt64() (int64, error) {
	s, err := d.DecodeString()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(s, 10, 64)
}

func (d *QueryDecoder) DecodeInt32() (int32, error) {
	v, err := d.DecodeInt64()
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func (d *QueryDecoder) DecodeInt() (int, error) {
	v, err := d.DecodeInt64()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func (d *QueryDecoder) DecodeFloat64() (float64, error) {
	s, err := d.DecodeString()
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(s, 64)
}

func (d *QueryDecoder) DecodeFloat32() (float32, error) {
	v, err := d.DecodeFloat64()
	if err != nil {
		return 0, err
	}
	return float32(v), nil
}

func (d *QueryDecoder) DecodeBool() (bool, error) {
	s, err := d.DecodeString()
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(s)
}

func (d *QueryDecoder) DecodeTime() (time.Time, error) {
	s, err := d.DecodeString()
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, s)
}

func (d *QueryDecoder) decodeArray(push func(string) error) error {
	if len(d.src) < 1 {
		return fmt.Errorf("empty array")
	}

	switch d.style {
	case QueryStyleForm:
		if d.explode {
			for _, v := range d.src {
				if err := push(v); err != nil {
					return err
				}
			}
			return nil
		}

		if len(d.src) != 1 {
			return fmt.Errorf("invalid value")
		}

		// TODO: use cursor
		for _, v := range strings.Split(d.src[0], ",") {
			if err := push(v); err != nil {
				return err
			}
		}
		return nil

	case QueryStyleSpaceDelimited:
		if d.explode {
			for _, v := range d.src {
				if err := push(v); err != nil {
					return err
				}
			}
			return nil
		}

		if len(d.src) != 1 {
			return fmt.Errorf("invalid value")
		}

		return fmt.Errorf("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if d.explode {
			for _, v := range d.src {
				if err := push(v); err != nil {
					return err
				}
			}
			return nil
		}

		if len(d.src) != 1 {
			return fmt.Errorf("invalid value")
		}

		// TODO: use cursor
		for _, v := range strings.Split(d.src[0], "|") {
			if err := push(v); err != nil {
				return err
			}
		}
		return nil

	case QueryStyleDeepObject:
		return fmt.Errorf("style '%s' cannot be used for arrays", d.style)

	default:
		panic("unreachable")
	}
}

func (d *QueryDecoder) DecodeStringArray() ([]string, error) {
	var values []string
	if err := d.decodeArray(func(s string) error {
		values = append(values, s)
		return nil
	}); err != nil {
		return nil, err
	}
	return values, nil
}

func (d *QueryDecoder) DecodeInt64Array() ([]int64, error) {
	var values []int64
	if err := d.decodeArray(func(s string) error {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		values = append(values, v)
		return nil
	}); err != nil {
		return nil, err
	}
	return values, nil
}

func (d *QueryDecoder) DecodeInt32Array() ([]int32, error) {
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

func (d *QueryDecoder) DecodeIntArray() ([]int, error) {
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

func (d *QueryDecoder) DecodeFloat64Array() ([]float64, error) {
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

func (d *QueryDecoder) DecodeFloat32Array() ([]float32, error) {
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

func (d *QueryDecoder) DecodeBoolArray() ([]bool, error) {
	var values []bool
	if err := d.decodeArray(func(s string) error {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		values = append(values, v)
		return nil
	}); err != nil {
		return nil, err
	}
	return values, nil
}
