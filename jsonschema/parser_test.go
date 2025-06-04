package jsonschema

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
)

type components map[string]*RawSchema

func (c components) ResolveReference(ref string) (*RawSchema, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid schema reference %q", ref)
	}

	name := strings.TrimPrefix(ref, prefix)
	s, ok := c[name]
	if !ok {
		return nil, errors.New("schema not found")
	}

	return s, nil
}

func testCtx() *jsonpointer.ResolveCtx {
	return jsonpointer.NewResolveCtx(&url.URL{Path: "/root.json"}, jsonpointer.DefaultDepthLimit)
}

func TestSchemaSimple(t *testing.T) {
	parser := NewParser(Settings{})

	out, err := parser.Parse(&RawSchema{
		Type: "object",
		Properties: []RawProperty{
			{
				Name:   "id",
				Schema: &RawSchema{Type: "integer"},
			},
			{
				Name:   "name",
				Schema: &RawSchema{Type: "string"},
			},
		},
		Required: []string{"id", "name"},
	}, testCtx())
	require.NoError(t, err)

	expect := &Schema{
		Type:     Object,
		Required: []string{"id", "name"},
		Properties: []Property{
			{
				Name:     "id",
				Schema:   &Schema{Type: Integer},
				Required: true,
			},
			{
				Name:     "name",
				Schema:   &Schema{Type: String},
				Required: true,
			},
		},
	}

	require.Equal(t, expect, out)
}

func TestSchemaRecursive(t *testing.T) {
	components := components{
		"Pet": {
			Type: "object",
			Properties: []RawProperty{
				{
					Name:   "id",
					Schema: &RawSchema{Type: "integer"},
				},
				{
					Name:   "name",
					Schema: &RawSchema{Type: "string"},
				},
				{
					Name: "friends",
					Schema: &RawSchema{
						Type: "array",
						Items: &RawItems{
							Item: &RawSchema{
								Ref: "#/components/schemas/Pet",
							},
						},
					},
				},
			},
			Required: []string{"id", "name", "friends"},
		},
	}

	pet := &Schema{
		Type:     Object,
		Required: []string{"id", "name", "friends"},
		Ref:      Ref{Loc: "/root.json", Ptr: "#/components/schemas/Pet"},
	}
	pet.Properties = []Property{
		{
			Name:     "id",
			Schema:   &Schema{Type: Integer},
			Required: true,
		},
		{
			Name:     "name",
			Schema:   &Schema{Type: String},
			Required: true,
		},
		{
			Name: "friends",
			Schema: &Schema{
				Type: Array,
				Item: pet,
			},
			Required: true,
		},
	}

	expectRefcache := map[jsonpointer.RefKey]*Schema{
		{Loc: "/root.json", Ptr: "#/components/schemas/Pet"}: {
			Type:     Object,
			Ref:      Ref{Loc: "/root.json", Ptr: "#/components/schemas/Pet"},
			Required: []string{"id", "name", "friends"},
			Properties: []Property{
				{
					Name:     "id",
					Schema:   &Schema{Type: Integer},
					Required: true,
				},
				{
					Name:     "name",
					Schema:   &Schema{Type: String},
					Required: true,
				},
				{
					Name: "friends",
					Schema: &Schema{
						Type: Array,
						Item: pet,
					},
					Required: true,
				},
			},
		},
	}

	parser := NewParser(Settings{
		Resolver: components,
	})

	out, err := parser.Parse(&RawSchema{
		Ref: "#/components/schemas/Pet",
	}, testCtx())
	require.NoError(t, err)
	require.Equal(t, expectRefcache, parser.refcache)
	require.Equal(t, pet, out)
}

func TestSchemaInfiniteRecursion(t *testing.T) {
	testCases := []RawSchema{
		{
			Type: "object",
			Ref:  "#/components/schemas/Type",
		},
	}

	for _, cse := range testCases {
		components := components{
			"Type": &cse,
		}
		parser := NewParser(Settings{
			Resolver: components,
		})
		_, err := parser.Parse(&RawSchema{
			Ref: "#/components/schemas/Type",
		}, testCtx())
		require.Error(t, err)
	}
}

func TestSchemaRefToRef(t *testing.T) {
	// This regression test checks ref-to-ref handling.
	//
	// Such schema caused a false-positive infinite recursion error before.
	components := components{
		"first": {
			Ref: "#/components/schemas/second",
		},
		"second": {
			Ref: "#/components/schemas/third",
		},
		"third": {
			Ref: "#/components/schemas/actual",
		},
		"actual": {
			Type: "integer",
		},
		"referer": {
			Type: "object",
			Properties: RawProperties{
				{"Ref1", &RawSchema{Ref: "#/components/schemas/first"}},
				{"Ref2", &RawSchema{Ref: "#/components/schemas/first"}},
				{"Ref3", &RawSchema{Ref: "#/components/schemas/second"}},
			},
		},
	}
	parser := NewParser(Settings{
		Resolver: components,
	})
	_, err := parser.Parse(&RawSchema{
		Ref: "#/components/schemas/referer",
	}, testCtx())
	require.NoError(t, err)
}

func TestSchemaSideEffects(t *testing.T) {
	expectSide := []*Schema{
		{
			Type:     Object,
			Required: []string{"name", "id", "age"},
			Properties: []Property{
				{
					Name:     "name",
					Schema:   &Schema{Type: String},
					Required: true,
				},
				{
					Name:     "age",
					Schema:   &Schema{Type: Integer},
					Required: true,
				},
				{
					Name:     "id",
					Schema:   &Schema{Type: Integer},
					Required: true,
				},
			},
		},
	}

	expect := &Schema{
		Type:     Object,
		Required: []string{"id", "name", "owner"},
		Properties: []Property{
			{
				Name:     "name",
				Schema:   &Schema{Type: String},
				Required: true,
			},
			{
				Name:     "owner",
				Schema:   expectSide[0],
				Required: true,
			},
		},
	}

	parser := NewParser(Settings{})

	out, err := parser.Parse(&RawSchema{
		Type: "object",
		Properties: []RawProperty{
			{
				Name:   "name",
				Schema: &RawSchema{Type: "string"},
			},
			{
				Name: "owner",
				Schema: &RawSchema{
					Type: "object",
					Properties: []RawProperty{
						{
							Name:   "name",
							Schema: &RawSchema{Type: "string"},
						},
						{
							Name:   "age",
							Schema: &RawSchema{Type: "integer"},
						},
						{
							Name:   "id",
							Schema: &RawSchema{Type: "integer"},
						},
					},
					Required: []string{"name", "id", "age"},
				},
			},
		},
		Required: []string{"id", "name", "owner"},
	}, testCtx())

	require.NoError(t, err)
	require.Equal(t, expect, out)
}

func TestSchemaReferencedArray(t *testing.T) {
	components := components{
		"Pets": {
			Type: "array",
			Items: &RawItems{
				Item: &RawSchema{
					Type: "string",
				},
			},
		},
	}

	pets := &Schema{
		Type: Array,
		Ref:  Ref{Loc: "/root.json", Ptr: "#/components/schemas/Pets"},
		Item: &Schema{Type: String},
	}

	expectRefcache := map[jsonpointer.RefKey]*Schema{
		{Loc: "/root.json", Ptr: "#/components/schemas/Pets"}: pets,
	}

	expect := &Schema{
		Type:     Object,
		Required: []string{"pets"},
		Properties: []Property{
			{
				Name:     "pets",
				Schema:   pets,
				Required: true,
			},
		},
	}

	parser := NewParser(Settings{
		Resolver: components,
	})

	out, err := parser.Parse(&RawSchema{
		Type: "object",
		Properties: []RawProperty{
			{
				Name: "pets",
				Schema: &RawSchema{
					Ref: "#/components/schemas/Pets",
				},
			},
		},
		Required: []string{"pets"},
	}, testCtx())

	require.NoError(t, err)
	require.Equal(t, expectRefcache, parser.refcache)
	require.Equal(t, expect, out)
}

func TestSchemaExtensions(t *testing.T) {
	tests := []struct {
		raw       string
		expect    *Schema
		expectErr bool
	}{
		{
			`{"type": "string", "x-ogen-name": "Foo"}`,
			&Schema{
				Type:      String,
				XOgenName: "Foo",
			},
			false,
		},
		// Unexported type name.
		{`{"type": "string", "x-ogen-name": "foo"}`, nil, true},
		// Invalid type.
		{`{"type": "string", "x-ogen-name": {}}`, nil, true},
		{
			`{"type": "string", "x-ogen-time-format": "2006-01-02T15:04:05.999999999Z07:00"}`,
			&Schema{
				Type:            String,
				XOgenTimeFormat: "2006-01-02T15:04:05.999999999Z07:00",
			},
			false,
		},
		{
			`{"type": "string", "x-ogen-type": "github.com/google/uuid.UUID"}`,
			&Schema{
				Type:      String,
				XOgenType: "github.com/google/uuid.UUID",
			},
			false,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			data := []byte(tt.raw)

			var raw RawSchema
			a.NoError(yaml.Unmarshal(data, &raw))

			const filename = "test.yaml"
			out, err := NewParser(Settings{
				File: location.NewFile(filename, filename, data),
			}).Parse(&raw, testCtx())
			if tt.expectErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			// Zero locator to simplify comparison.
			out.Pointer = location.Pointer{}
			a.Equal(tt.expect, out)
		})
	}
}

func TestInvalidMultipleOf(t *testing.T) {
	values := []int{0, -1, -10}
	parser := NewParser(Settings{
		Resolver: components{},
	})
	for _, typ := range []string{
		"integer",
		"number",
	} {
		typ := typ
		t.Run(typ, func(t *testing.T) {
			for _, v := range values {
				_, err := parser.Parse(&RawSchema{
					Type:       typ,
					MultipleOf: strconv.AppendInt(nil, int64(v), 10),
				}, testCtx())
				require.Errorf(t, err, "%d", v)
			}
		})
		_, err := parser.Parse(&RawSchema{
			Type:       typ,
			MultipleOf: []byte("true"),
		}, testCtx())
		require.Error(t, err)
	}
}
