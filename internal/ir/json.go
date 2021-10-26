package ir

// JSON returns json encoding/decoding rules for t.
func (t *Type) JSON() JSON {
	return JSON{
		t: t,
	}
}

// JSON specifies json encoding and decoding for Type.
type JSON struct {
	t *Type
}

// Format returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of string formats, like "json.WriteUUID",
// where UUID is Format.
func (j JSON) Format() string {
	if j.t.Schema == nil {
		return ""
	}
	switch j.t.Schema.Format {
	case "uuid":
		return "UUID"
	case "date":
		return "Date"
	case "time":
		return "Time"
	case "date-time":
		return "DateTime"
	case "duration":
		return "Duration"
	case "ip", "ipv4", "ipv6":
		return "IP"
	case "uri":
		return "URI"
	default:
		return ""
	}
}

// Type returns json value type that can represent Type.
//
// E.g. string primitive can be represented by StringValue.
// Blank string is returned if there is no appropriate json type.
func (j JSON) Type() string {
	if j.t.IsNumeric() {
		return "NumberValue"
	}
	if j.t.Is(KindArray) {
		return "ArrayValue"
	}
	if j.t.Is(KindStruct) {
		return "ObjectValue"
	}
	switch j.t.Primitive {
	case Bool:
		return "BoolValue"
	case String, Time, Duration, UUID, IP, URL:
		return "StringValue"
	default:
		return ""
	}
}

// raw denotes whether Type can be encoded or decoded using simple
// json method, e.g. j.WriteString.
//
// Mostly true for primitives or enums.
func (j JSON) raw() bool {
	if !j.t.Is(KindPrimitive, KindEnum) {
		return false
	}

	if j.t.IsNumeric() {
		return true
	}
	switch j.t.Primitive {
	case Bool, String:
		return true
	default:
		return false
	}
}

// f is name of json method for decoding and encoding to use.
//
// For example. if Type can be encoded via j.WriteString, the "String" value
// is returned.
//
// Blank string is returned otherwise.
func (j JSON) f() string {
	if !j.raw() {
		return ""
	}
	return capitalize(j.t.Primitive.String())
}

// JSONWrite returns function name from json package that writes value.
func (j JSON) Write() string {
	if j.f() == "" {
		return ""
	}
	return "Write" + j.f()
}

// JSONRead returns function name from json package that reads value.
func (j JSON) Read() string {
	if j.f() == "" {
		return ""
	}
	return "Read" + j.f()
}
