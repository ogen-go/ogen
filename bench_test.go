package ogen

import (
	"bytes"
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

	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/internal/techempower"
	"github.com/ogen-go/ogen/json"
)

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
			b.SetBytes(int64(len(data)))
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
			b.SetBytes(int64(len("Hello, world!")))
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

	b.Run("Manual", func(b *testing.B) {
		// Test with some manual optimizations.
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			js := json.NewCustomStream(json.ConfigFastest, w, 1024)
			js.WriteObjectStart()
			js.WriteObjectField("message")
			js.WriteString("Hello, world!")
			js.WriteObjectEnd()
		}))
		defer s.Close()

		ctx := context.Background()
		client := &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
				MaxIdleConns:        100,
			},
			CheckRedirect: nil,
		}
		b.Run("JSON", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			u, err := url.Parse(s.URL)
			require.NoError(b, err)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := ht.NewRequest(ctx, http.MethodGet, u, nil)
					res, err := client.Do(req)
					ht.PutRequest(req)
					if err != nil {
						b.Error(err)
						break
					}
					_, _ = io.Copy(io.Discard, res.Body)
					_ = res.Body.Close()
					if res.StatusCode != http.StatusOK {
						b.Error(res.StatusCode)
						break
					}
				}
			})
		})
	})

	b.Run("TechEmpower", func(b *testing.B) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types

		mux := chi.NewRouter()
		srv := techEmpowerServer{}
		techempower.Register(mux, srv)
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
		b.Run("OnlyHandler", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					hw, err := srv.JSON(ctx)
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

func BenchmarkJSON(b *testing.B) {
	b.Run("TechEmpower", func(b *testing.B) {
		h := techempower.HelloWorld{
			Message: "Hello, world!",
		}
		data := json.Encode(h)
		dataBytes := int64(len(data))

		b.Run("Encode", func(b *testing.B) {
			buf := new(bytes.Buffer)
			s := json.NewStream(buf)
			b.ReportAllocs()
			b.SetBytes(dataBytes)

			for i := 0; i < b.N; i++ {
				buf.Reset()
				h.WriteJSON(s)
				if err := s.Flush(); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("Decode", func(b *testing.B) {
			var v techempower.HelloWorld
			b.ReportAllocs()
			b.SetBytes(dataBytes)
			j := json.NewIterator()

			for i := 0; i < b.N; i++ {
				j.ResetBytes(data)
				if err := v.ReadJSON(j); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
