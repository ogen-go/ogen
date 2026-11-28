package integration

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	refoverrides "github.com/ogen-go/ogen/internal/integration/test_ref_overrides"
)

func TestRefOverrides(t *testing.T) {
	r := refoverrides.TestOK{}
	require.NoError(t, r.Decode(jx.DecodeStr(`{}`)))
	require.True(t, r.First.Set)
	require.True(t, r.Second.Set)
	require.Equal(t, refoverrides.Str("first"), r.First.Value)
	require.Equal(t, refoverrides.Str("second"), r.Second.Value)
	require.False(t, r.Third.Set)

	r = refoverrides.TestOK{}
	require.NoError(t, r.Decode(jx.DecodeStr(`{"first": "foo", "second": "bar"}`)))
	require.Equal(t, refoverrides.Str("foo"), r.First.Value)
	require.Equal(t, refoverrides.Str("bar"), r.Second.Value)
}
