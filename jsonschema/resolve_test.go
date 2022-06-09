package jsonschema

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

type external map[string]components

func (e external) Get(_ context.Context, loc string) (ReferenceResolver, error) {
	r, ok := e[loc]
	if !ok {
		return nil, errors.Errorf("unexpected location %q", loc)
	}
	return r, nil
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
				Type: "object",
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
					{
						Name: "remote_recursive",
						Schema: &RawSchema{
							Ref: "https://example.com/bar.json#/components/schemas/Recursive",
						},
					},
				},
			},
			"Property": {
				Type: "number",
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
				Type: "boolean",
			},
			"Recursive": {
				Ref: "foo.json#/components/schemas/Property",
			},
		},
	}

	parser := NewParser(Settings{
		External: remote,
		Resolver: root,
	})

	out, err := parser.Parse(&RawSchema{
		Type: "array",
		Items: &RawSchema{
			Ref: "#/components/schemas/LocalSchema",
		},
	})
	require.NoError(t, err)

	expect := &Schema{
		Type: Array,
		Item: &Schema{
			Ref:  "foo.json#/components/schemas/RemoteSchema",
			Type: Object,
			Properties: []Property{
				{
					Name:   "relative",
					Schema: &Schema{Ref: "#/components/schemas/Property", Type: Number},
				},
				{
					Name:   "absolute",
					Schema: &Schema{Ref: "foo.json#/components/schemas/Property", Type: Number},
				},
				{
					Name:   "remote_absolute",
					Schema: &Schema{Ref: "https://example.com/bar.json#/components/schemas/Property", Type: Boolean},
				},
				{
					Name:   "remote_recursive",
					Schema: &Schema{Ref: "foo.json#/components/schemas/Property", Type: Number},
				},
			},
		},
	}
	require.Equal(t, expect, out)
}
