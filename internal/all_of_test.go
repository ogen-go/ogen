package internal

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/test_allof"
)

type allofTestServer struct {
	api.UnimplementedHandler
}

func (s *allofTestServer) NullableStrings(ctx context.Context, req string) error {
	return nil
}

func (s *allofTestServer) SimpleInteger(ctx context.Context, req int) error {
	return nil
}

func (s *allofTestServer) ObjectsWithConflictingProperties(ctx context.Context, req api.ObjectsWithConflictingPropertiesReq) error {
	return nil
}

func (s *allofTestServer) ObjectsWithConflictingArrayProperty(ctx context.Context, req api.ObjectsWithConflictingArrayPropertyReq) error {
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

	ctx := context.Background()
	t.Run("nullableStrings", func(t *testing.T) {
		err := client.NullableStrings(ctx, "foo")
		require.EqualError(t, err, "validate: string: no regex match")

		err = client.NullableStrings(ctx, "127.0.0.1")
		require.NoError(t, err)
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
		err := client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Foo: "1234567890",
		})
		require.NoError(t, err)

		err = client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Bar: api.NewOptInt(1337),
			Foo: "1234567890",
		})
		require.EqualError(t, err, "validate: invalid: bar (int: value 1337 greater than 255)")

		err = client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Bar: api.NewOptInt(255),
			Foo: "1234567890",
		})
		require.NoError(t, err)

		err = client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Foo: "123456",
		})
		require.EqualError(t, err, "validate: invalid: foo (string: len 6 less than minimum 10)")
	})
	t.Run("objectsWithConflictingArrayProperty", func(t *testing.T) {
		err := client.ObjectsWithConflictingArrayProperty(ctx, api.ObjectsWithConflictingArrayPropertyReq{
			Foo: []int{},
			Bar: 5,
		})
		require.EqualError(t, err, "validate: invalid: foo (array: len 0 less than minimum 1)")

		err = client.ObjectsWithConflictingArrayProperty(ctx, api.ObjectsWithConflictingArrayPropertyReq{
			Foo: []int{1, 2, 3, 4},
			Bar: 5,
		})
		require.NoError(t, err)

		err = client.ObjectsWithConflictingArrayProperty(ctx, api.ObjectsWithConflictingArrayPropertyReq{
			Foo: []int{1, 2, 3, 4, 5, 6},
			Bar: 5,
		})
		require.EqualError(t, err, "validate: invalid: foo (array: len 6 greater than maximum 5)")
	})
}
