package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_additional_operations"
)

type additionalOperationsTestServer struct {
	api.UnimplementedHandler
}

func (s *additionalOperationsTestServer) Echo(ctx context.Context, req api.EchoReq) (r api.EchoOK, err error) {
	return api.EchoOK{Data: req}, nil
}

func TestAdditionalOperations(t *testing.T) {
	srv, err := api.NewServer(&additionalOperationsTestServer{})
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	defer s.Close()

	const text = "Hello, 世界"

	req, err := http.NewRequest("LINK", s.URL+"/echo", strings.NewReader(text))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.Client().Do(req)
	require.NoError(t, err)

	respText, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, text, string(respText))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	client, err := api.NewClient(s.URL)
	require.NoError(t, err)

	resp2, err := client.Echo(t.Context(), api.EchoReq{Data: strings.NewReader(text)})
	require.NoError(t, err)
	resp2Text, err := io.ReadAll(resp2)
	require.NoError(t, err)
	require.Equal(t, text, string(resp2Text))
}
