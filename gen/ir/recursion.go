package ir

import (
	"fmt"
	"reflect"
	"slices"
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

	if t == target || reflect.DeepEqual(t, target) {
		return true
	}
	if path.has(target) {
		return true
	}
	path.add(target)
	defer path.delete(target)

	switch target.Kind {
	case KindAlias:
		return t.recursive(target.AliasTo, path)
	case KindGeneric:
		return t.recursive(target.GenericOf, path)
	case KindStruct:
		return slices.ContainsFunc(target.Fields, func(f *Field) (r bool) {
			return t.recursive(f.Type, path)
		})
	case KindSum:
		return slices.ContainsFunc(target.SumOf, func(of *Type) bool {
			return t.recursive(of, path)
		})
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}

type walkpath struct {
	nodes map[*Type]struct{}
}

func (wp *walkpath) has(t *Type) bool {
	_, ok := wp.nodes[t]
	return ok
}

func (wp *walkpath) add(t *Type) {
	if wp.nodes == nil {
		wp.nodes = map[*Type]struct{}{}
	}
	wp.nodes[t] = struct{}{}
}

func (wp *walkpath) delete(t *Type) {
	delete(wp.nodes, t)
}
