package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_required_recursive"
)

func TestRequiredRecursive(t *testing.T) {
	for _, input := range []string{
		`{"kind":"leaf"}`,
		`{"b":{"a":{"kind":"leaf"}}}`,
		`{"b":{"a":{"b":{"a":{"kind":"leaf"}}}}}`,
	} {
		var value api.A
		require.NoError(t, value.UnmarshalJSON([]byte(input)), input)
		require.NoError(t, value.Validate())
		got, err := value.MarshalJSON()
		require.NoError(t, err)
		require.JSONEq(t, input, string(got))
	}

	for _, input := range []string{`{}`, `{"b":null}`, `{"b":{}}`, `{"b":{"a":null}}`} {
		var value api.ARecursive
		require.Error(t, value.UnmarshalJSON([]byte(input)), input)
	}
	var missing api.ARecursive
	require.Error(t, missing.Validate())
}
