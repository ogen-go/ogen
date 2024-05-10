package gen

import (
	"github.com/ogen-go/ogen/gen/ir"
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

// longestPrefix founds the longest common prefix of k1 and k2.
func longestPrefix(k1, k2 string) int {
	smin := len(k1)
	if l := len(k2); l < smin {
		smin = l
	}

	for i := 0; i < smin; i++ {
		if k1[i] != k2[i] {
			return i
		}
	}
	return smin
}

func (t *RouteTree) addRoute(m Route) error {
	path := m.Path

	n := t.Root
	if n == nil {
		n = new(RouteNode)
		t.Root = n
	}
	for {
		if path == "" {
			return n.AddRoute(m)
		}
		// Head is a first character of route.
		head := path[0]
		// Find parameter index.
		_, _, pend, err := nextPathPart(path)
		if err != nil {
			return err
		}

		parent := n
		// Check for existing node with same head.
		n = n.getChild(head)
		if n == nil {
			// If there is no child with such head, create a new one.
			r, err := parent.addChild(path, m.Operation, &RouteNode{
				prefix: path,
				head:   head,
			})
			if err != nil {
				return err
			}
			return r.AddRoute(m)
		}

		// Skip common parameter node.
		//
		// This condition is met if child have parameter part in same place, e.g.
		//
		// /pet/{name}
		// /pet/{name}/friends
		//
		if n.param != nil {
			path = path[pend:]
			continue
		}

		// Found the longest common prefix of existing node and new.
		commonPrefix := longestPrefix(path, n.prefix)
		if commonPrefix == len(n.prefix) {
			// If common prefix fully matched by existing node, try to create child.
			path = path[commonPrefix:]
			continue
		}

		// Otherwise, we try to replace existing node.
		newChild := &RouteNode{
			head:   path[0],
			prefix: path[:commonPrefix],
		}
		parent.replaceChild(path[0], newChild)

		// Add existing node as child of replacer.
		n.head = n.prefix[commonPrefix]
		n.prefix = n.prefix[commonPrefix:]
		if _, err := newChild.addChild(n.prefix, m.Operation, n); err != nil {
			return err
		}

		// Special case: if new node has exactly same path, replace existing node.
		path = path[commonPrefix:]
		if path == "" {
			return newChild.AddRoute(m)
		}

		// Otherwise, we add new node as second child of replacer.
		r, err := newChild.addChild(path, m.Operation, &RouteNode{
			prefix: path,
			head:   path[0],
		})
		if err != nil {
			return err
		}
		return r.AddRoute(m)
	}
}

func (t *RouteTree) Walk(cb func(level int, n *RouteNode)) {
	t.Root.walk(0, cb)
}
