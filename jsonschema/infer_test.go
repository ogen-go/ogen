package jsonschema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInfer_Apply(t *testing.T) {
	t.Run("Bad", func(t *testing.T) {
		for _, input := range []string{
			``,
			`r`,
			`-..`,
			`{`,
			`[`,
			`[{`,
			`[{}`,
			`{"foo": [-..]}`,
		} {
			var i Infer
			require.Errorf(t, i.Apply([]byte(input)), "input:\n%s", input)
		}
	})

	tests := []struct {
		result RawSchema
		inputs []string
	}{
		{RawSchema{Type: "integer"}, []string{"1", "2", "3"}},
		{RawSchema{Type: "number"}, []string{"1", "2.0", "3"}},
		{RawSchema{Type: "number"}, []string{"2.0"}},
		{RawSchema{Type: "number", Nullable: true}, []string{"2.0", "null"}},

		{RawSchema{Type: "boolean"}, []string{"true", "false"}},
		{RawSchema{Type: "boolean", Nullable: true}, []string{"true", "null"}},

		{RawSchema{Type: "array"}, []string{"[]"}},
		{RawSchema{
			Type:  "array",
			Items: &RawSchema{Type: "integer"},
		}, []string{"[1]"}},
		{RawSchema{
			Type:  "array",
			Items: &RawSchema{Type: "number"},
		}, []string{"[1, 10, 5, 0.5]"}},
		{RawSchema{
			Type: "array",
			Items: &RawSchema{
				OneOf: []*RawSchema{
					{Type: "integer"},
					{Type: "boolean"},
					{Type: "string"},
				},
			},
		}, []string{`[1, true, "foo"]`}},

		{RawSchema{Type: "object", Properties: RawProperties{}}, []string{
			`{}`,
		}},
		{RawSchema{
			Type: "object",
			Properties: RawProperties{
				{"foo", &RawSchema{Type: "integer"}},
			},
		}, []string{
			`{}`,
			`{"foo": 1}`,
			`{"foo": 2}`,
			`{"foo": 3}`,
		}},
		{RawSchema{
			Type:     "object",
			Required: []string{"foo"},
			Properties: RawProperties{
				{"bar", &RawSchema{Type: "string"}},
				{"foo", &RawSchema{Type: "integer"}},
			},
		}, []string{
			`{"foo": 1}`,
			`{"foo": 5}`,
			`{"foo": 2, "bar": "baz"}`,
		}},
		{RawSchema{
			Type:     "object",
			Required: []string{"required", "required_nullable"},
			Properties: RawProperties{
				{"optional", &RawSchema{Type: "integer"}},
				{"optional_nullable", &RawSchema{Type: "integer", Nullable: true}},
				{"required", &RawSchema{Type: "integer"}},
				{"required_nullable", &RawSchema{Type: "integer", Nullable: true}},
			},
		}, []string{
			`{"required": 10, "required_nullable": null, "optional": 10, "optional_nullable": null}`,
			`{"required": 10, "required_nullable": 10}`,
			`{"required": 10, "required_nullable": 10, "optional_nullable": 10}`,
		}},

		{RawSchema{Nullable: true}, []string{"null"}},
		{RawSchema{
			OneOf: []*RawSchema{
				{Type: "boolean"},
				{Type: "string"},
				{Type: "number"},
			},
		}, []string{"true", `"foo"`, "10", "1.0"}},
		{RawSchema{
			OneOf: []*RawSchema{
				{Type: "boolean"},
				{Type: "string"},
				{Type: "number"},
			},
		}, []string{"true", `"foo"`, "1.0", "10"}},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			var i Infer
			for _, input := range tt.inputs {
				a.NoError(i.Apply([]byte(input)))
			}
			a.Equal(tt.result, i.Target())
		})
	}
}
