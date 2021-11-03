package uri

import (
	"fmt"
	"strings"
)

type QueryEncoder struct {
	param   string
	style   QueryStyle // immutable
	explode bool       // immutable
	*scraper
}

type QueryEncoderConfig struct {
	Param   string
	Style   QueryStyle
	Explode bool
}

func NewQueryEncoder(cfg QueryEncoderConfig) *QueryEncoder {
	return &QueryEncoder{
		scraper: newScraper(),
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e *QueryEncoder) Result() []string {
	switch e.typ {
	case typeNotSet:
		return nil
	case typeValue:
		return e.value()
	case typeArray:
		return e.array()
	case typeObject:
		return e.object()
	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) value() []string {
	switch e.style {
	case QueryStyleForm:
		return []string{e.val}
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) array() []string {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			return e.items
		}

		return []string{strings.Join(e.items, ",")}

	case QueryStyleSpaceDelimited:
		if e.explode {
			return e.items
		}

		panic("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if e.explode {
			return e.items
		}

		return []string{strings.Join(e.items, "|")}

	case QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for arrays", e.style))

	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) object() []string {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			out := make([]string, 0, len(e.fields))
			for _, f := range e.fields {
				out = append(out, f.Name+"="+f.Value)
			}
			return out
		}

		var out string
		for i, f := range e.fields {
			out += f.Name + "," + f.Value
			if i != len(e.fields)-1 {
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

		out := make([]string, 0, len(e.fields))
		for _, f := range e.fields {
			out = append(out, e.param+"["+f.Name+"]="+f.Value)
		}
		return out

	default:
		panic("unreachable")
	}
}
