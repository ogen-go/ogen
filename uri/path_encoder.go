package uri

import (
	"fmt"
	"net/url"
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
	*receiver
}

type PathEncoderConfig struct {
	Param   string
	Style   PathStyle
	Explode bool
}

func NewPathEncoder(cfg PathEncoderConfig) *PathEncoder {
	return &PathEncoder{
		receiver: newReceiver(),
		param:    url.PathEscape(cfg.Param),
		style:    cfg.Style,
		explode:  cfg.Explode,
	}
}

func (e *PathEncoder) checkParam() error {
	switch e.style {
	case PathStyleMatrix:
		return checkNotContains(e.param, "=")
	default:
		return nil
	}
}

func (e *PathEncoder) Result() (r string, _ error) {
	if err := e.checkParam(); err != nil {
		return "", err
	}
	switch e.typ {
	case typeNotSet:
		panic("encoder was not called, no value")
	case typeValue:
		e.val = url.PathEscape(e.val)
		return e.value()
	case typeArray:
		chars := ","
		switch e.style {
		case PathStyleLabel:
			if e.explode {
				chars = "."
			}
		case PathStyleMatrix:
			if e.explode {
				chars = ";"
			}
		}

		for i, val := range e.items {
			if err := checkNotContains(val, chars); err != nil {
				return "", err
			}
			e.items[i] = url.PathEscape(val)
		}
		return e.array()
	case typeObject:
		type styleKey struct {
			style   PathStyle
			explode bool
		}
		type objectChars struct {
			kvSep, fieldSep string
		}

		chars, ok := map[styleKey]objectChars{
			{PathStyleSimple, false}: {",", ","},
			{PathStyleSimple, true}:  {"=", ","},

			{PathStyleLabel, false}: {",", ","},
			{PathStyleLabel, true}:  {"=", "."},

			{PathStyleMatrix, false}: {",", ","},
			{PathStyleMatrix, true}:  {"=", ";"},
		}[styleKey{e.style, e.explode}]

		for i, f := range e.fields {
			if ok {
				if err := checkNotContains(f.Name, chars.kvSep); err != nil {
					return "", err
				}
				if err := checkNotContains(f.Value, chars.fieldSep); err != nil {
					return "", err
				}
			}
			e.fields[i] = Field{
				Name:  url.PathEscape(f.Name),
				Value: url.PathEscape(f.Value),
			}
		}
		return e.object()
	default:
		panic(fmt.Sprintf("unexpected value type: %T", e.typ))
	}
}

func (e *PathEncoder) value() (string, error) {
	switch e.style {
	case PathStyleSimple:
		return e.val, nil
	case PathStyleLabel:
		return "." + e.val, nil
	case PathStyleMatrix:
		return ";" + e.param + "=" + e.val, nil
	default:
		panic("unreachable")
	}
}

func (e *PathEncoder) array() (string, error) {
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
		return string(result), nil
	case PathStyleLabel:
		if !e.explode {
			return "." + strings.Join(e.items, ","), nil
		}
		return "." + strings.Join(e.items, "."), nil
	case PathStyleMatrix:
		if !e.explode {
			return ";" + e.param + "=" + strings.Join(e.items, ","), nil
		}
		prefix := ";" + e.param + "="
		return prefix + strings.Join(e.items, prefix), nil
	default:
		panic("unreachable")
	}
}

func (e *PathEncoder) object() (string, error) {
	switch e.style {
	case PathStyleSimple:
		if e.explode {
			const kvSep, fieldSep = '=', ','
			return encodeObject(kvSep, fieldSep, e.fields), nil
		}

		const kvSep, fieldSep = ',', ','
		return encodeObject(kvSep, fieldSep, e.fields), nil

	case PathStyleLabel:
		var kvSep, fieldSep byte = ',', ','
		if e.explode {
			kvSep, fieldSep = '=', '.'
		}
		return "." + encodeObject(kvSep, fieldSep, e.fields), nil

	case PathStyleMatrix:
		if !e.explode {
			const kvSep, fieldSep = ',', ','
			return ";" + e.param + "=" + encodeObject(kvSep, fieldSep, e.fields), nil
		}
		const kvSep, fieldSep = '=', ';'
		return ";" + encodeObject(kvSep, fieldSep, e.fields), nil

	default:
		panic("unreachable")
	}
}
