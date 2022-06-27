package ogen

import (
	"context"
	"net/http/httptest"
	"testing"

	api "github.com/ogen-go/ogen/internal/test_allof"
	"github.com/stretchr/testify/require"
)

type allofTestServer struct {
	api.UnimplementedHandler
}

func (s *allofTestServer) SimpleInteger(ctx context.Context, req int) (api.SimpleIntegerOK, error) {
	return api.SimpleIntegerOK{}, nil
}

func (s *allofTestServer) ObjectsWithConflictingProperties(ctx context.Context, req api.ObjectsWithConflictingPropertiesReq) (r api.ObjectsWithConflictingPropertiesOK, _ error) {
	return api.ObjectsWithConflictingPropertiesOK{}, nil
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
	t.Run("simpleInteger", func(t *testing.T) {
		_, err := client.SimpleInteger(ctx, -7)
		require.EqualError(t, err, "validate: int: value -7 less than -5")

		_, err = client.SimpleInteger(ctx, -5)
		require.NoError(t, err)

		_, err = client.SimpleInteger(ctx, 5)
		require.NoError(t, err)

		_, err = client.SimpleInteger(ctx, 10)
		require.EqualError(t, err, "validate: int: value 10 greater than 5")
	})

	t.Run("objectsWithConflictingProperties", func(t *testing.T) {
		_, err := client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Foo: "1234567890",
		})
		require.NoError(t, err)

		_, err = client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Bar: api.NewOptInt(1337),
			Foo: "1234567890",
		})
		require.EqualError(t, err, "validate: invalid: bar (int: value 1337 greater than 255)")

		_, err = client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Bar: api.NewOptInt(255),
			Foo: "1234567890",
		})
		require.NoError(t, err)

		_, err = client.ObjectsWithConflictingProperties(ctx, api.ObjectsWithConflictingPropertiesReq{
			Foo: "123456",
		})
		require.EqualError(t, err, "validate: invalid: foo (string: len 6 less than minimum 10)")
	})
}
