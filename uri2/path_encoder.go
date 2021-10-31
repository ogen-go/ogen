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

	t      byte // 0 - not set, 1 - value, 2 - array, 3 - object
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
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e *PathEncoder) Value(v string) error {
	e.assert(1)
	e.val = v
	return nil
}

func (e *PathEncoder) Array(f func(e Encoder) error) error {
	arr := &pathArrayEncoder{}
	if err := f(arr); err != nil {
		return err
	}

	if !arr.set {
		return nil
	}

	e.assert(2)
	return nil
}

func (e *PathEncoder) Field(name string, f func(e Encoder) error) error {
	fenc := &pathFieldEncoder{}
	if err := f(fenc); err != nil {
		return err
	}

	if !fenc.set {
		return nil
	}

	e.assert(3)
	e.fields = append(e.fields, Field{
		Name:  name,
		Value: fenc.value,
	})
	return nil
}

func (e *PathEncoder) assert(expect byte) {
	if e.t == 0 {
		e.t = expect
		return
	}

	if e.t != expect {
		// TODO: Stringify
		panic(fmt.Sprintf("%d called with %d", e.t, expect))
	}

	switch e.t {
	case 1:
		panic("multiple e.Value calls")
	case 2:
		panic("multiple e.Array calls")
	case 3:
		// e.Field can be called multiple times.
	}
}

func (e *PathEncoder) Result() string {
	switch e.t {
	case 0:
		panic("encoder does not called")
	case 1:
		return e.value()
	case 2:
		return e.array()
	case 3:
		return e.object()
	default:
		panic("unreachable")
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
		return e.val
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
