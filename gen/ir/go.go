package ir

import (
	"fmt"
)

// PrintGoValue prints given value as Go value.
func PrintGoValue(v interface{}) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%#v", v)
	}
}
