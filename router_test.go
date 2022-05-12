package ogen

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

func TestRouter(t *testing.T) {
	handler := &sampleAPIServer{}
	s, err := api.NewServer(handler, handler)
	require.NoError(t, err)

	type testCase struct {
		Method    string
		Path      string
		Operation string
		Args      []string
	}
	test := func(method, route, op string, args ...string) testCase {
		if len(args) == 0 {
			args = []string{}
		}
		return testCase{
			Method:    method,
			Path:      route,
			Operation: op,
			Args:      args,
		}
	}
	get := func(p, op string, args ...string) testCase {
		return test(http.MethodGet, p, op, args...)
	}
	post := func(p, op string, args ...string) testCase {
		return test(http.MethodPost, p, op, args...)
	}
	put := func(p, op string, args ...string) testCase {
		return test(http.MethodPut, p, op, args...)
	}

	for i, tc := range []testCase{
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
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
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
	}
}
