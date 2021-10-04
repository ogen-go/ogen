package ogen

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/techempower"
)

type techEmpowerServer struct{}

func (t techEmpowerServer) Caching(ctx context.Context, params techempower.CachingParams) (*techempower.WorldObjects, error) {
	panic("implement me")
}

func (t techEmpowerServer) Updates(ctx context.Context, params techempower.UpdatesParams) (*techempower.WorldObjects, error) {
	panic("implement me")
}

func (t techEmpowerServer) Queries(ctx context.Context, params techempower.QueriesParams) (*techempower.WorldObjects, error) {
	return &techempower.WorldObjects{}, nil
}

func (t techEmpowerServer) DB(ctx context.Context) (*techempower.WorldObject, error) {
	return &techempower.WorldObject{
		ID:           1,
		RandomNumber: 10,
	}, nil
}

func (t techEmpowerServer) JSON(ctx context.Context) (*techempower.HelloWorld, error) {
	return &techempower.HelloWorld{
		Message: "Hello, world!",
	}, nil
}

func TestIntegration(t *testing.T) {
	t.Run("TechEmpower", func(t *testing.T) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types

		mux := chi.NewRouter()
		techempower.Register(mux, techEmpowerServer{})
		s := httptest.NewServer(mux)
		defer s.Close()

		client := techempower.NewClient(s.URL)
		ctx := context.Background()

		t.Run("JSON", func(t *testing.T) {
			res, err := client.JSON(ctx)
			require.NoError(t, err)
			require.Equal(t, "Hello, world!", res.Message)
		})
		t.Run("DB", func(t *testing.T) {
			res, err := client.DB(ctx)
			require.NoError(t, err)
			require.Equal(t, int64(1), res.ID)
			require.Equal(t, int64(10), res.RandomNumber)
		})
	})
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func BenchmarkIntegration(b *testing.B) {
	b.Run("Baseline", func(b *testing.B) {
		// Use baseline implementation to measure framework overhead.
		b.Run("Std", func(b *testing.B) {
			data := []byte(`Hello, world!`)
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(data)
			}))
			defer s.Close()

			client := s.Client()

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := func() error {
						res, err := client.Get(s.URL)
						if err != nil {
							return err
						}
						defer func() {
							_ = res.Body.Close()
						}()
						if _, err := io.ReadAll(res.Body); err != nil {
							return err
						}
						if res.StatusCode != http.StatusOK {
							return xerrors.Errorf("code: %d", res.StatusCode)
						}

						return nil
					}(); err != nil {
						b.Error(err)
					}
				}
			})
		})
		b.Run("Fasthttp", func(b *testing.B) {
			done := make(chan struct{})
			defer func() { <-done }()

			ln := newLocalListener()
			defer func() { _ = ln.Close() }()

			go func() {
				defer close(done)
				if err := fasthttp.Serve(ln, func(ctx *fasthttp.RequestCtx) {
					_, _ = ctx.WriteString("Hello, world!")
				}); err != nil {
					b.Error(err)
				}
			}()

			c := &fasthttp.Client{}
			u := (&url.URL{
				Host:   ln.Addr().String(),
				Scheme: "http",
			}).String()

			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				var dst []byte
				for pb.Next() {
					code, result, err := c.Get(dst, u)
					if err != nil {
						b.Error(err)
						return
					}
					if code != http.StatusOK {
						b.Errorf("bad code %d:", code)
						return
					}

					// Reusing buffer.
					dst = result[:0]
				}
			})
		})
	})

	b.Run("TechEmpower", func(b *testing.B) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types

		mux := chi.NewRouter()
		techempower.Register(mux, techEmpowerServer{})
		s := httptest.NewServer(mux)
		defer s.Close()

		client := techempower.NewClient(s.URL)
		ctx := context.Background()

		b.Run("JSON", func(b *testing.B) {
			b.ReportAllocs()
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
	})
}
