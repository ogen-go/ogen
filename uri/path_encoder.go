package uri

import (
	"fmt"
	"strings"
)

type PathStyle string

const (
	PathStyleSimple PathStyle = "simple"
	PathStyleLabel  PathStyle = "label"
	PathStyleMatrix PathStyle = "matrix"
)

func (s PathStyle) String() string { return string(s) }

type PathEncoder struct {
	param   string    // immutable
	style   PathStyle // immutable
	explode bool      // immutable

	t      vtyp
	val    string
	items  []string
	fields []Field
}

type PathEncoderConfig struct {
	Param   string
	Style   PathStyle
	Explode bool
}

func NewPathEncoder(cfg PathEncoderConfig) *PathEncoder {
	return &PathEncoder{
		t:       vtNotSet,
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e *PathEncoder) Value(v string) error {
	if e.t != vtNotSet && e.t != vtValue {
		return fmt.Errorf("encode value: already encoded as %s", e.t)
	}
	e.t = vtValue
	e.val = v
	return nil
}

func (e *PathEncoder) Array(f func(e Encoder) error) error {
	if e.t != vtNotSet && e.t != vtArray {
		return fmt.Errorf("encode array: already encoded as %s", e.t)
	}

	e.t = vtArray
	arr := &arrayEncoder{}
	if err := f(arr); err != nil {
		return err
	}

	e.items = arr.items
	return nil
}

func (e *PathEncoder) Field(name string, f func(e Encoder) error) error {
	if e.t != vtNotSet && e.t != vtObject {
		return fmt.Errorf("encode object: already encoded as %s", e.t)
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

func (e *PathEncoder) Result() string {
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
		panic(fmt.Sprintf("unexpected value type: %T", e.t))
	}
}

func (e *PathEncoder) value() string {
	switch e.style {
	case PathStyleSimple:
		return e.val
	case PathStyleLabel:
		return "." + e.val
	case PathStyleMatrix:
		return ";" + e.param + "=" + e.val
	default:
		panic("unreachable")
	}
}

func (e *PathEncoder) array() string {
	switch e.style {
	case PathStyleSimple:
		var result []rune
		ll := len(e.items)
		for i := 0; i < ll; i++ {
			result = append(result, []rune(e.items[i])...)
			if i != ll-1 {
				result = append(result, ',')
			}
		}
		return string(result)
	case PathStyleLabel:
		if !e.explode {
			return "." + strings.Join(e.items, ",")
		}
		return "." + strings.Join(e.items, ".")
	case PathStyleMatrix:
		if !e.explode {
			return ";" + e.param + "=" + strings.Join(e.items, ",")
		}
		prefix := ";" + e.param + "="
		return prefix + strings.Join(e.items, prefix)
	default:
		panic("unreachable")
	}
}

func (e *PathEncoder) object() string {
	switch e.style {
	case PathStyleSimple:
		if e.explode {
			const kvSep, fieldSep = '=', ','
			return encodeObject(kvSep, fieldSep, e.fields)
		}

		const kvSep, fieldSep = ',', ','
		return encodeObject(kvSep, fieldSep, e.fields)

	case PathStyleLabel:
		kvSep, fieldSep := ',', ','
		if e.explode {
			kvSep, fieldSep = '=', '.'
		}
		return "." + encodeObject(kvSep, fieldSep, e.fields)

	case PathStyleMatrix:
		if !e.explode {
			const kvSep, fieldSep = ',', ','
			return ";" + e.param + "=" + encodeObject(kvSep, fieldSep, e.fields)
		}
		const kvSep, fieldSep = '=', ';'
		return ";" + encodeObject(kvSep, fieldSep, e.fields)

	default:
		panic("unreachable")
	}
}
