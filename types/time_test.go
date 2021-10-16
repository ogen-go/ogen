package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	var v Time
	s := `"20:30:15"`

	require.NoError(t, v.UnmarshalJSON([]byte(s)))

	b, err := v.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, s, string(b))
}

func TestDate(t *testing.T) {
	var v Date
	s := `"2021-10-10"`

	require.NoError(t, v.UnmarshalJSON([]byte(s)))

	b, err := v.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, s, string(b))
}

func TestDuration(t *testing.T) {
	var v Duration
	s := `"100h0m0s"`

	require.NoError(t, v.UnmarshalJSON([]byte(s)))

	b, err := v.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, s, string(b))
}
