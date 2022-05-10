package uri

import (
	"fmt"
	"net/url"
	"strings"
)

type QueryEncoder struct {
	paramName string
	style     QueryStyle // immutable
	explode   bool       // immutable
	*receiver
	values url.Values
}

type QueryEncoderConfig struct {
	Param   string
	Style   QueryStyle
	Explode bool
}

func NewQueryEncoder(cfg QueryEncoderConfig, values url.Values) *QueryEncoder {
	if values == nil {
		values = make(url.Values)
	}
	return &QueryEncoder{
		receiver:  newReceiver(),
		paramName: cfg.Param,
		style:     cfg.Style,
		explode:   cfg.Explode,
		values:    values,
	}
}

func (e *QueryEncoder) Result() url.Values {
	switch e.typ {
	case typeNotSet:
		return e.values
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

func (e *QueryEncoder) value() url.Values {
	switch e.style {
	case QueryStyleForm:
		e.values[e.paramName] = []string{e.val}
		return e.values
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) array() url.Values {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			e.values[e.paramName] = e.items
			return e.values
		}

		e.values[e.paramName] = []string{strings.Join(e.items, ",")}
		return e.values

	case QueryStyleSpaceDelimited:
		if e.explode {
			e.values[e.paramName] = e.items
			return e.values
		}

		panic("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if e.explode {
			e.values[e.paramName] = e.items
			return e.values
		}

		e.values[e.paramName] = []string{strings.Join(e.items, "|")}
		return e.values

	case QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for arrays", e.style))

	default:
		panic("unreachable")
	}
}

func (e *QueryEncoder) object() url.Values {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			for _, f := range e.fields {
				e.values[f.Name] = []string{f.Value}
			}
			return e.values
		}

		var out string
		for i, f := range e.fields {
			out += f.Name + "," + f.Value
			if i != len(e.fields)-1 {
				out += ","
			}
		}

		e.values[e.paramName] = []string{out}
		return e.values

	case QueryStyleSpaceDelimited:
		panic("object cannot have spaceDelimited style")

	case QueryStylePipeDelimited:
		panic("object cannot have pipeDelimited style")

	case QueryStyleDeepObject:
		if !e.explode {
			panic("invalid deepObject style configuration")
		}

		for _, f := range e.fields {
			e.values[e.paramName+"["+f.Name+"]"] = []string{f.Value}
		}

		return e.values

	default:
		panic("unreachable")
	}
}
