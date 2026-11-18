package integration_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_nullable_empty_object"
)

// echoExtensions returns the received body as is.
type echoExtensions struct{}

func (echoExtensions) PatchExtensions(_ context.Context, req *api.Extensions) (*api.Extensions, error) {
	return req, nil
}

func decodeExtensions(t *testing.T, input string) api.Extensions {
	t.Helper()

	var got api.Extensions
	require.NoError(t, got.Decode(jx.DecodeStr(input)))
	return got
}

func encodeExtensions(t *testing.T, v *api.Extensions) string {
	t.Helper()

	var e jx.Encoder
	v.Encode(&e)
	return e.String()
}

// TestNullableEmptyObject checks that objects with no properties keep the
// difference between null, an empty object and an absent field.
func TestNullableEmptyObject(t *testing.T) {
	t.Run("DecodeNull", func(t *testing.T) {
		got := decodeExtensions(t, `{
			"optional": null,
			"required": null,
			"requiredLegacy": null,
			"optionalNoProperties": null,
			"notNullable": {}
		}`)

		require.True(t, got.Optional.Set)
		require.True(t, got.Optional.IsNull())
		require.True(t, got.Required.IsNull())
		require.True(t, got.RequiredLegacy.IsNull())
		require.True(t, got.OptionalNoProperties.Set)
		require.True(t, got.OptionalNoProperties.IsNull())
		require.NotNil(t, got.NotNullable)
	})

	t.Run("DecodeEmptyObject", func(t *testing.T) {
		got := decodeExtensions(t, `{
			"optional": {},
			"required": {},
			"requiredLegacy": {},
			"optionalNoProperties": {},
			"notNullable": {}
		}`)

		v, ok := got.Optional.Get()
		require.True(t, ok)
		require.Equal(t, api.ExtensionsOptional{}, v)
		require.True(t, got.Optional.Set)
		require.False(t, got.Optional.IsNull())

		_, ok = got.Required.Get()
		require.True(t, ok)
		require.False(t, got.Required.IsNull())

		_, ok = got.RequiredLegacy.Get()
		require.True(t, ok)

		_, ok = got.OptionalNoProperties.Get()
		require.True(t, ok)
	})

	t.Run("DecodeAbsent", func(t *testing.T) {
		got := decodeExtensions(t, `{"required": {}, "requiredLegacy": {}}`)

		require.False(t, got.Optional.Set)
		require.False(t, got.Optional.IsNull())
		_, ok := got.Optional.Get()
		require.False(t, ok)
		require.False(t, got.OptionalNoProperties.Set)
	})

	t.Run("Encode", func(t *testing.T) {
		v := &api.Extensions{
			Optional:             api.NewOptNilExtensionsOptional(api.ExtensionsOptional{}),
			Required:             api.NewNilExtensionsRequired(api.ExtensionsRequired{}),
			OptionalNoProperties: api.OptNilExtensionsOptionalNoProperties{Set: true, Null: true},
			NotNullable:          &api.ExtensionsNotNullable{},
		}
		v.RequiredLegacy.SetToNull()

		require.JSONEq(t, `{
			"optional": {},
			"required": {},
			"requiredLegacy": null,
			"optionalNoProperties": null,
			"notNullable": {}
		}`, encodeExtensions(t, v))
	})

	t.Run("EncodeUnsetOptional", func(t *testing.T) {
		v := &api.Extensions{
			NotNullable: &api.ExtensionsNotNullable{},
		}
		v.Required.SetToNull()
		v.RequiredLegacy.SetToNull()

		require.JSONEq(t, `{
			"required": null,
			"requiredLegacy": null,
			"notNullable": {}
		}`, encodeExtensions(t, v))
	})

	t.Run("RoundTrip", func(t *testing.T) {
		srv, err := api.NewServer(echoExtensions{})
		require.NoError(t, err)

		s := httptest.NewServer(srv)
		t.Cleanup(s.Close)

		client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
		require.NoError(t, err)

		in := &api.Extensions{
			Optional:    api.OptNilExtensionsOptional{Set: true, Null: true},
			Required:    api.NewNilExtensionsRequired(api.ExtensionsRequired{}),
			NotNullable: &api.ExtensionsNotNullable{},
		}
		in.RequiredLegacy.SetToNull()

		out, err := client.PatchExtensions(context.Background(), in)
		require.NoError(t, err)
		require.Equal(t, in, out)
	})
}
