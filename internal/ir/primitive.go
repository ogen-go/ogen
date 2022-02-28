package ir

import "fmt"

type PrimitiveType int

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

func (p PrimitiveType) String() string {
	switch p {
	case String:
		return "string"
	case ByteSlice:
		return "[]byte"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Time:
		return "time.Time"
	case Duration:
		return "time.Duration"
	case UUID:
		return "uuid.UUID"
	case IP:
		return "net.IP"
	case URL:
		return "url.URL"
	case Bool:
		return "bool"
	default:
		panic(fmt.Sprintf("unexpected PrimitiveType: %d", p))
	}
}

// IsString whether this type is string.
func (p PrimitiveType) IsString() bool {
	return p == String
}

const (
	None PrimitiveType = iota
	String
	ByteSlice
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	Time
	Duration
	UUID
	IP
	URL
	Bool
)
