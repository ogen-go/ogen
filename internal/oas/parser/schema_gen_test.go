package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func TestSchemaSimple(t *testing.T) {
	gen := &schemaGen{
		localRefs: make(map[string]*oas.Schema),
	}

	out, err := gen.Generate(&ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{
				Name:   "id",
				Schema: &ogen.Schema{Type: "integer"},
			},
			{
				Name:   "name",
				Schema: &ogen.Schema{Type: "string"},
			},
		},
		Required: []string{"id", "name"},
	})
	require.NoError(t, err)

	expect := &oas.Schema{
		Type: oas.Object,
		Properties: []oas.Property{
			{
				Name:     "id",
				Schema:   &oas.Schema{Type: oas.Integer},
				Required: true,
			},
			{
				Name:     "name",
				Schema:   &oas.Schema{Type: oas.String},
				Required: true,
			},
		},
	}

	require.Equal(t, expect, out)
}

func TestSchemaRecursive(t *testing.T) {
	spec := &ogen.Spec{
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				"Pet": {
					Type: "object",
					Properties: []ogen.Property{
						{
							Name:   "id",
							Schema: &ogen.Schema{Type: "integer"},
						},
						{
							Name:   "name",
							Schema: &ogen.Schema{Type: "string"},
						},
						{
							Name: "friends",
							Schema: &ogen.Schema{
								Type: "array",
								Items: &ogen.Schema{
									Ref: "#/components/schemas/Pet",
								},
							},
						},
					},
					Required: []string{"id", "name", "friends"},
				},
			},
		},
	}

	pet := &oas.Schema{
		Type: oas.Object,
		Ref:  "#/components/schemas/Pet",
	}
	pet.Properties = []oas.Property{
		{
			Name:     "id",
			Schema:   &oas.Schema{Type: oas.Integer},
			Required: true,
		},
		{
			Name:     "name",
			Schema:   &oas.Schema{Type: oas.String},
			Required: true,
		},
		{
			Name: "friends",
			Schema: &oas.Schema{
				Type: oas.Array,
				Item: pet,
			},
			Required: true,
		},
	}

	expectLocalRefs := map[string]*oas.Schema{
		"#/components/schemas/Pet": {
			Type: oas.Object,
			Ref:  "#/components/schemas/Pet",
			Properties: []oas.Property{
				{
					Name:     "id",
					Schema:   &oas.Schema{Type: oas.Integer},
					Required: true,
				},
				{
					Name:     "name",
					Schema:   &oas.Schema{Type: oas.String},
					Required: true,
				},
				{
					Name: "friends",
					Schema: &oas.Schema{
						Type: oas.Array,
						Item: pet,
					},
					Required: true,
				},
			},
		},
	}

	gen := &schemaGen{
		spec:      spec,
		localRefs: make(map[string]*oas.Schema),
	}

	out, err := gen.Generate(&ogen.Schema{
		Ref: "#/components/schemas/Pet",
	})
	require.NoError(t, err)
	require.Equal(t, expectLocalRefs, gen.localRefs)
	require.Equal(t, pet, out)
}

func TestSchemaSideEffects(t *testing.T) {
	expectSide := []*oas.Schema{
		{
			Type: oas.Object,
			Properties: []oas.Property{
				{
					Name:     "name",
					Schema:   &oas.Schema{Type: oas.String},
					Required: true,
				},
				{
					Name:     "age",
					Schema:   &oas.Schema{Type: oas.Integer},
					Required: true,
				},
				{
					Name:     "id",
					Schema:   &oas.Schema{Type: oas.Integer},
					Required: true,
				},
			},
		},
	}

	expect := &oas.Schema{
		Type: oas.Object,
		Properties: []oas.Property{
			{
				Name:     "name",
				Schema:   &oas.Schema{Type: oas.String},
				Required: true,
			},
			{
				Name:     "owner",
				Schema:   expectSide[0],
				Required: true,
			},
		},
	}

	gen := &schemaGen{
		localRefs: make(map[string]*oas.Schema),
	}

	out, err := gen.Generate(&ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{
				Name:   "name",
				Schema: &ogen.Schema{Type: "string"},
			},
			{
				Name: "owner",
				Schema: &ogen.Schema{
					Type: "object",
					Properties: []ogen.Property{
						{
							Name:   "name",
							Schema: &ogen.Schema{Type: "string"},
						},
						{
							Name:   "age",
							Schema: &ogen.Schema{Type: "integer"},
						},
						{
							Name:   "id",
							Schema: &ogen.Schema{Type: "integer"},
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
	spec := &ogen.Spec{
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				"Pets": {
					Type: "array",
					Items: &ogen.Schema{
						Type: "string",
					},
				},
			},
		},
	}

	pets := &oas.Schema{
		Type: oas.Array,
		Ref:  "#/components/schemas/Pets",
		Item: &oas.Schema{Type: oas.String},
	}

	expectLocalRefs := map[string]*oas.Schema{
		"#/components/schemas/Pets": pets,
	}

	expect := &oas.Schema{
		Type: oas.Object,
		Properties: []oas.Property{
			{
				Name:     "pets",
				Schema:   pets,
				Required: true,
			},
		},
	}

	gen := &schemaGen{
		spec:      spec,
		localRefs: make(map[string]*oas.Schema),
	}

	out, err := gen.Generate(&ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{
				Name: "pets",
				Schema: &ogen.Schema{
					Ref: "#/components/schemas/Pets",
				},
			},
		},
		Required: []string{"pets"},
	})

	require.NoError(t, err)
	require.Equal(t, expectLocalRefs, gen.localRefs)
	require.Equal(t, expect, out)
}
