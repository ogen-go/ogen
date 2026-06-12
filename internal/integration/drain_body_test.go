package integration

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/sample_api"
)

// trackingBody records whether it was read to EOF (drained) and closed.
type trackingBody struct {
	r       *strings.Reader
	drained bool
	closed  bool
}

func (b *trackingBody) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	if err == io.EOF {
		b.drained = true
	}
	return n, err
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// emptySecuritySource satisfies api.SecuritySource for operations without security.
type emptySecuritySource struct{}

func (emptySecuritySource) APIKey(context.Context, api.OperationName) (api.APIKey, error) {
	return api.APIKey{}, nil
}

// TestClientDrainsBodyOnUnexpectedContentType is a regression test for #1670:
// the generated client must read the response body to completion before closing
// it, even when the response can't be decoded (here, an unexpected Content-Type
// for a declared status code), so the net/http Transport can reuse the
// keep-alive TCP connection.
//
// The bare "unexpected status code" path is already drained by
// validate.UnexpectedStatusCodeWithResponse; the InvalidContentType path was
// not, which is the leak this exercises.
func TestClientDrainsBodyOnUnexpectedContentType(t *testing.T) {
	body := &trackingBody{r: strings.NewReader("upstream returned HTML error page")}

	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK, // Declared status code...
			Status:     "200 OK",
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			// ...but an unexpected Content-Type, so decoding fails.
			Header:  http.Header{"Content-Type": []string{"text/html"}},
			Body:    body,
			Request: r,
		}, nil
	})

	client, err := api.NewClient("https://example.com",
		emptySecuritySource{},
		api.WithClient(&http.Client{Transport: rt}),
	)
	require.NoError(t, err)

	_, err = client.NoAdditionalPropertiesTest(context.Background())
	require.Error(t, err, "undecodable response must be reported as an error")

	require.True(t, body.drained, "response body must be drained to EOF before close")
	require.True(t, body.closed, "response body must be closed")
}
