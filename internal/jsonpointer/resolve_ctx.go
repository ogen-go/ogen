package jsonpointer

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/location"
)

// RefKey is JSON reference key.
type RefKey struct {
	Loc string
	Ref string
}

// FromURL sets RefKey from URL.
func (r *RefKey) FromURL(u *url.URL) {
	{
		// Make copy.
		u2 := *u
		u2.Fragment = ""
		r.Loc = u2.String()
	}
	r.Ref = "#" + u.Fragment
}

type locstackItem struct {
	loc  string
	file location.File
}

// ResolveCtx is JSON pointer resolve context.
type ResolveCtx struct {
	// Location stack. Used for context-depending resolving.
	//
	// For resolve trace like
	//
	// 	"#/components/schemas/Schema" ->
	// 	"https://example.com/schema#Schema" ->
	//	"#/definitions/SchemaProperty"
	//
	// "#/definitions/SchemaProperty" should be resolved against "https://example.com/schema".
	locstack []locstackItem
	// root is root location.
	root *url.URL
	// Store references to detect infinite recursive references.
	refs       map[RefKey]struct{}
	depthLimit int
}

// DefaultDepthLimit is default depth limit for ResolveCtx.
const DefaultDepthLimit = 1000

// DefaultCtx creates new ResolveCtx with default depth limit.
func DefaultCtx() *ResolveCtx {
	return NewResolveCtx(nil, DefaultDepthLimit)
}

// NewResolveCtx creates new ResolveCtx.
func NewResolveCtx(root *url.URL, depthLimit int) *ResolveCtx {
	return &ResolveCtx{
		locstack:   nil,
		root:       root,
		refs:       map[RefKey]struct{}{},
		depthLimit: depthLimit,
	}
}

// Key creates new reference key.
func (r *ResolveCtx) Key(ref string) (key RefKey, _ error) {
	parser := url.Parse
	if r.root != nil {
		parser = r.root.Parse
	}
	if s := r.locstack; len(s) > 0 {
		base, err := url.Parse(s[len(s)-1].loc)
		if err != nil {
			return key, err
		}
		parser = func(rawURL string) (*url.URL, error) {
			u, err := base.Parse(rawURL)
			if err != nil {
				return nil, err
			}
			return u, nil
		}
	} else if strings.HasPrefix(ref, "#") {
		return RefKey{
			Ref: ref,
		}, nil
	}

	u, err := parser(ref)
	if err != nil {
		return RefKey{}, err
	}
	key.FromURL(u)
	return key, nil
}

// AddKey adds reference key to context.
func (r *ResolveCtx) AddKey(key RefKey, file location.File) error {
	if r.depthLimit <= 0 {
		return errors.New("depth limit exceeded")
	}
	if _, ok := r.refs[key]; ok {
		return errors.New("infinite recursion")
	}
	r.refs[key] = struct{}{}
	r.depthLimit--

	if loc := key.Loc; loc != "" {
		r.locstack = append(r.locstack, locstackItem{
			loc:  loc,
			file: file,
		})
	}
	return nil
}

// Delete removes reference from context.
func (r *ResolveCtx) Delete(key RefKey) {
	r.depthLimit++
	delete(r.refs, key)
	if key.Loc != "" && len(r.locstack) > 0 {
		r.locstack = r.locstack[:len(r.locstack)-1]
	}
}

// IsRoot returns true if location stack is empty.
func (r *ResolveCtx) IsRoot() bool {
	return len(r.locstack) == 0
}

// File returns last file from stack.
func (r *ResolveCtx) File() (f location.File) {
	s := r.locstack
	if len(s) == 0 {
		return f
	}
	return s[len(s)-1].file
}
