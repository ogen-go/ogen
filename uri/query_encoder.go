package uri

import (
	"fmt"
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

func (e *QueryEncoder) EncodeStrings(vs []string) []string {
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
