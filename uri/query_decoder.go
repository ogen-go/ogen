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
