package http

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func newReq() *http.Request {
	req, err := http.NewRequest(http.MethodGet, "https://example.org", nil)
	if err != nil {
		panic(err)
	}
	return req
}

func TestSet(t *testing.T) {
	req := newReq()
	newCtx := context.WithValue(context.Background(), "foo", "bar")
	Set(req, newCtx)
	if v := req.Context().Value("foo").(string); v != "bar" {
		t.Errorf("unexpected value %s", v)
	}
}

func TestSetValue(t *testing.T) {
	req := newReq()
	SetValue(req, "foo", "bar")
	if v := req.Context().Value("foo").(string); v != "bar" {
		t.Errorf("unexpected value %s", v)
	}
}

func BenchmarkSet(b *testing.B) {
	b.ReportAllocs()
	req := newReq()
	newCtx := context.WithValue(context.Background(), "foo", "bar")

	for i := 0; i < b.N; i++ {
		Set(req, newCtx)
	}
}

func BenchmarkWithContext(b *testing.B) {
	b.ReportAllocs()
	req := newReq()
	newCtx := context.WithValue(context.Background(), "foo", "bar")

	for i := 0; i < b.N; i++ {
		req.WithContext(newCtx)
	}
}

func BenchmarkNewRequest(b *testing.B) {
	ctx := context.Background()
	u, err := url.Parse("https://go.dev")
	require.NoError(b, err)

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			req := NewRequest(ctx, http.MethodGet, CloneURL(u), nil)
			PutRequest(req)
		}
	})
	b.Run("Std", func(b *testing.B) {
		b.ReportAllocs()
		uStr := u.String()
		for i := 0; i < b.N; i++ {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, uStr, nil)
			if err != nil {
				b.Fatal(err)
			}
			_ = req
		}
	})
}
