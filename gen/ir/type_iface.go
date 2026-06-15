package ir

import (
	"fmt"
	"maps"
	"slices"
	"strings"

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

	t.InterfaceMethods[name] = name + "()"
}

func (t *Type) AddMethodSignature(name, signature string) {
	if !t.Is(KindInterface) {
		panic(unreachable(t))
	}

	t.InterfaceMethods[name] = name + signature
}

func (t *Type) DeclareMethod(method string) {
	if !t.CanHaveMethods() {
		panic(unreachable(t))
	}
	if t.DeclaredMethods == nil {
		t.DeclaredMethods = map[string]struct{}{}
	}
	t.DeclaredMethods[method] = struct{}{}
}

func (t *Type) HasDeclaredMethod(method string) bool {
	if t == nil {
		return false
	}
	_, ok := t.DeclaredMethods[method]
	return ok
}

func (t *Type) Methods() []string {
	ms := make(map[string]string)
	switch t.Kind {
	case KindInterface:
		ms = t.InterfaceMethods
	case KindStruct, KindMap, KindAlias, KindEnum, KindGeneric, KindSum, KindStream:
		for i := range t.Implements {
			maps.Copy(ms, i.InterfaceMethods)
		}
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}

	names := xmaps.SortedKeys(ms)
	result := make([]string, 0, len(names))
	for _, name := range names {
		result = append(result, ms[name])
	}
	return result
}

func (t *Type) ListImplementations() []*Type {
	if !t.Is(KindInterface) {
		panic(unreachable(t))
	}

	result := make([]*Type, 0, len(t.Implementations))
	for impl := range t.Implementations {
		result = append(result, impl)
	}
	slices.SortStableFunc(result, func(a, b *Type) int {
		return strings.Compare(a.Name, b.Name)
	})
	return result
}
