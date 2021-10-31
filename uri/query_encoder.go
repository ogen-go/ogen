package uri

import (
	"fmt"
	"strings"
)

type QueryEncoder struct {
	param   string
	style   QueryStyle // immutable
	explode bool       // immutable

	t      vtyp
	val    string
	items  []string
	fields []Field
}

type QueryEncoderConfig struct {
	Param   string
	Style   QueryStyle
	Explode bool
}

func NewQueryEncoder(cfg QueryEncoderConfig) *QueryEncoder {
	return &QueryEncoder{
		t:       vtNotSet,
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e *QueryEncoder) Value(v string) error {
	if e.t != vtNotSet && e.t != vtValue {
		return fmt.Errorf("encode value: already encoded as %s", e.t)
	}

	e.t = vtValue
	e.val = v
	return nil
}

func (e *QueryEncoder) Array(f func(Encoder) error) error {
	if e.t != vtNotSet && e.t != vtArray {
		return fmt.Errorf("encode value: already encoded as %s", e.t)
	}

	e.t = vtArray
	arr := &arrayEncoder{}
	if err := f(arr); err != nil {
		return err
	}

	e.items = arr.items
	return nil
}

func (e *QueryEncoder) Field(name string, f func(Encoder) error) error {
	if e.t != vtNotSet && e.t != vtObject {
		return fmt.Errorf("encode value: already encoded as %s", e.t)
	}

	e.t = vtObject
	fenc := &fieldEncoder{}
	if err := f(fenc); err != nil {
		return err
	}

	if !fenc.set {
		return nil
	}

	e.fields = append(e.fields, Field{
		Name:  name,
		Value: fenc.value,
	})
	return nil
}

func (e *QueryEncoder) Result() []string {
	switch e.t {
	case vtNotSet:
		panic("encoder was not called, no value")
	case vtValue:
		return e.value()
	case vtArray:
		return e.array()
	case vtObject:
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
		panic(fmt.Sprintf("style '%s' cannot be used for primitive values", e.style))
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
		panic(fmt.Sprintf("style '%s' cannot be used for arrays", e.style))

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
