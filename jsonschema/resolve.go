package jsonschema

import (
	"context"
	"net/url"

	"github.com/go-faster/errors"
)

// ReferenceResolver resolves JSON schema references.
type ReferenceResolver interface {
	ResolveReference(ref string) (*RawSchema, error)
}

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
		refs:       map[refKey]struct{}{},
		depthLimit: depthLimit,
	}
}

func (r *resolveCtx) add(key refKey) error {
	if r.depthLimit <= 0 {
		return errors.New("reference depth limit exceeded")
	}
	if _, ok := r.refs[key]; ok {
		return errors.New("infinite recursion")
	}
	r.refs[key] = struct{}{}
	r.depthLimit--

	if key.loc != "" {
		r.locstack = append(r.locstack, key.loc)
	}
	return nil
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

func (p *Parser) getResolver(ctx *resolveCtx) (ReferenceResolver, error) {
	loc := ctx.lastLoc()

	r, ok := p.schemas[loc]
	if ok {
		return r, nil
	}

	root, err := p.external.Get(context.TODO(), loc)
	if err != nil {
		return nil, errors.Wrapf(err, "external %q", loc)
	}
	r = NewRootResolver(root)
	p.schemas[loc] = r

	return r, nil
}

func (p *Parser) resolve(ref string, ctx *resolveCtx) (*Schema, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	var key refKey
	key.fromURL(u)

	if s, ok := p.refcache[key]; ok {
		return s, nil
	}

	if err := ctx.add(key); err != nil {
		return nil, err
	}
	defer func() {
		// Drop the resolved ref to prevent false-positive infinite recursion detection.
		ctx.delete(key)
	}()

	resolver, err := p.getResolver(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := resolver.ResolveReference(key.ref)
	if err != nil {
		return nil, errors.Wrap(err, "find schema")
	}

	return p.parse1(raw, ctx, func(s *Schema) *Schema {
		s.Ref = ref
		p.refcache[key] = s
		return p.extendInfo(raw, s)
	})
}
