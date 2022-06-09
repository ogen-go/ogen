package jsonschema

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
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
	})
	require.NoError(t, err)

	expect := &Schema{
		Type: Object,
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
						Items: &RawSchema{
							Ref: "#/components/schemas/Pet",
						},
					},
				},
			},
			Required: []string{"id", "name", "friends"},
		},
	}

	pet := &Schema{
		Type: Object,
		Ref:  "#/components/schemas/Pet",
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

	expectRefcache := map[refKey]*Schema{
		refKey{ref: "#/components/schemas/Pet"}: {
			Type: Object,
			Ref:  "#/components/schemas/Pet",
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
	})
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
		})
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
	})
	require.NoError(t, err)
}

func TestSchemaSideEffects(t *testing.T) {
	expectSide := []*Schema{
		{
			Type: Object,
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
		Type: Object,
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
	})

	require.NoError(t, err)
	require.Equal(t, expect, out)
}

func TestSchemaReferencedArray(t *testing.T) {
	components := components{
		"Pets": {
			Type: "array",
			Items: &RawSchema{
				Type: "string",
			},
		},
	}

	pets := &Schema{
		Type: Array,
		Ref:  "#/components/schemas/Pets",
		Item: &Schema{Type: String},
	}

	expectRefcache := map[refKey]*Schema{
		refKey{ref: "#/components/schemas/Pets"}: pets,
	}

	expect := &Schema{
		Type: Object,
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
	})

	require.NoError(t, err)
	require.Equal(t, expectRefcache, parser.refcache)
	require.Equal(t, expect, out)
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
		t.Run(typ, func(t *testing.T) {
			for _, v := range values {
				_, err := parser.Parse(&RawSchema{
					Type:       typ,
					MultipleOf: []byte(fmt.Sprintf("%q", v)),
				})
				require.Errorf(t, err, "%d", v)
			}
		})
	}
}
