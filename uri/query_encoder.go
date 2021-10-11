package uri

import (
	"fmt"
	"strconv"
	"strings"
)

type QueryEncoder struct {
	style   QueryStyle // immutable
	explode bool       // immutable
}

type QueryEncoderConfig struct {
	Style   QueryStyle
	Explode bool
}

func NewQueryEncoder(cfg QueryEncoderConfig) *QueryEncoder {
	return &QueryEncoder{
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e *QueryEncoder) EncodeString(v string) string {
	switch e.style {
	case QueryStyleForm:
		return v
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style '%s' cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) EncodeStringArray(vs []string) []string {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			return vs
		}

		return []string{strings.Join(vs, ",")}

	case QueryStyleSpaceDelimited:
		if e.explode {
			return vs
		}

		panic("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if e.explode {
			return vs
		}

		return []string{strings.Join(vs, "|")}

	case QueryStyleDeepObject:
		panic(fmt.Sprintf("style '%s' cannot be used for arrays", e.style))

	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) EncodeBool(v bool) string {
	switch e.style {
	case QueryStyleForm:
		return strconv.FormatBool(v)
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style '%s' cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) EncodeInt64(v int64) string {
	switch e.style {
	case QueryStyleForm:
		return strconv.FormatInt(v, 10)
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style '%s' cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) EncodeBoolArray(vs []bool) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatBool(v))
	}
	return e.EncodeStringArray(strs)
}

func (e *QueryEncoder) EncodeInt64Array(vs []int64) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatInt(v, 10))
	}
	return e.EncodeStringArray(strs)
}
