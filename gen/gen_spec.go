package gen

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"sync/atomic"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/openapi/parser"
)

// EmbeddedSpec is the OpenAPI document embedded into generated code by the "server/spec" feature.
type EmbeddedSpec struct {
	// JSON is the document encoded as JSON.
	JSON string
	// YAML is the document encoded as YAML.
	YAML string
	// JSONETag is a quoted strong ETag of JSON.
	JSONETag string
	// YAMLETag is a quoted strong ETag of YAML.
	YAMLETag string
	// Title is the API title from the info object.
	Title string
	// Expanded is true if the document was produced by [parser.Expand], i.e.
	// external references were resolved.
	Expanded bool
}

// countingResolver wraps an [jsonschema.ExternalResolver] and records whether it was ever used.
type countingResolver struct {
	jsonschema.ExternalResolver
	used atomic.Bool
}

// Get implements [jsonschema.ExternalResolver].
func (r *countingResolver) Get(ctx context.Context, loc string) ([]byte, error) {
	r.used.Store(true)
	return r.ExternalResolver.Get(ctx, loc)
}

// Used reports whether the resolver was called at least once.
func (r *countingResolver) Used() bool {
	return r.used.Load()
}

// marshalNode converts any value to a YAML node by round-tripping through YAML.
//
// It is used instead of json.Marshal because specification extensions are inlined
// in YAML (`yaml:",inline"`) but excluded from JSON (`json:"-"`).
func marshalNode(v any) (*yaml.Node, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}
	var n yaml.Node
	if err := yaml.Unmarshal(data, &n); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return &n, nil
}

// specDocument returns the YAML node of the document to embed.
func specDocument(spec *ogen.Spec, api *openapi.API, source SpecSource, usedExternal bool) (_ *yaml.Node, expanded bool, _ error) {
	expand := func() (*yaml.Node, bool, error) {
		expanded, err := parser.Expand(api)
		if err != nil {
			return nil, false, errors.Wrap(err, "expand")
		}
		n, err := marshalNode(expanded)
		return n, true, err
	}
	raw := func() (*yaml.Node, bool, error) {
		if spec.Raw != nil {
			return spec.Raw, false, nil
		}
		// Spec was built programmatically (e.g. via DSL), there is no original document.
		n, err := marshalNode(spec)
		return n, false, err
	}

	switch source {
	case SpecSourceExpanded:
		return expand()
	case SpecSourceRaw:
		return raw()
	case "", SpecSourceAuto:
		if usedExternal {
			return expand()
		}
		return raw()
	default:
		return nil, false, errors.Errorf("unknown spec source %q", source)
	}
}

// stringKeys returns a deep copy of the node tree where every mapping key is tagged as a string.
//
// YAML allows non-string keys (e.g. `200:` in responses), but JSON does not, and
// [yaml.Node.EncodeJSON] refuses to convert them.
func stringKeys(n *yaml.Node) *yaml.Node {
	copies := map[*yaml.Node]*yaml.Node{}

	var clone func(n *yaml.Node, key bool) *yaml.Node
	clone = func(n *yaml.Node, key bool) *yaml.Node {
		if n == nil {
			return nil
		}
		if c, ok := copies[n]; ok {
			return c
		}

		c := new(yaml.Node)
		*c = *n
		copies[n] = c

		if key && c.Kind == yaml.ScalarNode {
			c.Tag = "!!str"
		}
		if n.Alias != nil {
			c.Alias = clone(n.Alias, key)
		}
		if len(n.Content) > 0 {
			c.Content = make([]*yaml.Node, len(n.Content))
			for i, child := range n.Content {
				c.Content[i] = clone(child, n.Kind == yaml.MappingNode && i%2 == 0)
			}
		}
		return c
	}
	return clone(n, false)
}

func etag(data []byte) string {
	sum := sha256.Sum256(data)
	return strconv.Quote(hex.EncodeToString(sum[:16]))
}

// embedSpec builds [EmbeddedSpec] for given spec.
func embedSpec(spec *ogen.Spec, api *openapi.API, source SpecSource, usedExternal bool) (*EmbeddedSpec, error) {
	node, expanded, err := specDocument(spec, api, source, usedExternal)
	if err != nil {
		return nil, err
	}

	yamlData, err := yaml.Marshal(node)
	if err != nil {
		return nil, errors.Wrap(err, "encode yaml")
	}

	var e jx.Encoder
	e.SetIdent(2)
	if err := stringKeys(node).EncodeJSON(&e); err != nil {
		return nil, errors.Wrap(err, "encode json")
	}
	jsonData := e.Bytes()

	return &EmbeddedSpec{
		JSON:     string(jsonData),
		YAML:     string(yamlData),
		JSONETag: etag(jsonData),
		YAMLETag: etag(yamlData),
		Title:    api.Info.Title,
		Expanded: expanded,
	}, nil
}
