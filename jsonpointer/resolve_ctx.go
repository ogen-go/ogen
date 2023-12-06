package jsonpointer

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/location"
)

// RefKey is JSON Reference key.
type RefKey struct {
	// Loc is an URL of JSON document.
	Loc string
	// Ptr is JSON Pointer.
	Ptr string
}

// String returns string representation of reference.
func (r RefKey) String() string {
	return r.Loc + r.Ptr
}

// IsZero returns true if RefKey is zero.
func (r RefKey) IsZero() bool {
	var r0 struct {
		Loc string
		Ptr string
	}
	return r == r0
}

// FromURL sets RefKey from URL.
func (r *RefKey) FromURL(u *url.URL) {
	{
		// Make copy.
		u2 := *u
		u2.Fragment = ""
		r.Loc = u2.String()
	}
	r.Ptr = "#" + u.Fragment
}

type locstackItem struct {
	loc  *url.URL
	file location.File
}

// ResolveCtx is JSON Reference resolve context.
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

// DummyURL is dummy URL for testing purposes.
func DummyURL() *url.URL {
	return &url.URL{
		Scheme: "jsonschema",
		Host:   "dummy",
	}
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

func (r ResolveCtx) last() (last locstackItem, ok bool) {
	s := r.locstack
	if len(s) == 0 {
		return last, ok
	}
	return s[len(s)-1], true
}

// Key creates new reference key.
func (r *ResolveCtx) Key(ref string) (key RefKey, _ error) {
	parser := r.root.Parse
	if last, ok := r.last(); ok {
		parser = last.loc.Parse
	} else if strings.HasPrefix(ref, "#") {
		key.Ptr = ref
		key.Loc = r.root.String()
		return key, nil
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

	loc, err := url.Parse(key.Loc)
	if err != nil {
		return errors.Wrap(err, "invalid location")
	}
	r.locstack = append(r.locstack, locstackItem{
		loc:  loc,
		file: file,
	})
	return nil
}

// Delete removes reference from context.
func (r *ResolveCtx) Delete(key RefKey) {
	r.depthLimit++
	delete(r.refs, key)
	if len(r.locstack) > 0 {
		r.locstack = r.locstack[:len(r.locstack)-1]
	}
}

// IsRoot returns true if location stack is empty.
func (r *ResolveCtx) IsRoot(key RefKey) bool {
	resolved := errors.Must(r.root.Parse(key.Loc))
	return resolved.String() == r.root.String()
}

// File returns last file from stack.
func (r *ResolveCtx) File() (f location.File) {
	s := r.locstack
	if len(s) == 0 {
		return f
	}
	return s[len(s)-1].file
}
