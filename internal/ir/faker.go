package ir

import "fmt"

func (p PrimitiveType) FakeValue() string {
	switch p {
	case String:
		return `"string"`
	case ByteSlice:
		return `[]byte("[]byte")`
	case Int:
		return "int(0)"
	case Int8:
		return "int8(0)"
	case Int16:
		return "int16(0)"
	case Int32:
		return "int32(0)"
	case Int64:
		return "int64(0)"
	case Uint:
		return "uint(0)"
	case Uint8:
		return "uint8(0)"
	case Uint16:
		return "uint16(0)"
	case Uint32:
		return "uint32(0)"
	case Uint64:
		return "uint64(0)"
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
