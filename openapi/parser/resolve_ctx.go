package parser

import (
	"net/url"

	"github.com/go-faster/errors"
)

type refKey struct {
	loc string
	ref string
}

func (r *refKey) fromURL(u *url.URL) {
	{
		// Make copy.
		u2 := *u
		u2.Fragment = ""
		r.loc = u2.String()
	}
	r.ref = "#" + u.Fragment
}

type resolveCtx struct {
	// Location stack. Used for context-depending resolving.
	//
	// For resolve trace like
	//
	// 	"#/components/schemas/Schema" ->
	// 	"https://example.com/schema#Schema" ->
	//	"#/definitions/SchemaProperty"
	//
	// "#/definitions/SchemaProperty" should be resolved against "https://example.com/schema".
	locstack []string
	// Store references to detect infinite recursive references.
	refs       map[refKey]struct{}
	depthLimit int
}

func newResolveCtx(depthLimit int) *resolveCtx {
	return &resolveCtx{
		locstack:   nil,
		refs:       map[refKey]struct{}{},
		depthLimit: depthLimit,
	}
}

func (r *resolveCtx) add(ref string) (key refKey, _ error) {
	u, err := url.Parse(ref)
	if err != nil {
		return refKey{}, err
	}
	key.fromURL(u)

	if r.depthLimit <= 0 {
		return refKey{}, errors.New("depth limit exceeded")
	}
	if _, ok := r.refs[key]; ok {
		return key, errors.New("infinite recursion")
	}
	r.refs[key] = struct{}{}
	r.depthLimit--

	if key.loc != "" {
		r.locstack = append(r.locstack, key.loc)
	}
	return key, nil
}

func (r *resolveCtx) delete(key refKey) {
	r.depthLimit++
	delete(r.refs, key)
	if key.loc != "" && len(r.locstack) > 0 {
		r.locstack = r.locstack[:len(r.locstack)-1]
	}
}

func (r *resolveCtx) lastLoc() string {
	s := r.locstack
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}
