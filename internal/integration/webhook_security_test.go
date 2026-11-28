package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_webhook_security"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"
)

const webhookSignatureHeader = "X-Signature"

type testWebhookSecurity struct {
	token     string
	signature string
}

func (t *testWebhookSecurity) GetStatus(ctx context.Context) error {
	return nil
}

func (t *testWebhookSecurity) EventWebhook(ctx context.Context, req *api.Event) error {
	return nil
}

func (t *testWebhookSecurity) SignedWebhook(ctx context.Context, req *api.Event) error {
	return nil
}

func (t *testWebhookSecurity) HandleTokenHeader(ctx context.Context, operationName api.OperationName, v api.TokenHeader) (context.Context, error) {
	if v.APIKey != t.token {
		return nil, errors.Errorf("invalid token: %q", v.APIKey)
	}
	return ctx, nil
}

func (t *testWebhookSecurity) HandleSignature(ctx context.Context, operationName api.OperationName, v api.Signature) (context.Context, error) {
	if got := v.Request.Header.Get(webhookSignatureHeader); got != t.signature {
		return nil, errors.Errorf("invalid signature: %q", got)
	}
	return ctx, nil
}

type testWebhookSecuritySource struct {
	token     string
	signature string
}

func (t testWebhookSecuritySource) TokenHeader(ctx context.Context, operationName api.OperationName) (r api.TokenHeader, _ error) {
	if t.token == "" {
		return r, ogenerrors.ErrSkipClientSecurity
	}
	return api.TokenHeader{APIKey: t.token}, nil
}

func (t testWebhookSecuritySource) Signature(ctx context.Context, operationName api.OperationName, req *http.Request) error {
	if t.signature == "" {
		return ogenerrors.ErrSkipClientSecurity
	}
	req.Header.Set(webhookSignatureHeader, t.signature)
	return nil
}

func TestWebhookSecurity(t *testing.T) {
	h := &testWebhookSecurity{
		token:     "token",
		signature: "signature",
	}

	srv, err := api.NewServer(h, h)
	require.NoError(t, err)
	whSrv, err := api.NewWebhookServer(h, h)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/status", srv)
	mux.Handle("/event", whSrv.Handler("event"))
	mux.Handle("/signed", whSrv.Handler("signed"))
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)

	ctx := context.Background()
	event := &api.Event{ID: "1"}

	requireUnauthorized := func(t *testing.T, err error) {
		t.Helper()
		var statusErr *validate.UnexpectedStatusCodeError
		require.ErrorAs(t, err, &statusErr)
		require.Equal(t, http.StatusUnauthorized, statusErr.StatusCode)
	}

	t.Run("Valid", func(t *testing.T) {
		src := testWebhookSecuritySource{token: h.token, signature: h.signature}

		client, err := api.NewClient(s.URL, src, api.WithClient(s.Client()))
		require.NoError(t, err)
		require.NoError(t, client.GetStatus(ctx))

		whClient, err := api.NewWebhookClient(src, api.WithClient(s.Client()))
		require.NoError(t, err)
		require.NoError(t, whClient.EventWebhook(ctx, s.URL+"/event", event))
		require.NoError(t, whClient.SignedWebhook(ctx, s.URL+"/signed", event))
	})
	t.Run("Invalid", func(t *testing.T) {
		whClient, err := api.NewWebhookClient(testWebhookSecuritySource{
			token:     "wrong",
			signature: "wrong",
		}, api.WithClient(s.Client()))
		require.NoError(t, err)
		requireUnauthorized(t, whClient.EventWebhook(ctx, s.URL+"/event", event))
		requireUnauthorized(t, whClient.SignedWebhook(ctx, s.URL+"/signed", event))
	})
	t.Run("Missing", func(t *testing.T) {
		whClient, err := api.NewWebhookClient(testWebhookSecuritySource{}, api.WithClient(s.Client()))
		require.NoError(t, err)
		require.ErrorIs(t, whClient.EventWebhook(ctx, s.URL+"/event", event), ogenerrors.ErrSecurityRequirementIsNotSatisfied)
		require.ErrorIs(t, whClient.SignedWebhook(ctx, s.URL+"/signed", event), ogenerrors.ErrSecurityRequirementIsNotSatisfied)

		for _, path := range []string{"/event", "/signed"} {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+path, strings.NewReader(`{"id":"1"}`))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp, err := s.Client().Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
