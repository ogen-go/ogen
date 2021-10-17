package uri

import "strconv"

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

func (e PathEncoder) EncodeString(v string) string {
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

func (e PathEncoder) EncodeStringArray(vs []string) string {
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

func (e PathEncoder) EncodeInt64(v int64) string {
	return e.EncodeString(strconv.FormatInt(v, 10))
}

func (e PathEncoder) EncodeInt32(v int32) string {
	return e.EncodeInt64(int64(v))
}

func (e PathEncoder) EncodeInt(v int) string {
	return e.EncodeInt64(int64(v))
}

func (e PathEncoder) EncodeFloat64(v float64) string {
	return e.EncodeString(strconv.FormatFloat(v, 'f', 10, 64))
}

func (e PathEncoder) EncodeFloat32(v float32) string {
	return e.EncodeFloat64(float64(v))
}

func (e PathEncoder) EncodeBool(v bool) string {
	return e.EncodeString(strconv.FormatBool(v))
}

func (e PathEncoder) EncodeInt64Array(vs []int64) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatInt(v, 10))
	}
	return e.EncodeStringArray(strs)
}

func (e PathEncoder) EncodeInt32Array(vs []int32) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatInt(int64(v), 10))
	}
	return e.EncodeStringArray(strs)
}

func (e PathEncoder) EncodeIntArray(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatInt(int64(v), 10))
	}
	return e.EncodeStringArray(strs)
}

func (e PathEncoder) EncodeFloat64Array(vs []float64) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatFloat(v, 'f', 10, 64))
	}
	return e.EncodeStringArray(strs)
}

func (e PathEncoder) EncodeFloat32Array(vs []float32) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatFloat(float64(v), 'f', 10, 64))
	}
	return e.EncodeStringArray(strs)
}

func (e PathEncoder) EncodeBoolArray(vs []bool) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.FormatBool(v))
	}
	return e.EncodeStringArray(strs)
}
