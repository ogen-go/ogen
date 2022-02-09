package gen

import "fmt"

func unreachable(v interface{}) string {
	return fmt.Sprintf("unreachable: %v", v)
}
