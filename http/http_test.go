package http

import (
	"context"
	"net/http"
	"testing"
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
	newCtx := context.WithValue(context.Background(), testKey{}, "bar")
	Set(req, newCtx)
	if v := req.Context().Value(testKey{}).(string); v != "bar" {
		t.Errorf("unexpected value %s", v)
	}
}

func TestSetValue(t *testing.T) {
	req := newReq()
	SetValue(req, testKey{}, "bar")
	if v := req.Context().Value(testKey{}).(string); v != "bar" {
		t.Errorf("unexpected value %s", v)
	}
}

func BenchmarkSet(b *testing.B) {
	b.ReportAllocs()
	req := newReq()
	newCtx := context.WithValue(context.Background(), testKey{}, "bar")

	for i := 0; i < b.N; i++ {
		Set(req, newCtx)
	}
}

type testKey struct{}

func BenchmarkWithContext(b *testing.B) {
	b.ReportAllocs()
	req := newReq()
	newCtx := context.WithValue(context.Background(), testKey{}, "bar")

	for i := 0; i < b.N; i++ {
		req.WithContext(newCtx)
	}
}
