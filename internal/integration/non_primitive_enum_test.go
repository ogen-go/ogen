package integration

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_non_primitive_enum"
)

func TestNonPrimitiveEnum(t *testing.T) {
	t.Run("ObjectEnum", func(t *testing.T) {
		t.Run("Decode", func(t *testing.T) {
			// Test decoding each variant
			tests := []struct {
				name     string
				json     string
				wantType api.ObjectEnumType
			}{
				{"foo", `{"type":"foo","value":1}`, api.ObjectEnumFooObjectEnum},
				{"bar", `{"type":"bar","value":2}`, api.ObjectEnumBarObjectEnum},
				{"baz", `{"type":"baz","value":3}`, api.ObjectEnumBazObjectEnum},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var obj api.ObjectEnum
					d := jx.DecodeStr(tt.json)
					err := obj.Decode(d)
					require.NoError(t, err)
					require.Equal(t, tt.wantType, obj.Type)
				})
			}
		})

		t.Run("Encode", func(t *testing.T) {
			// Test encoding ObjectEnumFoo
			obj := api.NewObjectEnumFooObjectEnum(api.ObjectEnumFoo{
				Type:  "foo",
				Value: 1,
			})

			var enc jx.Encoder
			obj.Encode(&enc)
			require.Contains(t, enc.String(), `"type":"foo"`)
			require.Contains(t, enc.String(), `"value":1`)
		})

		t.Run("Getters", func(t *testing.T) {
			// Test getter methods
			obj := api.NewObjectEnumBarObjectEnum(api.ObjectEnumBar{
				Type:  "bar",
				Value: 2,
			})

			// IsObjectEnumBar should be true
			require.True(t, obj.IsObjectEnumBar())
			require.False(t, obj.IsObjectEnumFoo())
			require.False(t, obj.IsObjectEnumBaz())

			// GetObjectEnumBar should return the value
			bar, ok := obj.GetObjectEnumBar()
			require.True(t, ok)
			require.Equal(t, "bar", bar.Type)
			require.Equal(t, int64(2), bar.Value)
		})
	})
}
