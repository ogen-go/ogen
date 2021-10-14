package ast

// Type is an abstraction for Interface, Schema and Pointer.
type Type interface {
	Type() string
}
