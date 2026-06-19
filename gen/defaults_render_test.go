package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func TestDefaultRenderHelpers(t *testing.T) {
	t.Run("LocalVarAndDepth", func(t *testing.T) {
		e := DefaultElem{Depth: 2}
		require.Equal(t, "defaultVal2", e.LocalVar())
		require.Equal(t, 3, e.NextDepth())
	})

	t.Run("Slice", func(t *testing.T) {
		require.Equal(t, []any{"a", "b"}, defaultSlice(ir.Default{Value: []any{"a", "b"}, Set: true}))
		require.Nil(t, defaultSlice(ir.Default{Value: "x", Set: true}))
	})

	t.Run("StructFields", func(t *testing.T) {
		typ := &ir.Type{
			Fields: []*ir.Field{
				{Name: "Name", Type: &ir.Type{}, Spec: &jsonschema.Property{Name: "name"}},
				{Name: "Count", Type: &ir.Type{}, Spec: &jsonschema.Property{Name: "count"}},
				{Name: "Extra", Type: &ir.Type{}, Spec: &jsonschema.Property{Name: "extra"}},
			},
		}
		got := defaultStructFields(typ, ir.Default{
			Value: map[string]any{"name": "x", "count": int64(5)},
			Set:   true,
		})
		require.Len(t, got, 2, "only fields present in the default map are emitted")
		require.Equal(t, "Name", got[0].Field.Name)
		require.Equal(t, "x", got[0].Value)
		require.Equal(t, "Count", got[1].Field.Name)
		require.Equal(t, int64(5), got[1].Value)
	})

	t.Run("MapEntriesSorted", func(t *testing.T) {
		got := defaultMapEntries(ir.Default{
			Value: map[string]any{"b": 2, "a": 1},
			Set:   true,
		})
		require.Equal(t, []DefaultMapEntry{{Key: "a", Value: 1}, {Key: "b", Value: 2}}, got)
		require.Empty(t, defaultMapEntries(ir.Default{Value: map[string]any{}, Set: true}))
		require.Nil(t, defaultMapEntries(ir.Default{Set: false}))
	})

	t.Run("JSONDeterministic", func(t *testing.T) {
		s, err := defaultJSON(map[string]any{"type": "foo", "a": 1})
		require.NoError(t, err)
		require.Equal(t, `{"a":1,"type":"foo"}`, s, "keys must be sorted for reproducible output")
	})
}
