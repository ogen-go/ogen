package uri

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkCloneURL(b *testing.B) {
	u, err := url.Parse("https://go.dev")
	require.NoError(b, err)
	var uCloned *url.URL

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		uCloned = Clone(u)
	}

	if uCloned.Host != "go.dev" {
		b.Fatal(uCloned)
	}
}
