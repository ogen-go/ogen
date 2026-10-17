//go:build go1.27

package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetExternalTypeGo127StdlibUUID(t *testing.T) {
	got, err := getExternalType("uuid.UUID")
	require.NoError(t, err)
	require.Equal(t, ExternalType{
		PackagePath: "uuid",
		PackageName: "uuid",
		TypeName:    "UUID",
		Encode:      ExternalText,
		Decode:      ExternalText,
	}, got)
}
