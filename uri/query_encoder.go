package uri

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
	return e.EncodeString(strconv.FormatBool(v))
}

func (e *QueryEncoder) EncodeInt64(v int64) string {
	return e.EncodeString(strconv.FormatInt(v, 10))
}

func (e *QueryEncoder) EncodeInt32(v int32) string {
	return e.EncodeInt64(int64(v))
}

func (e *QueryEncoder) EncodeInt(v int) string {
	return e.EncodeInt64(int64(v))
}

func (e *QueryEncoder) EncodeFloat64(v float64) string {
	return e.EncodeString(strconv.FormatFloat(v, 'f', 10, 64))
}

func (e *QueryEncoder) EncodeFloat32(v float32) string {
	return e.EncodeFloat64(float64(v))
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

func (e *QueryEncoder) EncodeInt32Array(vs []int32) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatInt(int64(v), 10))
	}
	return e.EncodeStringArray(strs)
}

func (e *QueryEncoder) EncodeIntArray(vs []int) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatInt(int64(v), 10))
	}
	return e.EncodeStringArray(strs)
}

func (e *QueryEncoder) EncodeFloat64Array(vs []float64) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatFloat(v, 'f', 10, 64))
	}
	return e.EncodeStringArray(strs)
}

func (e *QueryEncoder) EncodeFloat32Array(vs []float32) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatFloat(float64(v), 'f', 10, 64))
	}
	return e.EncodeStringArray(strs)
}

func (e *QueryEncoder) EncodeTime(v time.Time) string {
	return e.EncodeString(v.Format(time.RFC3339))
}
