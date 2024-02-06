package ir

import "fmt"

type PrimitiveType int

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
	case Bool:
		return "bool"
	case Null:
		return "struct{}"
	case Time:
		return "time.Time"
	case Duration:
		return "time.Duration"
	case UUID:
		return "uuid.UUID"
	case MAC:
		return "net.HardwareAddr"
	case IP:
		return "netip.Addr"
	case URL:
		return "url.URL"
	case File:
		return "ht.MultipartFile"
	case Custom:
		return "<custom>"
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
	Bool
	Null
	Time
	Duration
	UUID
	MAC
	IP
	URL
	File
	Custom
)
