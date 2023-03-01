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

func (e *queryParamEncoder) serialize() error {
	switch e.typ {
	case typeNotSet:
		return nil
	case typeValue:
		return e.encodeValue()
	case typeArray:
		return e.encodeArray()
	case typeObject:
		return e.encodeObject()
	default:
		panic("unreachable")
	}
}

func (e *queryParamEncoder) encodeValue() error {
	switch e.style {
	case QueryStyleForm:
		e.values[e.paramName] = []string{e.val}
		return nil
	case QueryStyleSpaceDelimited,
		QueryStylePipeDelimited,
		QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for primitive values", e.style))
	default:
		panic("unreachable")
	}
}

func (e *queryParamEncoder) encodeArray() error {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			e.values[e.paramName] = e.items
			return nil
		}

		const sep = ","
		for _, v := range e.items {
			if err := checkNotContains(v, sep); err != nil {
				return err
			}
		}
		e.values[e.paramName] = []string{strings.Join(e.items, sep)}
		return nil

	case QueryStyleSpaceDelimited:
		if e.explode {
			e.values[e.paramName] = e.items
			return nil
		}

		panic("spaceDelimited with explode: false not supported")

	case QueryStylePipeDelimited:
		if e.explode {
			e.values[e.paramName] = e.items
			return nil
		}

		const sep = "|"
		for _, v := range e.items {
			if err := checkNotContains(v, sep); err != nil {
				return err
			}
		}
		e.values[e.paramName] = []string{strings.Join(e.items, sep)}
		return nil

	case QueryStyleDeepObject:
		panic(fmt.Sprintf("style %q cannot be used for arrays", e.style))

	default:
		panic("unreachable")
	}
}

func (e *queryParamEncoder) encodeObject() error {
	switch e.style {
	case QueryStyleForm:
		if e.explode {
			for _, f := range e.fields {
				e.values[f.Name] = []string{f.Value}
			}
			return nil
		}

		const (
			kvSep    = ","
			fieldSep = ","
		)
		var out string

		for i, f := range e.fields {
			if err := checkNotContains(f.Name, kvSep); err != nil {
				return err
			}
			if err := checkNotContains(f.Value, fieldSep); err != nil {
				return err
			}

			out += f.Name + fieldSep + f.Value
			if i != len(e.fields)-1 {
				out += kvSep
			}
		}

		e.values[e.paramName] = []string{out}
		return nil

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

		return nil

	default:
		panic("unreachable")
	}
}
