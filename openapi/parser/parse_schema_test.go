package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func TestParseDiscriminator(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/get": {
				Get: &ogen.Operation{
					OperationID: "testGet",
					Description: "operation description",
					Responses: map[string]*ogen.Response{
						"201": {
							Content: map[string]ogen.Media{
								"application/json": {
									Schema: &ogen.Schema{Ref: "#/components/schemas/TestPet"},
								},
							},
						},
					},
				},
			},
		},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				"TestPet": {
					Properties: ogen.Properties{
						{
							Name:   "PetRef",
							Schema: &ogen.Schema{Ref: "#/components/schemas/PetRef"},
						},
						{
							Name:   "PetSchemaName",
							Schema: &ogen.Schema{Ref: "#/components/schemas/PetSchemaName"},
						},
						{
							Name:   "PetImplicit",
							Schema: &ogen.Schema{Ref: "#/components/schemas/PetImplicit"},
						},
					},
				},
				"PetRef": {
					OneOf: []*ogen.Schema{
						{Ref: "#/components/schemas/Cat"},
						{Ref: "#/components/schemas/Dog"},
						{Ref: "#/components/schemas/Cow"},
					},
					Discriminator: &ogen.Discriminator{
						PropertyName: "petType",
						Mapping: map[string]string{
							"cat": "#/components/schemas/Cat",
							"dog": "#/components/schemas/Dog",
							"cow": "#/components/schemas/Cow",
						},
					},
				},
				"PetSchemaName": {
					OneOf: []*ogen.Schema{
						{Ref: "#/components/schemas/Cat"},
						{Ref: "#/components/schemas/Dog"},
						{Ref: "#/components/schemas/Cow"},
					},
					Discriminator: &ogen.Discriminator{
						PropertyName: "petType",
						Mapping: map[string]string{
							"cat": "Cat",
							"dog": "Dog",
							"cow": "Cow",
						},
					},
				},
				"PetImplicit": {
					OneOf: []*ogen.Schema{
						{Ref: "#/components/schemas/Cat"},
						{Ref: "#/components/schemas/Dog"},
						{Ref: "#/components/schemas/Cow"},
					},
					Discriminator: &ogen.Discriminator{
						PropertyName: "petType",
					},
				},
				"Cat": {
					Type:     "object",
					Required: []string{"petType", "meow"},
					Properties: ogen.Properties{
						{Name: "petType", Schema: &ogen.Schema{Type: "string"}},
						{Name: "meow", Schema: &ogen.Schema{Type: "string"}},
					},
				},
				"Dog": {
					Type:     "object",
					Required: []string{"petType", "bark"},
					Properties: ogen.Properties{
						{Name: "petType", Schema: &ogen.Schema{Type: "string"}},
						{Name: "bark", Schema: &ogen.Schema{Type: "string"}},
					},
				},
				"Cow": {
					Type:     "object",
					Required: []string{"petType", "moo"},
					Properties: ogen.Properties{
						{Name: "petType", Schema: &ogen.Schema{Type: "string"}},
						{Name: "moo", Schema: &ogen.Schema{Type: "string"}},
					},
				},
			},
		},
	}

	a := require.New(t)

	ref := func(ptr string) refKey {
		return refKey{Loc: testRootURL.String(), Ptr: ptr}
	}
	spec, err := Parse(root, Settings{
		RootURL: testRootURL,
	})
	a.NoError(err)
	{
		s := spec.Components.Schemas["PetRef"]
		m := s.Discriminator.Mapping
		a.Equal(m["cat"].Ref, ref("#/components/schemas/Cat"))
		a.Equal(m["dog"].Ref, ref("#/components/schemas/Dog"))
		a.Equal(m["cow"].Ref, ref("#/components/schemas/Cow"))
	}
	{
		s := spec.Components.Schemas["PetSchemaName"]
		m := s.Discriminator.Mapping
		a.Equal(m["cat"].Ref, ref("#/components/schemas/Cat"))
		a.Equal(m["dog"].Ref, ref("#/components/schemas/Dog"))
		a.Equal(m["cow"].Ref, ref("#/components/schemas/Cow"))
	}
}
