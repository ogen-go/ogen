package ir

import (
	"fmt"
	"reflect"

	"golang.org/x/exp/slices"
)

func (t *Type) RecursiveTo(target *Type) bool {
	return t.recursive(target, &walkpath{})
}

func (t *Type) recursive(target *Type, path *walkpath) bool {
	{
		// This is a list of types that cannot cause recursion.
		//
		// Primitive - has no fields.
		// Enum      - has no fields.
		// Any       - has no fields.
		// Pointer   - prevents recursion.
		// Array     - prevents recursion.
		// Map       - prevents recursion.
		whitelist := []Kind{KindPrimitive, KindEnum, KindAny, KindPointer, KindArray, KindMap}
		if t.Is(whitelist...) || target.Is(whitelist...) {
			return false
		}
	}

	if reflect.DeepEqual(t, target) {
		return true
	}

	if path.has(target) {
		return false
	}
	path = path.append(target)

	switch target.Kind {
	case KindAlias:
		return t.recursive(target.AliasTo, path)
	case KindGeneric:
		return t.recursive(target.GenericOf, path)
	case KindStruct:
		for _, f := range target.Fields {
			if f.Spec != nil && !f.Spec.Required {
				continue
			}
			if t.recursive(f.Type, path) {
				return true
			}
		}
		return false
	case KindSum:
		for _, of := range target.SumOf {
			if t.recursive(of, path) {
				return true
			}
		}
		return false
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}

type walkpath struct {
	nodes []*Type
}

func (wp *walkpath) has(t *Type) bool {
	return slices.Contains(wp.nodes, t)
}

func (wp *walkpath) append(t *Type) *walkpath {
	return &walkpath{
		append(
			wp.nodes[:len(wp.nodes):len(wp.nodes)],
			t,
		),
	}
}
