package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_sse"
	"github.com/ogen-go/ogen/sse"
)

// sseHandler implements api.Handler, streaming events from the configured
// callbacks. A nil callback sends an empty stream.
type sseHandler struct {
	dataOnly          func(ctx context.Context, s api.DataOnlyOKSender) error
	dataOnlyKeepAlive func(ctx context.Context, s api.DataOnlyOKSender) error
	dataOnlyString    func(ctx context.Context, s api.DataOnlyStringOKSender) error
	dataOnlyTime      func(ctx context.Context, s api.DataOnlyTimeOKSender) error
	fullEvents        func(ctx context.Context, s api.FullEventsOKSender) error
	optionalStream    func(ctx context.Context, s api.OptionalStreamOKSender) error
	withHeaders       func(ctx context.Context, s api.WithHeadersOKSender) error
}

func (h sseHandler) DataOnly(context.Context) (*api.DataOnlyOK, error) {
	return &api.DataOnlyOK{Events: h.dataOnly, KeepAlive: h.dataOnlyKeepAlive}, nil
}

func (h sseHandler) DataOnlyString(context.Context) (*api.DataOnlyStringOK, error) {
	return &api.DataOnlyStringOK{Events: h.dataOnlyString}, nil
}

func (h sseHandler) DataOnlyTime(context.Context) (*api.DataOnlyTimeOK, error) {
	return &api.DataOnlyTimeOK{Events: h.dataOnlyTime}, nil
}

func (h sseHandler) FullEvents(context.Context) (*api.FullEventsOK, error) {
	return &api.FullEventsOK{Events: h.fullEvents}, nil
}

func (h sseHandler) OptionalStream(_ context.Context, params api.OptionalStreamParams) (api.OptionalStreamRes, error) {
	if params.Fail.Or(false) {
		return &api.Error{Message: "stream is not available"}, nil
	}
	return &api.OptionalStreamOK{Events: h.optionalStream}, nil
}

func (h sseHandler) WithHeaders(context.Context) (*api.WithHeadersOKHeaders, error) {
	return &api.WithHeadersOKHeaders{
		XStreamID: "stream-1",
		Response:  &api.WithHeadersOK{Events: h.withHeaders},
	}, nil
}

func testSSEServer(t *testing.T, h sseHandler) (client *api.Client, url string) {
	t.Helper()

	srv, err := api.NewServer(h)
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	t.Cleanup(s.Close)

	client, err = api.NewClient(s.URL, api.WithClient(s.Client()))
	require.NoError(t, err)
	return client, s.URL
}

// readRawStream reads the raw response body of an SSE endpoint.
func readRawStream(t *testing.T, url string) (resp *http.Response, body string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, string(data)
}

func TestSSEServerDataOnly(t *testing.T) {
	messages := []api.Message{
		{Text: "one", Count: api.NewOptInt(1)},
		{Text: "two"},
	}
	handler := sseHandler{
		dataOnly: func(ctx context.Context, s api.DataOnlyOKSender) error {
			if err := s.Comment(ctx, "keep-alive"); err != nil {
				return err
			}
			for i, msg := range messages {
				event := api.DataOnlyOKEvent{
					ID:   string(rune('a' + i)),
					Type: "update",
					Data: msg,
				}
				if i == 0 {
					event.Retry = api.NewOptInt(1000)
				}
				if err := s.Send(ctx, event); err != nil {
					return err
				}
			}
			return nil
		},
	}
	client, url := testSSEServer(t, handler)

	t.Run("Wire", func(t *testing.T) {
		resp, body := readRawStream(t, url+"/data-only")
		require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
		require.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
		require.Equal(t, ":keep-alive\n"+
			"id: a\nevent: update\nretry: 1000\ndata: {\"text\":\"one\",\"count\":1}\n\n"+
			"id: b\nevent: update\ndata: {\"text\":\"two\"}\n\n", body)
	})
	t.Run("Client", func(t *testing.T) {
		stream, err := client.DataOnly(t.Context())
		require.NoError(t, err)
		defer func() {
			require.NoError(t, stream.Close())
		}()

		// Note that reading past the last event reconnects the stream, as
		// required by the SSE specification.
		got := make([]api.DataOnlyOKEvent, 0, len(messages))
		for range messages {
			event, err := stream.Next(t.Context())
			require.NoError(t, err)
			got = append(got, event)
		}
		require.Equal(t, []api.DataOnlyOKEvent{
			{
				ID:    "a",
				Type:  "update",
				Data:  messages[0],
				Retry: api.NewOptInt(1000),
			},
			{
				ID:   "b",
				Type: "update",
				Data: messages[1],
			},
		}, got)
	})
}

func TestSSEServerStringData(t *testing.T) {
	handler := sseHandler{
		dataOnlyString: func(ctx context.Context, s api.DataOnlyStringOKSender) error {
			// Multi-line data must be split into multiple data fields.
			return s.Send(ctx, api.DataOnlyStringOKEvent{Data: "one\ntwo"})
		},
		dataOnlyTime: func(ctx context.Context, s api.DataOnlyTimeOKSender) error {
			return s.Send(ctx, api.DataOnlyTimeOKEvent{
				Data: time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC),
			})
		},
	}
	client, url := testSSEServer(t, handler)

	t.Run("String", func(t *testing.T) {
		_, body := readRawStream(t, url+"/data-only-string")
		require.Equal(t, "data: one\ndata: two\n\n", body)

		stream, err := client.DataOnlyString(t.Context())
		require.NoError(t, err)
		defer func() {
			require.NoError(t, stream.Close())
		}()

		event, err := stream.Next(t.Context())
		require.NoError(t, err)
		require.Equal(t, "one\ntwo", event.Data)
		require.Equal(t, sse.DefaultEventType, event.Type)
	})
	t.Run("DateTime", func(t *testing.T) {
		// String data is sent as-is, without JSON quoting.
		_, body := readRawStream(t, url+"/data-only-time")
		require.Equal(t, "data: 2026-07-22T10:00:00Z\n\n", body)

		stream, err := client.DataOnlyTime(t.Context())
		require.NoError(t, err)
		defer func() {
			require.NoError(t, stream.Close())
		}()

		event, err := stream.Next(t.Context())
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC), event.Data)
	})
}

func TestSSEServerFullShape(t *testing.T) {
	events := []api.FullEventsOKEvent{
		api.NewEventAFullEventsOKEvent(api.EventA{
			Event: api.EventAEventEventA,
			ID:    "1",
			Data:  api.Message{Text: "one"},
		}),
		api.NewEventBFullEventsOKEvent(api.EventB{
			Event: api.EventBEventEventB,
			ID:    "2",
			Data:  "plain text",
		}),
	}
	handler := sseHandler{
		fullEvents: func(ctx context.Context, s api.FullEventsOKSender) error {
			for _, event := range events {
				if err := s.Send(ctx, event); err != nil {
					return err
				}
			}
			return nil
		},
	}
	client, url := testSSEServer(t, handler)

	t.Run("Wire", func(t *testing.T) {
		// Discriminator is written as the event field, and string data of
		// EventB is sent unquoted.
		_, body := readRawStream(t, url+"/full")
		require.Equal(t, "id: 1\nevent: event_a\ndata: {\"text\":\"one\"}\n\n"+
			"id: 2\nevent: event_b\ndata: plain text\n\n", body)
	})
	t.Run("Client", func(t *testing.T) {
		stream, err := client.FullEvents(t.Context())
		require.NoError(t, err)
		defer func() {
			require.NoError(t, stream.Close())
		}()

		got := make([]api.FullEventsOKEvent, 0, len(events))
		for range events {
			event, err := stream.Next(t.Context())
			require.NoError(t, err)
			got = append(got, event)
		}
		require.Equal(t, events, got)
	})
}

func TestSSEServerResponseHeaders(t *testing.T) {
	handler := sseHandler{
		withHeaders: func(ctx context.Context, s api.WithHeadersOKSender) error {
			return s.Send(ctx, api.WithHeadersOKEvent{Data: api.Message{Text: "one"}})
		},
	}
	client, url := testSSEServer(t, handler)

	resp, body := readRawStream(t, url+"/with-headers")
	require.Equal(t, "stream-1", resp.Header.Get("X-Stream-Id"))
	require.Equal(t, "data: {\"text\":\"one\"}\n\n", body)

	res, err := client.WithHeaders(t.Context())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, res.Response.Close())
	}()
	require.Equal(t, "stream-1", res.XStreamID)

	event, err := res.Response.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, "one", event.Data.Text)
}

func TestSSEServerEmptyStream(t *testing.T) {
	// Handler returned a stream without events.
	_, url := testSSEServer(t, sseHandler{})

	resp, body := readRawStream(t, url+"/data-only")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	require.Empty(t, body)
}

func TestSSEServerNonStreamResponse(t *testing.T) {
	client, _ := testSSEServer(t, sseHandler{
		optionalStream: func(ctx context.Context, s api.OptionalStreamOKSender) error {
			return s.Send(ctx, api.OptionalStreamOKEvent{Data: api.Message{Text: "one"}})
		},
	})

	t.Run("Stream", func(t *testing.T) {
		res, err := client.OptionalStream(t.Context(), api.OptionalStreamParams{})
		require.NoError(t, err)

		stream, ok := res.(*api.OptionalStreamOK)
		require.True(t, ok, "unexpected response type %T", res)
		defer func() {
			require.NoError(t, stream.Close())
		}()

		event, err := stream.Next(t.Context())
		require.NoError(t, err)
		require.Equal(t, "one", event.Data.Text)
	})
	t.Run("Error", func(t *testing.T) {
		res, err := client.OptionalStream(t.Context(), api.OptionalStreamParams{
			Fail: api.NewOptBool(true),
		})
		require.NoError(t, err)
		require.Equal(t, &api.Error{Message: "stream is not available"}, res)
	})
}

func TestSSEServerHandlerError(t *testing.T) {
	// The response is already sent, so an error returned from Events must only
	// stop the stream, without writing an error response into it.
	var handled error
	srv, err := api.NewServer(sseHandler{
		dataOnly: func(ctx context.Context, s api.DataOnlyOKSender) error {
			if err := s.Send(ctx, api.DataOnlyOKEvent{Data: api.Message{Text: "one"}}); err != nil {
				return err
			}
			return errors.New("stream failed")
		},
	}, api.WithErrorHandler(func(_ context.Context, w http.ResponseWriter, _ *http.Request, err error) {
		handled = err
	}))
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	t.Cleanup(s.Close)

	resp, body := readRawStream(t, s.URL+"/data-only")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "data: {\"text\":\"one\"}\n\n", body)
	require.NoError(t, handled, "error handler must not be called for a sent response")
}

func TestSSEServerClientDisconnect(t *testing.T) {
	var (
		streamDone = make(chan error, 1)
		sent       = make(chan struct{})
	)
	client, _ := testSSEServer(t, sseHandler{
		dataOnly: func(ctx context.Context, s api.DataOnlyOKSender) error {
			defer close(streamDone)
			for i := 0; ; i++ {
				err := s.Send(ctx, api.DataOnlyOKEvent{Data: api.Message{Text: "tick"}})
				if err != nil {
					streamDone <- err
					return err
				}
				if i == 0 {
					close(sent)
				}
				select {
				case <-ctx.Done():
					streamDone <- ctx.Err()
					return ctx.Err()
				case <-time.After(time.Millisecond):
				}
			}
		},
	})

	ctx, cancel := context.WithCancel(t.Context())
	stream, err := client.DataOnly(ctx)
	require.NoError(t, err)

	event, err := stream.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "tick", event.Data.Text)

	<-sent
	// Disconnected client must stop the server stream.
	cancel()
	require.NoError(t, stream.Close())

	select {
	case err := <-streamDone:
		require.Error(t, err)
	case <-time.After(time.Minute):
		t.Fatal("server stream is still running")
	}
}

func TestSSEServerKeepAlive(t *testing.T) {
	t.Run("Custom", func(t *testing.T) {
		// A custom KeepAlive writes comments interleaved with events. It must
		// run concurrently with Events and stop once Events returns.
		release := make(chan struct{})
		_, url := testSSEServer(t, sseHandler{
			dataOnly: func(ctx context.Context, s api.DataOnlyOKSender) error {
				// Give keep-alive time to emit before the single event.
				select {
				case <-release:
				case <-ctx.Done():
					return ctx.Err()
				}
				return s.Send(ctx, api.DataOnlyOKEvent{Data: api.Message{Text: "one"}})
			},
			dataOnlyKeepAlive: func(ctx context.Context, s api.DataOnlyOKSender) error {
				// Emit one keep-alive comment, then let Events proceed and
				// keep pinging until Events returns and cancels ctx.
				if err := s.Comment(ctx, "ping"); err != nil {
					return err
				}
				close(release)
				return sse.KeepAliveEvery(ctx, s, 5*time.Millisecond, "ping")
			},
		})

		resp, body := readRawStream(t, url+"/data-only")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, ":ping\n", "expected keep-alive comment")
		require.Contains(t, body, "data: {\"text\":\"one\"}\n\n", "expected event")
	})

	t.Run("Disabled", func(t *testing.T) {
		// A KeepAlive that returns immediately disables keep-alive entirely.
		_, url := testSSEServer(t, sseHandler{
			dataOnly: func(ctx context.Context, s api.DataOnlyOKSender) error {
				return s.Send(ctx, api.DataOnlyOKEvent{Data: api.Message{Text: "one"}})
			},
			dataOnlyKeepAlive: func(context.Context, api.DataOnlyOKSender) error {
				return nil
			},
		})

		_, body := readRawStream(t, url+"/data-only")
		require.Equal(t, "data: {\"text\":\"one\"}\n\n", body)
	})
}
