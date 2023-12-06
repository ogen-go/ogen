package jsonpointer

import "fmt"

// NotFoundError reports that requested value is not found.
type NotFoundError struct {
	Pointer string
}

// Error implements error.
func (n *NotFoundError) Error() string {
	return fmt.Sprintf("can't find value for %q", n.Pointer)
}
