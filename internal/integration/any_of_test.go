package integration

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_anyof"
)

func TestValidateSum(t *testing.T) {
	for i, tc := range []struct {
		Input string
		Error bool
	}{
		{
			`{"medium": "text", "sizeLimit": "aboba"}`,
			true,
		},
		{
			`{"medium": "text", "sizeLimit": 10}`,
			false,
		},
		{
			`{"medium": "text", "sizeLimit": "10"}`,
			false,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			m := api.JaegerAnyOf{}
			require.NoError(t, m.Decode(jx.DecodeStr(tc.Input)))

			checker := require.NoError
			if tc.Error {
				checker = require.Error
			}
			checker(t, m.Validate())
		})
	}
}

func TestAnyOf(t *testing.T) {
	t.Run("JaegerAnyOf", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.JaegerAnyOfSizeLimitType
			Error    bool
		}{
			{`10`, api.IntJaegerAnyOfSizeLimit, false},
			{`"10"`, api.StringJaegerAnyOfSizeLimit, false},
			{`true`, "", true},
			{`null`, "", true},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.JaegerAnyOfSizeLimit{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				require.Equal(t, tc.Expected, r.Type)
			})
		}
	})
	t.Run("OneUUID", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.OneUUIDSubscriptionIDType
			Error    bool
		}{
			{`"fc9d49c6-1f3d-4ecb-92c7-be6d5049b3c8"`, api.SubscriptionUUIDOneUUIDSubscriptionID, false},
			{`true`, "", true},
			{`null`, "", true},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.OneUUIDSubscriptionID{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				require.Equal(t, tc.Expected, r.Type)
			})
		}
	})
}
