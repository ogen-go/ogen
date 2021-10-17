package http

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkCloneURL(b *testing.B) {
	u, err := url.Parse("https://go.dev")
	require.NoError(b, err)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uCloned := CloneURL(u)
		PutURL(uCloned)
	}
}
