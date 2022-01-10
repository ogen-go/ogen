package ogen

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
)

func TestRouter(t *testing.T) {
	s := api.NewServer(&sampleAPIServer{})

	type testCase struct {
		Method    string
		Path      string
		Operation string
	}
	test := func(m, p, op string) testCase {
		return testCase{
			Method:    m,
			Path:      p,
			Operation: op,
		}
	}
	get := func(p, op string) testCase {
		return test(http.MethodGet, p, op)
	}
	post := func(p, op string) testCase {
		return test(http.MethodPost, p, op)
	}
	put := func(p, op string) testCase {
		return test(http.MethodPut, p, op)
	}

	for i, tc := range []testCase{
		get("/pet/name/10", "PetNameByID"),
		get("/pet/friendNames/10", "PetFriendsNamesByID"),
		get("/pet", "PetGet"),
		get("/pet/avatar", "PetGetAvatarByID"),
		post("/pet/avatar", "PetUploadAvatarByID"),
		get("/pet/name", "PetGetByName"),
		get("/foobar", "FoobarGet"),
		post("/foobar", "FoobarPost"),
		put("/foobar", "FoobarPut"),
		get("/error", "ErrorGet"),
		get("/test/header", "GetHeader"),
		get("/name/10/foobar1234barh-buzz!-kek", "DataGetFormat"),
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
		})
	}
}
