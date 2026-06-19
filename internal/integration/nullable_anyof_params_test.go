package integration_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_nullable_anyof_params"
)

// nullableAnyOfHandler records the last received parameters so the test can
// assert how nullable anyOf parameters are decoded.
type nullableAnyOfHandler struct {
	last api.ListDesktopsParams
}

func (h *nullableAnyOfHandler) ListDesktops(_ context.Context, params api.ListDesktopsParams) (api.DesktopImageType, error) {
	h.last = params
	return api.DesktopImageTypeStock, nil
}

// TestNullableAnyOfParams verifies that parameters made nullable via
// anyOf: [X, {"type": "null"}] (OpenAPI 3.1 style) generate
// OptNil types and round-trip correctly for both the $ref and the scalar
// branch, in query and header locations.
func TestNullableAnyOfParams(t *testing.T) {
	h := &nullableAnyOfHandler{}
	srv, err := api.NewServer(h)
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	t.Cleanup(s.Close)

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	require.NoError(t, err)

	t.Run("Set", func(t *testing.T) {
		_, err := client.ListDesktops(context.Background(), api.ListDesktopsParams{
			ImageType: api.NewOptNilDesktopImageType(api.DesktopImageTypeUser),
			DesktopID: api.NewOptNilString("d-1"),
			XTraceID:  api.NewOptNilString("trace-1"),
		})
		require.NoError(t, err)

		imageType, ok := h.last.ImageType.Get()
		require.True(t, ok, "image_type must be present")
		require.Equal(t, api.DesktopImageTypeUser, imageType)

		desktopID, ok := h.last.DesktopID.Get()
		require.True(t, ok, "desktop_id must be present")
		require.Equal(t, "d-1", desktopID)

		traceID, ok := h.last.XTraceID.Get()
		require.True(t, ok, "X-Trace-Id must be present")
		require.Equal(t, "trace-1", traceID)
	})

	t.Run("Unset", func(t *testing.T) {
		_, err := client.ListDesktops(context.Background(), api.ListDesktopsParams{})
		require.NoError(t, err)

		require.False(t, h.last.ImageType.IsSet(), "image_type must be unset")
		require.False(t, h.last.DesktopID.IsSet(), "desktop_id must be unset")
		require.False(t, h.last.XTraceID.IsSet(), "X-Trace-Id must be unset")
	})

	// A client-side null collapses to an absent parameter on the wire: form
	// serialization has no representation for JSON null, so OptNil.Get reports
	// ok=false and the encoder writes nothing. The server therefore decodes it
	// as unset. This is ogen's standard nullable query/header behavior, shared
	// with every other nullable parameter form (e.g. allOf + nullable); the
	// anyOf-null spec we added must behave identically.
	t.Run("NullCollapsesToUnset", func(t *testing.T) {
		var imageType api.OptNilDesktopImageType
		imageType.SetToNull()
		var traceID api.OptNilString
		traceID.SetToNull()

		_, err := client.ListDesktops(context.Background(), api.ListDesktopsParams{
			ImageType: imageType,
			XTraceID:  traceID,
		})
		require.NoError(t, err)

		require.False(t, h.last.ImageType.IsSet(), "null image_type is sent as absent")
		require.False(t, h.last.XTraceID.IsSet(), "null X-Trace-Id is sent as absent")
	})
}
