package jsonschema

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
)

type external map[string]components

func (e external) Get(_ context.Context, loc string) ([]byte, error) {
	loc = strings.TrimPrefix(loc, "/")
	r, ok := e[loc]
	if !ok {
		return nil, errors.Errorf("unexpected location %q", loc)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}

	enc := jx.GetEncoder()
	enc.Obj(func(e *jx.Encoder) {
		enc.FieldStart("components")
		enc.Obj(func(e *jx.Encoder) {
			enc.FieldStart("schemas")
			enc.Raw(data)
		})
	})

	return enc.Bytes(), nil
}

func TestExternalReference(t *testing.T) {
	root := components{
		"LocalSchema": {
			Ref: "foo.json#/components/schemas/RemoteSchema",
		},
	}
	remote := external{
		"foo.json": components{
			"RemoteSchema": {
				Type: StringArray{"object"},
				Properties: RawProperties{
					{
						Name: "relative",
						Schema: &RawSchema{
							Ref: "#/components/schemas/Property",
						},
					},
					{
						Name: "absolute",
						Schema: &RawSchema{
							Ref: "foo.json#/components/schemas/Property",
						},
					},
					{
						Name: "remote_absolute",
						Schema: &RawSchema{
							Ref: "https://example.com/bar.json#/components/schemas/Property",
						},
					},
				},
			},
			"Property": {
				Type: StringArray{"number"},
			},
		},
		"https://example.com/bar.json": components{
			"SecondaryRemoteSchema": {
				Ref: "#/components/schemas/Alias",
			},
			"Alias": {
				Ref: "https://example.com/bar.json#/components/schemas/Property",
			},
			"Property": {
				Type: StringArray{"boolean"},
			},
		},
	}

	parser := NewParser(Settings{
		External: remote,
		Resolver: root,
	})

	out, err := parser.Parse(&RawSchema{
		Type: StringArray{"array"},
		Items: &RawItems{
			Item: &RawSchema{
				Ref: "#/components/schemas/LocalSchema",
			},
		},
	}, testCtx())
	require.NoError(t, err)

	expect := &Schema{
		Type: Array,
		Item: &Schema{
			Ref:  Ref{Loc: "/foo.json", Ptr: "#/components/schemas/RemoteSchema"},
			Type: Object,
			Properties: []Property{
				{
					Name:   "relative",
					Schema: &Schema{Ref: Ref{Loc: "/foo.json", Ptr: "#/components/schemas/Property"}, Type: Number},
				},
				{
					Name:   "absolute",
					Schema: &Schema{Ref: Ref{Loc: "/foo.json", Ptr: "#/components/schemas/Property"}, Type: Number},
				},
				{
					Name:   "remote_absolute",
					Schema: &Schema{Ref: Ref{Loc: "https://example.com/bar.json", Ptr: "#/components/schemas/Property"}, Type: Boolean},
				},
			},
		},
	}
	zeroLocator(out)
	require.Equal(t, expect, out)
}

func zeroLocator(s *Schema) {
	var zeroed location.Pointer
	if s == nil {
		return
	}
	s.Pointer = zeroed

	zeroLocator(s.Item)
	for _, p := range s.Properties {
		zeroLocator(p.Schema)
	}
	zeroMany := func(many []*Schema) {
		for _, s := range many {
			zeroLocator(s)
		}
	}
	zeroMany(s.AllOf)
	zeroMany(s.OneOf)
	zeroMany(s.AnyOf)
}

func TestLimitDepth(t *testing.T) {
	root := components{
		"Schema1": {
			Ref: "#/components/schemas/Schema2",
		},
		"Schema2": {
			Ref: "#/components/schemas/Schema3",
		},
		"Schema3": {
			Ref: "#/components/schemas/Schema4",
		},
		"Schema4": {
			Type: StringArray{"string"},
		},
	}

	tests := []struct {
		limit   int
		checker func(t require.TestingT, err error, args ...any)
	}{
		{1, require.Error},
		{2, require.Error},
		{3, require.Error},
		{4, require.NoError},
	}

	for _, tt := range tests {
		parser := NewParser(Settings{
			Resolver: root,
		})
		ctx := jsonpointer.NewResolveCtx(&url.URL{Path: "/limit.json"}, tt.limit)
		_, err := parser.Resolve("#/components/schemas/Schema1", ctx)
		tt.checker(t, err, "limit: %d", tt.limit)
	}
}
