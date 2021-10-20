package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func TestSchemaSimple(t *testing.T) {
	gen := &schemaGen{
		localRefs: make(map[string]*ast.Schema),
	}

	out, err := gen.Generate("Pet", ogen.Schema{
		Type: "object",
		Properties: map[string]ogen.Schema{
			"id":   {Type: "integer"},
			"name": {Type: "string"},
		},
		Required: []string{"id", "name"},
	})
	require.NoError(t, err)

	expect := &ast.Schema{
		Name: "Pet",
		Kind: ast.KindStruct,
		Fields: []ast.SchemaField{
			{
				Name: "ID",
				Type: ast.Primitive("int"),
				Tag:  "id",
			},
			{
				Name: "Name",
				Type: ast.Primitive("string"),
				Tag:  "name",
			},
		},
	}

	require.Equal(t, expect, out)
}

func TestSchemaRecursive(t *testing.T) {
	spec := &ogen.Spec{
		Components: &ogen.Components{
			Schemas: map[string]ogen.Schema{
				"Pet": {
					Type: "object",
					Properties: map[string]ogen.Schema{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
						"friends": {
							Type: "array",
							Items: &ogen.Schema{
								Ref: "#/components/schemas/Pet",
							},
						},
					},
					Required: []string{"id", "name", "friends"},
				},
			},
		},
	}

	pet := &ast.Schema{
		Name: "Pet",
		Kind: ast.KindStruct,
		Doc:  "Pet describes #/components/schemas/Pet.",
	}
	pet.Fields = []ast.SchemaField{
		{
			Name: "Friends",
			Type: ast.Array(pet),
			Tag:  "friends",
		},
		{
			Name: "ID",
			Type: ast.Primitive("int"),
			Tag:  "id",
		},
		{
			Name: "Name",
			Type: ast.Primitive("string"),
			Tag:  "name",
		},
	}

	expectLocalRefs := map[string]*ast.Schema{
		"#/components/schemas/Pet": {
			Name: "Pet",
			Kind: ast.KindStruct,
			Doc:  "Pet describes #/components/schemas/Pet.",
			Fields: []ast.SchemaField{
				{
					Name: "Friends",
					Type: ast.Array(pet),
					Tag:  "friends",
				},
				{
					Name: "ID",
					Type: ast.Primitive("int"),
					Tag:  "id",
				},
				{
					Name: "Name",
					Type: ast.Primitive("string"),
					Tag:  "name",
				},
			},
		},
	}

	gen := &schemaGen{
		spec:      spec,
		localRefs: make(map[string]*ast.Schema),
	}

	out, err := gen.Generate("", ogen.Schema{
		Ref: "#/components/schemas/Pet",
	})
	require.NoError(t, err)
	require.Equal(t, expectLocalRefs, gen.localRefs)
	require.Equal(t, pet, out)
}

func TestSchemaSideEffects(t *testing.T) {
	expectSide := []*ast.Schema{
		{
			Kind: ast.KindStruct,
			Name: "PetOwner",
			Fields: []ast.SchemaField{
				{
					Name: "Age",
					Type: ast.Primitive("int"),
					Tag:  "age",
				},
				{
					Name: "ID",
					Type: ast.Primitive("int"),
					Tag:  "id",
				},
				{
					Name: "Name",
					Type: ast.Primitive("string"),
					Tag:  "name",
				},
			},
		},
	}

	expect := &ast.Schema{
		Name: "Pet",
		Kind: ast.KindStruct,
		Fields: []ast.SchemaField{
			{
				Name: "Name",
				Type: ast.Primitive("string"),
				Tag:  "name",
			},
			{
				Name: "Owner",
				Type: expectSide[0],
				Tag:  "owner",
			},
		},
	}

	gen := &schemaGen{
		localRefs: make(map[string]*ast.Schema),
	}

	out, err := gen.Generate("pet", ogen.Schema{
		Type: "object",
		Properties: map[string]ogen.Schema{
			"name": {Type: "string"},
			"owner": {
				Type: "object",
				Properties: map[string]ogen.Schema{
					"name": {Type: "string"},
					"id":   {Type: "integer"},
					"age":  {Type: "integer"},
				},
				Required: []string{"name", "id", "age"},
			},
		},
		Required: []string{"id", "name", "owner"},
	})

	require.NoError(t, err)
	require.Equal(t, expect, out)
	require.Equal(t, expectSide, gen.side)
}

func TestSchemaReferencedArray(t *testing.T) {
	spec := &ogen.Spec{
		Components: &ogen.Components{
			Schemas: map[string]ogen.Schema{
				"Pets": {
					Type: "array",
					Items: &ogen.Schema{
						Type: "string",
					},
				},
			},
		},
	}

	pets := &ast.Schema{
		Kind: ast.KindAlias,
		Name: "Pets",
		AliasTo: &ast.Schema{
			Kind:        ast.KindArray,
			NilSemantic: ast.NilInvalid,
			Item: &ast.Schema{
				Kind:      ast.KindPrimitive,
				Primitive: "string",
			},
		},
	}

	expectLocalRefs := map[string]*ast.Schema{
		"#/components/schemas/Pets": pets,
	}

	expect := &ast.Schema{
		Kind: ast.KindStruct,
		Name: "TestObj",
		Fields: []ast.SchemaField{
			{
				Name: "Pets",
				Type: pets,
				Tag:  "pets",
			},
		},
	}

	gen := &schemaGen{
		spec:      spec,
		localRefs: make(map[string]*ast.Schema),
	}

	out, err := gen.Generate("testObj", ogen.Schema{
		Type: "object",
		Properties: map[string]ogen.Schema{
			"pets": {
				Ref: "#/components/schemas/Pets",
			},
		},
		Required: []string{"pets"},
	})

	require.NoError(t, err)
	require.Equal(t, expectLocalRefs, gen.localRefs)
	require.Equal(t, expect, out)
}
