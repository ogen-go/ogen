package jsonschema

import (
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/json"
)

// RootResolver is ReferenceResolver implementation.
type RootResolver struct {
	root       []byte
	parsedRoot *RawSchema
}

// NewRootResolver creates new RootResolver.
func NewRootResolver(root []byte) *RootResolver {
	return &RootResolver{root: root}
}

func findPath(ref string, buf []byte) ([]byte, error) {
	for _, part := range strings.Split(ref, "/") {
		found := false
		if err := jx.DecodeBytes(buf).ObjBytes(func(d *jx.Decoder, key []byte) error {
			switch string(key) {
			case part:
				found = true
				raw, err := d.RawAppend(nil)
				if err != nil {
					return errors.Wrapf(err, "parse %q", key)
				}
				buf = raw
			default:
				return d.Skip()
			}
			return nil
		}); err != nil {
			return nil, err
		}

		if !found {
			return nil, errors.Errorf("find %q", part)
		}
	}
	return buf, nil
}

func splitURI(u string) (base, fragment string) {
	hash := strings.IndexByte(u, '#')
	if hash == -1 {
		return u, "#"
	}
	f := u[hash:]
	if f == "#/" {
		f = "#"
	}
	return u[0:hash], f
}

// ResolveReference implements ReferenceResolver.
func (r RootResolver) ResolveReference(ref string) (rawSchema *RawSchema, err error) {
	ref = strings.TrimSpace(ref)
	base, fragment := splitURI(ref)

	buf := r.root
	if base != "" {
		return nil, errors.Errorf("external base %q is not supported", base)
	}

	if fragment == "#" {
		if r.parsedRoot != nil {
			return r.parsedRoot, nil
		}
		defer func() {
			if err == nil {
				r.parsedRoot = rawSchema
			}
		}()
	} else {
		fragment = strings.TrimPrefix(fragment, "#/")
		buf, err = findPath(fragment, buf)
		if err != nil {
			return nil, errors.Wrapf(err, "find %q", ref)
		}
	}

	if err := json.Unmarshal(buf, &rawSchema); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return rawSchema, nil
}
