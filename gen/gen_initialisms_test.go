package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func TestInitialismsFeatureE2E(t *testing.T) {
	objSchema := &ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{Name: "userId", Schema: &ogen.Schema{Type: "string"}},
			{Name: "httpUrl", Schema: &ogen.Schema{Type: "string"}},
		},
	}
	spec := &ogen.Spec{
		OpenAPI: "3.0.3",
		Info:    ogen.Info{Title: "test", Version: "1.0.0"},
		Paths: ogen.Paths{
			"/obj": &ogen.PathItem{
				Get: &ogen.Operation{
					OperationID: "getObj",
					Responses: ogen.Responses{
						"200": &ogen.Response{
							Description: "ok",
							Content: map[string]ogen.Media{
								"application/json": {Schema: objSchema},
							},
						},
					},
				},
			},
		},
	}

	gen := func(t *testing.T, enable bool) map[string]struct{} {
		opts := Options{}
		if enable {
			opts.Generator.Features = &FeatureOptions{Enable: FeatureSet{NamingInitialisms.Name: {}}}
		}
		g, err := NewGenerator(spec, opts)
		require.NoError(t, err)
		fields := map[string]struct{}{}
		for _, typ := range g.tstorage.types {
			for _, f := range typ.Fields {
				fields[f.Name] = struct{}{}
			}
		}
		return fields
	}

	off := gen(t, false)
	require.Contains(t, off, "UserId")
	require.Contains(t, off, "HttpUrl")

	on := gen(t, true)
	require.Contains(t, on, "UserID")
	require.Contains(t, on, "HTTPURL")
}

// TestInitialismsVariantNames ensures the feature also reaches enum and
// discriminator variant naming, not only struct fields.
func TestInitialismsVariantNames(t *testing.T) {
	enumName := func(t *testing.T, initialisms bool) string {
		gen, err := namer{initialisms: initialisms}.enumVariantNameGen("Status", []any{"userId"})
		require.NoError(t, err)
		name, err := gen("userId", 0)
		require.NoError(t, err)
		return name
	}
	discriminatorName := func(t *testing.T, initialisms bool) string {
		gen, err := namer{initialisms: initialisms}.discriminatorMappingNameGen("Pet", []string{"userId"})
		require.NoError(t, err)
		name, err := gen("userId", 0)
		require.NoError(t, err)
		return name
	}

	t.Run("Enum", func(t *testing.T) {
		require.Equal(t, "StatusUserId", enumName(t, false))
		require.Equal(t, "StatusUserID", enumName(t, true))
	})
	t.Run("Discriminator", func(t *testing.T) {
		require.Equal(t, "PetUserId", discriminatorName(t, false))
		require.Equal(t, "PetUserID", discriminatorName(t, true))
	})
}
