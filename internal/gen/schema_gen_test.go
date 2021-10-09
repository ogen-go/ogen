package gen

import (
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
	"github.com/stretchr/testify/require"
)

func TestSchemaGen(t *testing.T) {
	tests := []struct {
		TestName  string
		Spec      *ogen.Spec
		Name      string
		Input     ogen.Schema
		Expect    *ast.Schema
		Err       error
		InputRefs map[string]*ast.Schema
		Side      []*ast.Schema
	}{
		{
			TestName: "Simple",
			Name:     "pet",
			Input: ogen.Schema{
				Type: "object",
				Properties: map[string]ogen.Schema{
					"id":   {Type: "integer"},
					"name": {Type: "string"},
				},
				Required: []string{"id", "name"},
			},
			Expect: &ast.Schema{
				Name: "Pet",
				Kind: ast.KindStruct,
				Fields: []ast.SchemaField{
					{
						Name: "ID",
						Type: "int",
						Tag:  "id",
					},
					{
						Name: "Name",
						Type: "string",
						Tag:  "name",
					},
				},
			},
		},
		{
			TestName: "Recursive",
			Spec: &ogen.Spec{
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
			},
			Name: "",
			Input: ogen.Schema{
				Ref: "#/components/schemas/Pet",
			},
			InputRefs: make(map[string]*ast.Schema),
			Expect: &ast.Schema{
				Name: "Pet",
				Kind: ast.KindStruct,
				Fields: []ast.SchemaField{
					{
						Name: "Friends",
						Type: "[]Pet",
						Tag:  "friends",
					},
					{
						Name: "ID",
						Type: "int",
						Tag:  "id",
					},
					{
						Name: "Name",
						Type: "string",
						Tag:  "name",
					},
				},
			},
		},
		{
			TestName: "TestSideEffects",
			Name:     "pet",
			Input: ogen.Schema{
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
			},
			Expect: &ast.Schema{
				Name: "Pet",
				Kind: ast.KindStruct,
				Fields: []ast.SchemaField{
					{
						Name: "Name",
						Type: "string",
						Tag:  "name",
					},
					{
						Name: "Owner",
						Type: "PetOwner",
						Tag:  "owner",
					},
				},
			},
			Side: []*ast.Schema{
				{
					Kind: ast.KindStruct,
					Name: "PetOwner",
					Fields: []ast.SchemaField{
						{
							Name: "Age",
							Type: "int",
							Tag:  "age",
						},
						{
							Name: "ID",
							Type: "int",
							Tag:  "id",
						},
						{
							Name: "Name",
							Type: "string",
							Tag:  "name",
						},
					},
				},
			},
		},
		{
			TestName: "ReferencedArray",
			Spec: &ogen.Spec{
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
			},
			Name: "TestObj",
			Input: ogen.Schema{
				Type: "object",
				Properties: map[string]ogen.Schema{
					"pets": {
						Ref: "#/components/schemas/Pets",
					},
				},
				Required: []string{"pets"},
			},
			Expect: &ast.Schema{
				Kind: ast.KindStruct,
				Name: "TestObj",
				Fields: []ast.SchemaField{
					{
						Name: "Pets",
						// TODO: type Pets []string?
						Type: "[]string",
						Tag:  "pets",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			gen := &schemaGen{
				spec:       test.Spec,
				globalRefs: test.InputRefs,
				localRefs:  make(map[string]*ast.Schema),
			}

			out, err := gen.Generate(test.Name, test.Input)
			if test.Err == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.Err.Error())
			}

			require.Equal(t, test.Expect, out, "schema check")
			require.Equal(t, test.Side, gen.side, "sideEffects check")
		})
	}
}
