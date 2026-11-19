package integration_test

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_required_nullable_recursive"
)

// TestRequiredNullableRecursive verifies that required nullable properties
// taking part in a recursive cycle accept an explicit null, reject an absent
// property, and round-trip through encoding.
//
// See https://github.com/ogen-go/ogen/issues/1791.
func TestRequiredNullableRecursive(t *testing.T) {
	type decoder interface {
		Decode(d *jx.Decoder) error
		MarshalJSON() ([]byte, error)
	}
	tests := []struct {
		name    string
		newVal  func() decoder
		input   string
		wantErr bool
	}{
		// anyOf: [AnyOfA, {type: 'null'}]
		{"AnyOf/Absent", func() decoder { return new(api.AnyOfB) }, `{}`, true},
		{"AnyOf/Null", func() decoder { return new(api.AnyOfB) }, `{"a":null}`, false},
		{"AnyOf/NestedNull", func() decoder { return new(api.AnyOfB) }, `{"a":{"b":{"a":null}}}`, false},
		{"AnyOf/NestedAbsent", func() decoder { return new(api.AnyOfB) }, `{"a":{"b":{}}}`, true},
		{"AnyOf/NonNullableNull", func() decoder { return new(api.AnyOfA) }, `{"b":null}`, true},
		// type: [object, 'null']
		{"TypeList/Absent", func() decoder { return new(api.TypeListNode) }, `{}`, true},
		{"TypeList/Null", func() decoder { return new(api.TypeListNode) }, `{"next":null}`, false},
		{"TypeList/NestedNull", func() decoder { return new(api.TypeListNode) }, `{"next":{"next":null}}`, false},
		{"TypeList/NestedAbsent", func() decoder { return new(api.TypeListNode) }, `{"next":{}}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			v := tt.newVal()
			err := v.Decode(jx.DecodeStr(tt.input))
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)

			data, err := v.MarshalJSON()
			a.NoError(err)
			a.JSONEq(tt.input, string(data))
		})
	}
}
