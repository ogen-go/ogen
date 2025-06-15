package ir

import (
	"fmt"
)

func (t *Type) FakeValue() string {
	switch p := t.Primitive; p {
	case String:
		return `"string"`
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
		return fmt.Sprintf("%s(0)", p)
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
	case MAC:
		return `net.HardwareAddr{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}`
	case IP:
		if s := t.Schema; s != nil && s.Format == "ipv6" {
			return `netip.MustParseAddr("::1")`
		}
		return `netip.MustParseAddr("127.0.0.1")`
	case URL:
		return `url.URL{Scheme:"https", Host:"github.com", Path:"/ogen-go/ogen"}`
	case Bool:
		return "true"
	case Null:
		return "struct{}{}"
	default:
		panic("unexpected PrimitiveType: " + p)
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
		if f.Spec != nil && f.Spec.Required {
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
