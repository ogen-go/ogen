package ir

// JSON returns json specification for t.
func (t *Type) JSON() JSON {
	return JSON{
		Type: t,
	}
}

// JSON specifies json encoding and decoding for Type.
type JSON struct {
	Type *Type
}

// Format returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of string formats, like "json.WriteUUID",
// where UUID is Format.
func (j JSON) Format() string {
	if j.Type.Schema == nil {
		return ""
	}
	switch j.Type.Schema.Format {
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
