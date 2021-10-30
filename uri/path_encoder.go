package uri

import "strings"

type PathEncoder struct {
	param   string    // immutable
	style   PathStyle // immutable
	explode bool      // immutable
}

type PathEncoderConfig struct {
	Param   string
	Style   PathStyle
	Explode bool
}

func NewPathEncoder(cfg PathEncoderConfig) PathEncoder {
	return PathEncoder{
		param:   cfg.Param,
		style:   cfg.Style,
		explode: cfg.Explode,
	}
}

func (e PathEncoder) EncodeValue(v string) string {
	switch e.style {
	case PathStyleSimple:
		return v
	case PathStyleLabel:
		return "." + v
	case PathStyleMatrix:
		return ";" + e.param + "=" + v
	default:
		panic("unreachable")
	}
}

func (e PathEncoder) EncodeArray(vs []string) string {
	switch e.style {
	case PathStyleSimple:
		var result []rune
		ll := len(vs)
		for i := 0; i < ll; i++ {
			result = append(result, []rune(vs[i])...)
			if i != ll-1 {
				result = append(result, ',')
			}
		}
		return string(result)
	case PathStyleLabel:
		result := []rune{'.'}
		ll := len(vs)
		delim := ','
		if e.explode {
			delim = '.'
		}
		for i := 0; i < ll; i++ {
			result = append(result, []rune(vs[i])...)
			if i != ll-1 {
				result = append(result, delim)
			}
		}
		return string(result)
	case PathStyleMatrix:
		if !e.explode {
			var result []rune
			result = append(result, ';')
			result = append(result, []rune(e.param)...)
			result = append(result, '=')

			ll := len(vs)
			for i := 0; i < ll; i++ {
				result = append(result, []rune(vs[i])...)
				if i != ll-1 {
					result = append(result, ',')
				}
			}
			return string(result)
		}

		var result []rune
		ll := len(vs)
		for i := 0; i < ll; i++ {
			result = append(result, ';')
			result = append(result, []rune(e.param)...)
			result = append(result, '=')
			result = append(result, []rune(vs[i])...)
		}

		return string(result)
	default:
		panic("unreachable")
	}
}

type Field struct {
	Name  string
	Value string
}

func (e PathEncoder) EncodeObject(fields []Field) string {
	switch e.style {
	case PathStyleSimple:
		if e.explode {
			const kvSep, fieldSep = '=', ','
			return encodeObject(kvSep, fieldSep, fields)
		}

		const kvSep, fieldSep = ',', ','
		return encodeObject(kvSep, fieldSep, fields)

	case PathStyleLabel:
		kvSep, fieldSep := ',', ','
		if e.explode {
			kvSep, fieldSep = '=', '.'
		}
		return "." + encodeObject(kvSep, fieldSep, fields)

	case PathStyleMatrix:
		var result string

		if !e.explode {
			result += e.param + "="
			const kvSep, fieldSep = ',', ','
			result += encodeObject(kvSep, fieldSep, fields)
		} else {
			const kvSep, fieldSep = '=', ';'
			result += encodeObject(kvSep, fieldSep, fields)
		}

		return ";" + result

	default:
		panic("unreachable")
	}
}

func encodeObject(kvSep, fieldSep rune, fields []Field) string {
	var elems []string
	for _, f := range fields {
		elems = append(elems, f.Name+string(kvSep)+f.Value)
	}
	return strings.Join(elems, string(fieldSep))
}
