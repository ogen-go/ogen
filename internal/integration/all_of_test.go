package integration

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_allof"
	"github.com/ogen-go/ogen/validate"
)

type allofTestServer struct {
	api.UnimplementedHandler
}

func (s *allofTestServer) NullableStrings(ctx context.Context, req api.NilString) error {
	return nil
}

func (s *allofTestServer) SimpleInteger(ctx context.Context, req int) error {
	return nil
}

func (s *allofTestServer) ObjectsWithConflictingProperties(ctx context.Context, req *api.ObjectsWithConflictingPropertiesReq) error {
	return nil
}

func (s *allofTestServer) ObjectsWithConflictingArrayProperty(ctx context.Context, req *api.ObjectsWithConflictingArrayPropertyReq) error {
	return nil
}

func (s *allofTestServer) StringsNotype(ctx context.Context, req api.NilString) error {
	return nil
}

func TestAllof(t *testing.T) {
	var client *api.Client
	{
		srv, err := api.NewServer(&allofTestServer{})
		require.NoError(t, err)

		s := httptest.NewServer(srv)
		defer s.Close()

		client, err = api.NewClient(s.URL, api.WithClient(s.Client()))
		require.NoError(t, err)
	}

	regexPattern := "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$"

	ctx := context.Background()
	t.Run("nullableStrings", func(t *testing.T) {
		err := client.NullableStrings(ctx, api.NilString{})
		regexMatchErr, ok := errors.Into[*validate.NoRegexMatchError](err)
		require.True(t, ok, "validate: string: no regex match")
		require.ErrorContains(
			t,
			regexMatchErr,
			fmt.Sprintf("no regex match: %s", regexPattern),
		)

		err = client.NullableStrings(ctx, api.NewNilString("foo"))
		regexMatchErr, ok = errors.Into[*validate.NoRegexMatchError](err)
		require.True(t, ok, "validate: string: no regex match")
		require.ErrorContains(
			t,
			regexMatchErr,
			fmt.Sprintf("no regex match: %s", regexPattern),
		)

		err = client.NullableStrings(ctx, api.NewNilString("127.0.0.1"))
		require.NoError(t, err)
	})
	t.Run("stringsNotype", func(t *testing.T) {
		err := client.StringsNotype(ctx, api.NilString{})
		require.NoError(t, err)

		err = client.StringsNotype(ctx, api.NewNilString("foo"))
		require.NoError(t, err)

		err = client.StringsNotype(ctx, api.NewNilString(strings.Repeat("1", 16)))
		require.EqualError(t, err, "validate: string: len 16 greater than maximum 15")
	})
	t.Run("simpleInteger", func(t *testing.T) {
		err := client.SimpleInteger(ctx, -7)
		require.EqualError(t, err, "validate: int: value -7 less than -5")

		err = client.SimpleInteger(ctx, -5)
		require.NoError(t, err)

		err = client.SimpleInteger(ctx, 5)
		require.NoError(t, err)

		err = client.SimpleInteger(ctx, 10)
		require.EqualError(t, err, "validate: int: value 10 greater than 5")
	})
	t.Run("objectsWithConflictingProperties", func(t *testing.T) {
		err := client.ObjectsWithConflictingProperties(ctx, &api.ObjectsWithConflictingPropertiesReq{
			Foo: "1234567890",
		})
		require.NoError(t, err)

		err = client.ObjectsWithConflictingProperties(ctx, &api.ObjectsWithConflictingPropertiesReq{
			Bar: api.NewOptInt(1337),
			Foo: "1234567890",
		})
		require.EqualError(t, err, "validate: invalid: bar (int: value 1337 greater than 255)")

		err = client.ObjectsWithConflictingProperties(ctx, &api.ObjectsWithConflictingPropertiesReq{
			Bar: api.NewOptInt(255),
			Foo: "1234567890",
		})
		require.NoError(t, err)

		err = client.ObjectsWithConflictingProperties(ctx, &api.ObjectsWithConflictingPropertiesReq{
			Foo: "123456",
		})
		require.EqualError(t, err, "validate: invalid: foo (string: len 6 less than minimum 10)")
	})
	t.Run("objectsWithConflictingArrayProperty", func(t *testing.T) {
		err := client.ObjectsWithConflictingArrayProperty(ctx, &api.ObjectsWithConflictingArrayPropertyReq{
			Foo: []int{},
			Bar: 5,
		})
		require.EqualError(t, err, "validate: invalid: foo (array: len 0 less than minimum 1)")

		err = client.ObjectsWithConflictingArrayProperty(ctx, &api.ObjectsWithConflictingArrayPropertyReq{
			Foo: []int{1, 2, 3, 4},
			Bar: 5,
		})
		require.NoError(t, err)

		err = client.ObjectsWithConflictingArrayProperty(ctx, &api.ObjectsWithConflictingArrayPropertyReq{
			Foo: []int{1, 2, 3, 4, 5, 6},
			Bar: 5,
		})
		require.EqualError(t, err, "validate: invalid: foo (array: len 6 greater than maximum 5)")
	})
}
