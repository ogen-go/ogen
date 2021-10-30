package uri

import (
	"fmt"
	"strings"
)

type QueryEncoder struct {
	param   string
	style   QueryStyle // immutable
	explode bool       // immutable
}

type QueryEncoderConfig struct {
	Param   string
	Style   QueryStyle
	Explode bool
}

func NewQueryEncoder(cfg QueryEncoderConfig) *QueryEncoder {
	return &QueryEncoder{
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e *QueryEncoder) EncodeValue(v string) string {
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

func (e *QueryEncoder) EncodeArray(vs []string) []string {
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

func (e *QueryEncoder) EncodeObject(fields []Field) []string {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			out := make([]string, 0, len(fields))
			for _, f := range fields {
				out = append(out, f.Name+"="+f.Value)
			}
			return out
		}

		var out string
		for i, f := range fields {
			out += f.Name + "," + f.Value
			if i != len(fields)-1 {
				out += ","
			}
		}
		return []string{out}

	case QueryStyleSpaceDelimited:
		panic("object cannot have spaceDelimited style")

	case QueryStylePipeDelimited:
		panic("object cannot have pipeDelimited style")

	case QueryStyleDeepObject:
		if !e.explode {
			panic("invalid deepObject style configuration")
		}

		out := make([]string, 0, len(fields))
		for _, f := range fields {
			out = append(out, e.param+"["+f.Name+"]="+f.Value)
		}
		return out

	default:
		panic("unreachable")
	}
}
