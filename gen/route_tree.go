package gen

import (
	"unicode/utf8"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
)

// RouteTree is Radix tree of routes.
type RouteTree struct {
	Root *RouteNode
}

func findParam(op *ir.Operation, name string) (*ir.Parameter, bool) {
	if op == nil {
		return nil, false
	}
	for _, p := range op.Params {
		if p.Spec != nil && p.Spec.In.Path() && p.Spec.Name == name {
			return p, true
		}
	}
	return nil, false
}

func longestPrefix(k1, k2 string) int {
	min := len(k1)
	if l := len(k2); l < min {
		min = l
	}

	for i := 0; i < min; i++ {
		if k1[i] != k2[i] {
			return i
		}
	}
	return min
}

func (t *RouteTree) addRoute(path string, operation *ir.Operation) error {
	if !utf8.ValidString(path) {
		return errors.Errorf("invalid path: path must be valid UTF-8 string")
	}

	n := t.Root
	if n == nil {
		n = new(RouteNode)
		t.Root = n
	}
	for {
		if len(path) == 0 {
			n.op = operation
			return nil
		}
		head := path[0]
		_, _, pend, _, err := nextPathPart(path)
		if err != nil {
			return err
		}

		parent := n
		n = n.getChild(head)
		if n == nil {
			r, err := parent.addChild(path, operation, &RouteNode{
				prefix: path,
				head:   head,
			})
			if err != nil {
				return err
			}
			r.op = operation
			return nil
		}

		if n.param != nil {
			path = path[pend:]
			continue
		}

		commonPrefix := longestPrefix(path, n.prefix)
		if commonPrefix == len(n.prefix) {
			path = path[commonPrefix:]
			continue
		}

		newChild := &RouteNode{
			head:   path[0],
			prefix: path[:commonPrefix],
			op:     operation,
		}
		parent.replaceChild(path[0], newChild)

		n.head = n.prefix[commonPrefix]
		n.prefix = n.prefix[commonPrefix:]
		if _, err := newChild.addChild(n.prefix, operation, n); err != nil {
			return err
		}

		path = path[commonPrefix:]
		if len(path) == 0 {
			newChild.op = operation
			return nil
		}

		r, err := newChild.addChild(path, operation, &RouteNode{
			prefix: path,
			head:   path[0],
		})
		if err != nil {
			return err
		}
		r.op = operation
		return nil
	}
}

func (t *RouteTree) Walk(cb func(level int, n *RouteNode)) {
	t.Root.walk(0, cb)
}
