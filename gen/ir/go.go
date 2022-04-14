package ir

import "fmt"

// PrintGoValue prints given value as Go value.
func PrintGoValue(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return ""
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%#v", v)
	}
}
