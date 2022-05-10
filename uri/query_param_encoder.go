package uri

import (
	"fmt"
	"net/url"
	"strings"
)

type QueryStyle string

const (
	QueryStyleForm           QueryStyle = "form"
	QueryStyleSpaceDelimited QueryStyle = "spaceDelimited"
	QueryStylePipeDelimited  QueryStyle = "pipeDelimited"
	QueryStyleDeepObject     QueryStyle = "deepObject"
)

type queryParamEncoder struct {
	*receiver
	values url.Values

	paramName string     // immutable
	style     QueryStyle // immutable
	explode   bool       // immutable
}

func (e *queryParamEncoder) serialize() {
	switch e.typ {
	case typeNotSet:
		return
	case typeValue:
		e.encodeValue()
	case typeArray:
		e.encodeArray()
	case typeObject:
		e.encodeObject()
	default:
		panic("unreachable")
	}
}

func (e *queryParamEncoder) encodeValue() {
	switch e.style {
	case QueryStyleForm:
		e.values[e.paramName] = []string{e.val}
		return
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *queryParamEncoder) encodeArray() {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			e.values[e.paramName] = e.items
			return
		}

		e.values[e.paramName] = []string{strings.Join(e.items, ",")}
		return

	case QueryStyleSpaceDelimited:
		if e.explode {
			e.values[e.paramName] = e.items
			return
		}

		panic("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if e.explode {
			e.values[e.paramName] = e.items
			return
		}

		e.values[e.paramName] = []string{strings.Join(e.items, "|")}
		return

	case QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for arrays", e.style))

	default:
		panic("unreachable")
	}
}

func (e *queryParamEncoder) encodeObject() {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			for _, f := range e.fields {
				e.values[f.Name] = []string{f.Value}
			}
			return
		}

		var out string
		for i, f := range e.fields {
			out += f.Name + "," + f.Value
			if i != len(e.fields)-1 {
				out += ","
			}
		}

		e.values[e.paramName] = []string{out}
		return

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

		return

	default:
		panic("unreachable")
	}
}
