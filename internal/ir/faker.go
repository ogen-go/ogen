package ir

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen/validate"
)

func generateInt(p PrimitiveType, va validate.Int) string {
	mul := va.MultipleOf
	if mul <= 0 {
		mul = 1
	}
	switch {
	case !va.MinSet && !va.MaxSet:
		return fmt.Sprintf("%s(%d)", p, mul)
	case va.MinSet && va.MaxSet:
		max := va.Max
		if va.MaxExclusive {
			max--
		}
		min := va.Min
		if va.MinExclusive {
			min++
		}
		for i := min; i <= max; i++ {
			if i%int64(mul) == 0 {
				return fmt.Sprintf("%s(%d)", p, i)
			}
		}
		panic(fmt.Sprintf("unable to generate valid value %+v", va))
	default:
		val := va.Min
		if va.MaxSet {
			val = va.Max
		}
		return fmt.Sprintf("%s(%d)", p, val)
	}
}

func (t *Type) FakeValue() string {
	va := t.Validators
	switch p := t.Primitive; p {
	case String:
		return strconv.Quote(strings.Repeat("a", va.String.MinLength))
	case ByteSlice:
		return `[]byte("[]byte")`
	case Int,
		Int8,
		Int16,
		Int32,
		Int64,
		Uint,
		Uint8,
		Uint16,
		Uint32,
		Uint64:
		return generateInt(p, va.Int)
	case Float32:
		return "float32(0)"
	case Float64:
		return "float64(0)"
	case Time:
		return "time.Now()"
	case Duration:
		return "time.Duration(5 * time.Second)"
	case UUID:
		return "uuid.New()"
	case IP:
		return `net.ParseIP("127.0.0.1")`
	case URL:
		return `url.URL{Scheme:"https", Host:"github.com", Path:"/ogen-go/ogen"}`
	case Bool:
		return "true"
	default:
		panic(fmt.Sprintf("unexpected PrimitiveType: %d", p))
	}
}

func (t Type) FakeFields() (r []*Field) {
	obj := t.Validators.Object
	if !obj.MaxPropertiesSet {
		return t.Fields
	}

	required := 0
	for _, f := range t.Fields {
		// Count required fields
		if !f.Type.IsGeneric() {
			required++
			if required > obj.MaxProperties {
				panic(fmt.Sprintf(" required fields(%d) > maximumProperties(%d)", obj.MaxProperties, required))
			}
			r = append(r, f)
		}
	}
	for _, f := range t.Fields {
		// Count optional fields
		if f.Type.IsGeneric() {
			if len(r) >= obj.MaxProperties {
				break
			}
			r = append(r, f)
		}
	}
	return r
}
