package integration

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_unique_mac_addresses"
	"github.com/ogen-go/ogen/validate"
)

func TestUniqueMACAddresses(t *testing.T) {
	for _, tt := range []struct {
		Name, Field, Items, WantError string
	}{
		{"Distinct", "addresses", `["00:11:22:33:44:55","00:11:22:33:44:56"]`, ""},
		{"Duplicate", "addresses", `["00:11:22:33:44:55","00:11:22:33:44:55"]`, "duplicate element [0]"},
		{"NonAdjacentDuplicate", "addresses", `["00:11:22:33:44:55","00:11:22:33:44:56","00:11:22:33:44:55"]`, "duplicate element [0]"},
		{"LaterDuplicate", "addresses", `["00:11:22:33:44:55","00:11:22:33:44:56","00:11:22:33:44:56"]`, "duplicate element [1]"},
		{"EmptyAliases", "aliases", `[]`, ""},
		{"DistinctAliases", "aliases", `["00:11:22:33:44:55","00:11:22:33:44:56"]`, ""},
		{"DuplicateAliases", "aliases", `["00:11:22:33:44:55","00:11:22:33:44:55"]`, "duplicate element [0]"},
		{"NullableAndValue", "nullable", `[null,"00:11:22:33:44:55"]`, ""},
		{"DistinctNullable", "nullable", `["00:11:22:33:44:55","00:11:22:33:44:56"]`, ""},
		{"DuplicateNullable", "nullable", `["00:11:22:33:44:55","00:11:22:33:44:55"]`, "duplicate element [0]"},
		{"DuplicateNull", "nullable", `[null,null]`, "duplicate element [0]"},
		{"NullableAliasAndValue", "nullableAliases", `[null,"00:11:22:33:44:55"]`, ""},
		{"DuplicateNullableAliases", "nullableAliases", `["00:11:22:33:44:55","00:11:22:33:44:55"]`, "duplicate element [0]"},
		{"DuplicateNullAliases", "nullableAliases", `[null,null]`, "duplicate element [0]"},
		{"NoUniqueConstraint", "unrestricted", `["00:11:22:33:44:55","00:11:22:33:44:55"]`, ""},
		{"DistinctStrings", "names", `["a","b"]`, ""},
		{"DuplicateStrings", "names", `["a","a"]`, "duplicate element [0]"},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			input := fmt.Sprintf(`{"addresses":["00:11:22:33:44:55"],%q:%s}`, tt.Field, tt.Items)
			if tt.Field == "addresses" {
				input = fmt.Sprintf(`{%q:%s}`, tt.Field, tt.Items)
			}
			var value api.AddressSet
			require.NoError(t, value.Decode(jx.DecodeStr(input)))
			err := value.Validate()
			if tt.WantError == "" {
				require.NoError(t, err)
				return
			}
			var validationError *validate.Error
			require.ErrorAs(t, err, &validationError)
			require.Len(t, validationError.Fields, 1)
			require.Equal(t, tt.Field, validationError.Fields[0].Name)
			require.ErrorContains(t, validationError.Fields[0].Error, tt.WantError)
		})
	}
}
