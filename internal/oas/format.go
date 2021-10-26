package oas

// Format of Schema.
type Format string

// Possible formats.
const (
	FormatNone     Format = ""
	FormatUUID     Format = "uuid"
	FormatDate     Format = "date"
	FormatTime     Format = "time"
	FormatDateTime Format = "date-time"
	FormatDuration Format = "duration"
	FormatURI      Format = "uri"
	FormatIPv4     Format = "ipv4"
	FormatIPv6     Format = "ipv6"
	FormatByte     Format = "byte"
	FormatPassword Format = "password"
	FormatInt64    Format = "int64"
	FormatInt32    Format = "int32"
	FormatFloat    Format = "float"
	FormatDouble   Format = "double"

	// TODO(ernado): infer from OneOf(ipv4, ipv6) and remove

	FormatIP Format = "ip" // custom, non-standard
)
