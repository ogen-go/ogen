// Package jsonpointer contains RFC 6901 JSON Pointer implementation.
package jsonpointer

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
)

// Resolve takes given pointer and returns byte slice of requested value if any.
// If value not found, returns NotFoundError.
func Resolve(ptr string, node *yaml.Node) (*yaml.Node, error) {
	if node == nil {
		return nil, errors.New("root is nil")
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	switch {
	case ptr == "" || ptr == "#":
		return node, nil
	case ptr[0] == '/':
		return find(ptr, node)
	case ptr[0] == '#': // Note that length is bigger than 1.
		unescaped, err := url.PathUnescape(ptr[1:])
		if err != nil {
			return nil, errors.Wrap(err, "unescape")
		}
		// Fast-path to not parse URL.
		return find(unescaped, node)
	}

	u, err := url.Parse(ptr)
	if err != nil {
		return nil, err
	}
	return find(u.Fragment, node)
}

func find(ptr string, node *yaml.Node) (*yaml.Node, error) {
	if ptr == "" {
		return node, nil
	}

	if ptr[0] != '/' {
		return nil, errors.Errorf("invalid pointer %q: pointer must start with '/'", ptr)
	}
	// Cut first /.
	ptr = ptr[1:]

	err := splitFunc(ptr, '/', func(part string) (err error) {
		part = unescape(part)
		var (
			result *yaml.Node
			ok     bool
		)
		switch tt := node.Kind; tt {
		case yaml.MappingNode:
			result, ok = findKey(node, part)
		case yaml.SequenceNode:
			result, ok, err = findIdx(node, part)
			if err != nil {
				return errors.Wrapf(err, "find index %q", part)
			}
		default:
			return errors.Errorf("unexpected type %q", node.ShortTag())
		}
		if !ok {
			return &NotFoundError{Pointer: ptr}
		}

		node = result
		return err
	})
	return node, err
}

func findIdx(n *yaml.Node, part string) (result *yaml.Node, ok bool, _ error) {
	index, err := strconv.ParseUint(part, 10, 64)
	if err != nil {
		return nil, false, errors.Wrap(err, "index")
	}

	children := n.Content
	if index >= uint64(len(children)) {
		return nil, false, nil
	}
	return children[index], true, nil
}

func findKey(n *yaml.Node, part string) (*yaml.Node, bool) {
	children := n.Content
	for i := 0; i < len(children); i += 2 {
		key, value := children[i], children[i+1]
		if key.Value == part {
			return value, true
		}
	}
	return nil, false
}

var unescapeReplacer = strings.NewReplacer(
	"~1", "/",
	"~0", "~",
)

func unescape(part string) string {
	// Replacer always creates new string, check that unescape is really necessary.
	if !strings.Contains(part, "~1") && !strings.Contains(part, "~0") {
		return part
	}
	return unescapeReplacer.Replace(part)
}
