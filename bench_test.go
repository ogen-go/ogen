package ogen

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ogen-go/ogen/internal/techempower"
)

type techEmpowerServer struct{}

func (t techEmpowerServer) JSON(ctx context.Context) (*techempower.HelloWorld, error) {
	return &techempower.HelloWorld{
		Message: "Hello, world!",
	}, nil
}

func BenchmarkIntegration(b *testing.B) {
	// Using TechEmpower as most popular general purpose framework benchmark.
	b.Run("TechEmpower", func(b *testing.B) {
		b.ReportAllocs()

		mux := chi.NewRouter()
		techempower.Register(mux, techEmpowerServer{})
		s := httptest.NewServer(mux)
		defer s.Close()

		client := techempower.NewClient(s.URL)
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				hw, err := client.JSON(ctx)
				if err != nil {
					b.Error(err)
					return
				}
				if hw.Message != "Hello, world!" {
					b.Error("mismatch")
				}
			}
		})
	})
}
