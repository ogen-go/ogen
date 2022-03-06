package ogen

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
)

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
	test := func(m, p, op string, args ...string) testCase {
		if len(args) == 0 {
			args = []string{}
		}
		return testCase{
			Method:    m,
			Path:      p,
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
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			a := require.New(t)
			r, ok := s.FindRoute(tc.Method, tc.Path)
			if tc.Operation == "" {
				a.False(ok)
				return
			}
			a.True(ok)
			a.Equal(tc.Operation, r.OperationID())
			a.Equal(tc.Args, r.Args())
		})
	}
}
