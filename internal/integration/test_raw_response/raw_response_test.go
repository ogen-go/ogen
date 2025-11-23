package api

import (
	"cmp"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testHandler implements both Handler and RawHandler interfaces for testing
type testHandler struct {
	UnimplementedHandler
}

// Normal handler - returns structured JSON response
func (h *testHandler) GetNormalData(ctx context.Context) (*GetNormalDataOK, error) {
	return &GetNormalDataOK{
		Message: NewOptString("normal response"),
	}, nil
}

// Raw handler - writes raw response directly to ResponseWriter
func (h *testHandler) GetRawData(ctx context.Context, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	rc := http.NewResponseController(w)
	_, writeErr := w.Write([]byte(`{"data": "raw response from server"}`))
	flushErr := rc.Flush()
	return cmp.Or(writeErr, flushErr)
}

// Mixed handler - writes raw response for octet-stream content type
func (h *testHandler) GetMixedData(ctx context.Context, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(200)
	rc := http.NewResponseController(w)
	_, writeErr := w.Write([]byte("binary data from server"))
	flushErr := rc.Flush()
	return cmp.Or(writeErr, flushErr)
}

func TestRawResponse(t *testing.T) {
	// Create handler
	handler := &testHandler{}

	// Create server with both handler interfaces
	srv, err := NewServer(handler, handler)
	require.NoError(t, err)

	// Create test server
	httpServer := httptest.NewServer(srv)
	defer httpServer.Close()

	// Create client
	client, err := NewClient(httpServer.URL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("NormalData", func(t *testing.T) {
		// Test normal structured response
		resp, err := client.GetNormalData(ctx)
		require.NoError(t, err)
		message, ok := resp.Message.Get()
		require.True(t, ok)
		assert.Equal(t, "normal response", message)
	})

	t.Run("RawData", func(t *testing.T) {
		// Test raw response - should return http.Response directly
		resp, err := client.GetRawData(ctx)
		require.NoError(t, err)

		// The response should be a raw response type
		rawResp, ok := resp.(*GetRawDataOKRawApplicationJSON)
		require.True(t, ok, "Expected GetRawDataOKRawApplicationJSON, got %T", resp)
		defer rawResp.Response.Body.Close()

		// Verify headers and status
		assert.Equal(t, 200, rawResp.Response.StatusCode)
		assert.Equal(t, "application/json", rawResp.Response.Header.Get("Content-Type"))

		// Verify body content
		body, err := io.ReadAll(rawResp.Response.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"data": "raw response from server"}`, string(body))
	})

	t.Run("MixedData", func(t *testing.T) {
		// Test mixed response - should return raw response for octet-stream
		resp, err := client.GetMixedData(ctx)
		require.NoError(t, err)

		// The response should be a raw response type for octet-stream
		rawResp, ok := resp.(*GetMixedDataOKRawApplicationOctetStream)
		require.True(t, ok, "Expected GetMixedDataOKRawApplicationOctetStream, got %T", resp)
		defer rawResp.Response.Body.Close()

		// Verify headers and status
		assert.Equal(t, 200, rawResp.Response.StatusCode)
		assert.Equal(t, "application/octet-stream", rawResp.Response.Header.Get("Content-Type"))

		// Verify body content
		body, err := io.ReadAll(rawResp.Response.Body)
		require.NoError(t, err)
		assert.Equal(t, "binary data from server", string(body))
	})
}
