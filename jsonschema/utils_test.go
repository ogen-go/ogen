package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRawSchemaFields(t *testing.T) {
	flse := false

	tests := []struct {
		Schema *RawSchema
		Expect []string
	}{
		{
			Schema: &RawSchema{
				Type: StringArray{"object"},
				Properties: RawProperties{
					RawProperty{
						Name:   "foo",
						Schema: &RawSchema{Type: StringArray{"null"}},
					},
				},
				AdditionalProperties: &AdditionalProperties{Bool: &flse},
			},
			Expect: []string{"type", "properties", "additionalProperties"},
		},
	}

	for _, test := range tests {
		fields, err := getRawSchemaFields(test.Schema)
		require.NoError(t, err)
		require.ElementsMatch(t, test.Expect, fields)
	}
}
