package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
)

func BenchmarkFindRoute(b *testing.B) {
	bench := func(method, path string) func(b *testing.B) {
		return func(b *testing.B) {
			handler := &sampleAPIServer{}
			s, err := api.NewServer(handler, handler)
			require.NoError(b, err)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s.FindRoute(method, path)
			}
		}
	}

	b.Run("Plain", bench(http.MethodGet, "/pet"))
	b.Run("Parameters", bench(http.MethodGet, "/pet/name/10"))
}

func FuzzRouter(f *testing.F) {
	for _, tc := range routerTestCases() {
		f.Add(tc.Method, tc.Path)
	}

	handler := &sampleAPIServer{}
	s, err := api.NewServer(handler, handler)
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, method, path string) {
		s.FindRoute(method, path)
	})
}

type routerTestCase struct {
	Method    string
	Path      string
	Operation string
	Args      []string
	Defined   bool
}

func (r routerTestCase) defined() routerTestCase {
	return routerTestCase{
		Method:    r.Method,
		Path:      r.Path,
		Operation: r.Operation,
		Args:      r.Args,
		Defined:   true,
	}
}

func routerTestCases() []routerTestCase {
	test := func(method, route, op string, args ...string) routerTestCase {
		if len(args) == 0 {
			args = []string{}
		}
		return routerTestCase{
			Method:    method,
			Path:      route,
			Operation: op,
			Args:      args,
		}
	}
	get := func(p, op string, args ...string) routerTestCase {
		return test(http.MethodGet, p, op, args...)
	}
	post := func(p, op string, args ...string) routerTestCase {
		return test(http.MethodPost, p, op, args...)
	}
	put := func(p, op string, args ...string) routerTestCase {
		return test(http.MethodPut, p, op, args...)
	}
	del := func(p, op string, args ...string) routerTestCase {
		return test(http.MethodDelete, p, op, args...)
	}

	return []routerTestCase{
		get("/pet/name/10", "PetNameByID", "10"),
		get("/pet/friendNames/10", "PetFriendsNamesByID", "10"),
		get("/pet", "PetGet"),
		get("/pet/avatar", "PetGetAvatarByID"),
		post("/pet/avatar", "PetUploadAvatarByID"),
		get("/pet/aboba", "PetGetByName", "aboba"),
		get("/pet/aboba/avatar", "PetGetAvatarByName", "aboba"),
		get("/foobar", "FoobarGet"),
		post("/foobar", "FoobarPost"),
		put("/foobar", "FoobarPut"),
		get("/error", "ErrorGet"),
		get("/test/header", "GetHeader"),
		// "/name/{id}/{foo}1234{bar}-{baz}!{kek}"
		get("/name/10/foobar1234barh-buzz!-kek", "DataGetFormat",
			"10", "foobar", "barh", "buzz", "-kek"),
		get("/testObjectQueryParameter", "TestObjectQueryParameter"),
		post("/testFloatValidation", "TestFloatValidation"),
		get("/test/header", "GetHeader"),

		get("/test", ""),
		post("/test", ""),
		post("/pet/friendNames/10", "").defined(),
		post("/pet/aboba", "").defined(),
		del("/foobar", "").defined(),
		del("/name/10/foobar1234barh-buzz!-kek", "").defined(),
		post("/test/header", "").defined(),
	}
}

func TestRouter(t *testing.T) {
	handler := &sampleAPIServer{}
	s, err := api.NewServer(handler, handler)
	require.NoError(t, err)

	for i, tc := range routerTestCases() {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			t.Run("FindRoute", func(t *testing.T) {
				a := require.New(t)
				r, ok := s.FindRoute(tc.Method, tc.Path)
				if tc.Operation == "" {
					a.False(ok, r.OperationID())
					return
				}
				a.True(ok, tc.Operation)
				a.Equal(tc.Operation, r.OperationID())
				a.Equal(tc.Args, r.Args())
			})

			if tc.Operation == "" {
				t.Run("ServeHTTP", func(t *testing.T) {
					code := http.StatusNotFound
					if tc.Defined {
						code = http.StatusMethodNotAllowed
					}
					require.HTTPStatusCode(t, s.ServeHTTP, tc.Method, tc.Path, nil, code)
				})
			}
		})
	}
}
