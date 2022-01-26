package ir

import (
	"fmt"
	"reflect"
)

func (t *Type) RecursiveTo(target *Type) bool {
	return t.recursive(target, &walkpath{})
}

func (t *Type) recursive(target *Type, path *walkpath) bool {
	if t.Is(KindPrimitive, KindPointer, KindMap, KindArray, KindEnum) ||
		target.Is(KindPrimitive, KindPointer, KindMap, KindArray, KindEnum) {
		return false
	}

	if reflect.DeepEqual(t, target) {
		return true
	}

	if path.has(t) {
		return false
	}

	path = path.append(t)

	switch target.Kind {
	case KindAlias:
		return t.recursive(target.AliasTo, path)
	case KindGeneric:
		return t.recursive(target.GenericOf, path)
	case KindStruct:
		for _, f := range target.Fields {
			if !f.Spec.Required {
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
	for _, n := range wp.nodes {
		if n == t {
			return true
		}
	}
	return false
}

func (wp *walkpath) append(t *Type) *walkpath {
	return &walkpath{
		append(
			wp.nodes[:len(wp.nodes):len(wp.nodes)],
			t,
		),
	}
}
