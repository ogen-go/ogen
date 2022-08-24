package ir

import "fmt"

func unreachable(v any) string {
	return fmt.Sprintf("unreachable: %v", v)
}
