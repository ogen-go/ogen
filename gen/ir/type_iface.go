package ir

import (
	"fmt"
	"sort"

	"github.com/ogen-go/ogen/internal/xmaps"
)

func (t *Type) CanHaveMethods() bool {
	return !t.Is(KindPrimitive, KindArray, KindPointer, KindAny)
}

func (t *Type) Implement(i *Type) {
	if !t.CanHaveMethods() || !i.Is(KindInterface) {
		panic(unreachable(t))
	}

	if t.Implements == nil {
		t.Implements = map[*Type]struct{}{}
	}

	i.Implementations[t] = struct{}{}
	t.Implements[i] = struct{}{}
}

func (t *Type) Unimplement(i *Type) {
	if !t.CanHaveMethods() || !i.Is(KindInterface) {
		panic(unreachable(t))
	}

	delete(i.Implementations, t)
	delete(t.Implements, i)
}

func (t *Type) AddMethod(name string) {
	if !t.Is(KindInterface) {
		panic(unreachable(t))
	}

	t.InterfaceMethods[name] = struct{}{}
}

func (t *Type) Methods() []string {
	ms := make(map[string]struct{})
	switch t.Kind {
	case KindInterface:
		ms = t.InterfaceMethods
	case KindStruct, KindMap, KindAlias, KindEnum, KindGeneric, KindSum, KindStream:
		for i := range t.Implements {
			for m := range i.InterfaceMethods {
				ms[m] = struct{}{}
			}
		}
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}

	return xmaps.SortedKeys(ms)
}

func (t *Type) ListImplementations() []*Type {
	if !t.Is(KindInterface) {
		panic(unreachable(t))
	}

	result := make([]*Type, 0, len(t.Implementations))
	for impl := range t.Implementations {
		result = append(result, impl)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
