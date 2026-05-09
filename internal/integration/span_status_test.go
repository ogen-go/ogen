package integration_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	errapi "github.com/ogen-go/ogen/internal/integration/sample_err"
	spanapi "github.com/ogen-go/ogen/internal/integration/test_span_status"
	whapi "github.com/ogen-go/ogen/internal/integration/test_webhooks"
	"github.com/ogen-go/ogen/middleware"
)

type spanStatusHandler struct {
	spanapi.UnimplementedHandler
	bodyRes    spanapi.SpanStatusBodyRes
	noBodyRes  *spanapi.SpanStatusNoBodyDef
	middleware spanapi.Middleware
	apiKeyErr  error
}

var _ spanapi.Handler = (*spanStatusHandler)(nil)

func (h *spanStatusHandler) SpanStatusBody(context.Context) (spanapi.SpanStatusBodyRes, error) {
	return h.bodyRes, nil
}

func (h *spanStatusHandler) SpanStatusNoBody(context.Context) (*spanapi.SpanStatusNoBodyDef, error) {
	return h.noBodyRes, nil
}

func (h *spanStatusHandler) SpanStatusRequestChecks(context.Context, *spanapi.SpanStatusRequestChecksReq, spanapi.SpanStatusRequestChecksParams) error {
	return nil
}

func (h *spanStatusHandler) HandleAPIKey(ctx context.Context, _ spanapi.OperationName, _ spanapi.APIKey) (context.Context, error) {
	return ctx, h.apiKeyErr
}

func (h *spanStatusHandler) newServer(t testing.TB, tp trace.TracerProvider) http.Handler {
	t.Helper()

	opts := []spanapi.ServerOption{
		spanapi.WithTracerProvider(tp),
		spanapi.WithMeterProvider(metricnoop.NewMeterProvider()),
	}
	if h.middleware != nil {
		opts = append(opts, spanapi.WithMiddleware(h.middleware))
	}
	srv, err := spanapi.NewServer(h, h, opts...)
	require.NoError(t, err)

	return srv
}

type spanStatusUnimplementedHandler struct {
	spanapi.UnimplementedHandler
}

func (h *spanStatusUnimplementedHandler) HandleAPIKey(ctx context.Context, _ spanapi.OperationName, _ spanapi.APIKey) (context.Context, error) {
	return ctx, nil
}

func (h *spanStatusUnimplementedHandler) newServer(t testing.TB, tp trace.TracerProvider) http.Handler {
	t.Helper()

	srv, err := spanapi.NewServer(h, h,
		spanapi.WithTracerProvider(tp),
		spanapi.WithMeterProvider(metricnoop.NewMeterProvider()),
	)
	require.NoError(t, err)

	return srv
}

type spanStatusErrHandler struct {
	errapi.UnimplementedHandler
	dataGetErr  error
	newErrorRes *errapi.ErrorStatusCode
}

var _ errapi.Handler = (*spanStatusErrHandler)(nil)

func (h *spanStatusErrHandler) DataGet(context.Context) (*errapi.Data, error) {
	return nil, h.dataGetErr
}

func (h *spanStatusErrHandler) NewError(context.Context, error) *errapi.ErrorStatusCode {
	return h.newErrorRes
}

func (h *spanStatusErrHandler) newServer(t testing.TB, tp trace.TracerProvider) http.Handler {
	t.Helper()

	srv, err := errapi.NewServer(h,
		errapi.WithTracerProvider(tp),
		errapi.WithMeterProvider(metricnoop.NewMeterProvider()),
	)
	require.NoError(t, err)

	return srv
}

type spanStatusWebhookHandler struct {
	whapi.UnimplementedHandler
	statusCode int
}

var _ whapi.WebhookHandler = (*spanStatusWebhookHandler)(nil)

func (h *spanStatusWebhookHandler) StatusWebhook(context.Context) (*whapi.StatusWebhookOK, error) {
	return &whapi.StatusWebhookOK{
		Status: whapi.NewOptString("ok"),
	}, nil
}

func (h *spanStatusWebhookHandler) UpdateDelete(context.Context) (whapi.UpdateDeleteRes, error) {
	return &whapi.ErrorStatusCode{
		StatusCode: h.statusCode,
		Response:   whapi.Error{Error: "test error"},
	}, nil
}

func (h *spanStatusWebhookHandler) newServer(t testing.TB, tp trace.TracerProvider) *whapi.WebhookServer {
	t.Helper()

	srv, err := whapi.NewWebhookServer(h,
		whapi.WithTracerProvider(tp),
		whapi.WithMeterProvider(metricnoop.NewMeterProvider()),
	)
	require.NoError(t, err)

	return srv
}

type wantSpanStatus struct {
	code codes.Code
}

type spanStatusServerHandler interface {
	newServer(testing.TB, trace.TracerProvider) http.Handler
}

func newSpanStatusTrace(t *testing.T) (*tracetest.InMemoryExporter, *sdktrace.TracerProvider) {
	t.Helper()

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() {
		require.NoError(t, tp.Shutdown(context.Background()))
	})

	return exp, tp
}

// TestServerSpanStatus verifies that the span status set by generated response encoders
// complies with OTel HTTP Semantic Conventions.
//
// Spec: https://opentelemetry.io/docs/specs/semconv/http/http-spans/#status
//   - 1xx/2xx/3xx: leave span status Unset (server handled the request normally)
//   - 4xx:         leave span status Unset (client-side error, not a server failure)
//   - 5xx:         set span status to Error
func TestServerSpanStatus(t *testing.T) {
	tests := []struct {
		name           string
		handler        *spanStatusHandler
		method         string
		path           string
		wantHTTPStatus int
		want           wantSpanStatus
	}{
		{
			name: "2xx/fixed: SpanStatusBody returning 200 should leave span Unset",
			handler: &spanStatusHandler{
				bodyRes: &spanapi.SpanStatusBodyOK{Message: "ok"},
			},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusOK,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "dynamic status code default with body: SpanStatusBody returning 0 should write 200 and leave span Unset",
			handler: &spanStatusHandler{
				bodyRes: &spanapi.ErrorStatusCode{
					Response: spanapi.Error{Code: 0, Message: "test error"},
				},
			},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusOK,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "3xx/dynamic no-content: SpanStatusNoBody returning 302 should leave span Unset",
			handler: &spanStatusHandler{
				noBodyRes: &spanapi.SpanStatusNoBodyDef{
					StatusCode: http.StatusFound,
				},
			},
			method:         http.MethodPut,
			path:           "/span-status",
			wantHTTPStatus: http.StatusFound,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "4xx/fixed: SpanStatusBody returning 404 should leave span Unset",
			handler: &spanStatusHandler{
				bodyRes: &spanapi.SpanStatusBodyNotFound{},
			},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusNotFound,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "4xx/dynamic with body: SpanStatusBody returning 400 should leave span Unset",
			handler: &spanStatusHandler{
				bodyRes: &spanapi.ErrorStatusCode{
					StatusCode: http.StatusBadRequest,
					Response:   spanapi.Error{Code: 0, Message: "test error"},
				},
			},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusBadRequest,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "5xx/fixed: SpanStatusBody returning 500 should set span Error",
			handler: &spanStatusHandler{
				bodyRes: &spanapi.SpanStatusBodyInternalServerError{},
			},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
		{
			name: "5xx/dynamic no-content: SpanStatusNoBody returning 500 should set span Error",
			handler: &spanStatusHandler{
				noBodyRes: &spanapi.SpanStatusNoBodyDef{
					StatusCode: http.StatusInternalServerError,
				},
			},
			method:         http.MethodPut,
			path:           "/span-status",
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
		{
			name: "5xx/dynamic with body: SpanStatusBody returning 500 should set span Error",
			handler: &spanStatusHandler{
				bodyRes: &spanapi.ErrorStatusCode{
					StatusCode: http.StatusInternalServerError,
					Response:   spanapi.Error{Code: 0, Message: "test error"},
				},
			},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tp := newSpanStatusTrace(t)
			srv := tt.handler.newServer(t, tp)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			require.Equal(t, tt.wantHTTPStatus, rec.Code)

			requireServerSpanStatus(t, exp, tt.want)
		})
	}
}

func TestServerSpanStatusMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        *spanStatusHandler
		path           string
		wantHTTPStatus int
		want           wantSpanStatus
	}{
		{
			name: "Middleware returning error should set span Error",
			handler: &spanStatusHandler{
				middleware: func(middleware.Request, middleware.Next) (middleware.Response, error) {
					return middleware.Response{}, errors.New("middleware failed")
				},
			},
			path:           "/span-status",
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tp := newSpanStatusTrace(t)
			srv := tt.handler.newServer(t, tp)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			require.Equal(t, tt.wantHTTPStatus, rec.Code)

			requireServerSpanStatus(t, exp, tt.want)
		})
	}
}

func TestServerSpanStatusHandlerErrorPaths(t *testing.T) {
	tests := []struct {
		name           string
		handler        spanStatusServerHandler
		method         string
		path           string
		body           io.Reader
		apiKey         string
		wantHTTPStatus int
		want           wantSpanStatus
	}{
		{
			name:           "DecodeParams returning 400 should leave span Unset",
			handler:        &spanStatusHandler{},
			method:         http.MethodPost,
			path:           "/span-status?id=not-int",
			body:           strings.NewReader(`{"message":"ok"}`),
			apiKey:         "test-key",
			wantHTTPStatus: http.StatusBadRequest,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name:           "DecodeRequest returning 400 should leave span Unset",
			handler:        &spanStatusHandler{},
			method:         http.MethodPost,
			path:           "/span-status?id=1",
			body:           strings.NewReader(`{"broken":`),
			apiKey:         "test-key",
			wantHTTPStatus: http.StatusBadRequest,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "API key handler error returning 401 should leave span Unset",
			handler: &spanStatusHandler{
				apiKeyErr: errors.New("api key failed"),
			},
			method:         http.MethodPost,
			path:           "/span-status?id=1",
			body:           strings.NewReader(`{"message":"ok"}`),
			apiKey:         "test-key",
			wantHTTPStatus: http.StatusUnauthorized,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name:           "Unsatisfied security requirement returning 401 should leave span Unset",
			handler:        &spanStatusHandler{},
			method:         http.MethodPost,
			path:           "/span-status?id=1",
			body:           strings.NewReader(`{"message":"ok"}`),
			wantHTTPStatus: http.StatusUnauthorized,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name:           "Handler ErrNotImplemented returning 501 should set span Error",
			handler:        &spanStatusUnimplementedHandler{},
			method:         http.MethodGet,
			path:           "/span-status",
			wantHTTPStatus: http.StatusNotImplemented,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tp := newSpanStatusTrace(t)
			srv := tt.handler.newServer(t, tp)

			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			require.Equal(t, tt.wantHTTPStatus, rec.Code)

			requireServerSpanStatus(t, exp, tt.want)
		})
	}
}

func TestServerSpanStatusConvenientErrors(t *testing.T) {
	tests := []struct {
		name           string
		handler        *spanStatusErrHandler
		wantHTTPStatus int
		want           wantSpanStatus
	}{
		{
			name: "Direct typed error returning 400 should leave span Unset",
			handler: &spanStatusErrHandler{
				dataGetErr: &errapi.ErrorStatusCode{
					StatusCode: http.StatusBadRequest,
					Response:   errapi.Error{Code: 0, Message: "test error"},
				},
			},
			wantHTTPStatus: http.StatusBadRequest,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "NewError returning 400 should leave span Unset",
			handler: &spanStatusErrHandler{
				dataGetErr: errors.New("test handler error"),
				newErrorRes: &errapi.ErrorStatusCode{
					StatusCode: http.StatusBadRequest,
					Response:   errapi.Error{Code: 0, Message: "test error"},
				},
			},
			wantHTTPStatus: http.StatusBadRequest,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "NewError returning 500 should set span Error",
			handler: &spanStatusErrHandler{
				dataGetErr: errors.New("test handler error"),
				newErrorRes: &errapi.ErrorStatusCode{
					StatusCode: http.StatusInternalServerError,
					Response:   errapi.Error{Code: 0, Message: "test error"},
				},
			},
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tp := newSpanStatusTrace(t)
			srv := tt.handler.newServer(t, tp)

			req := httptest.NewRequest(http.MethodGet, "/data", nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			require.Equal(t, tt.wantHTTPStatus, rec.Code)

			requireServerSpanStatus(t, exp, tt.want)
		})
	}
}

func TestServerSpanStatusWebhooks(t *testing.T) {
	tests := []struct {
		name           string
		handler        *spanStatusWebhookHandler
		webhookName    string
		method         string
		wantHTTPStatus int
		want           wantSpanStatus
	}{
		{
			name:           "Webhook fixed 200 should leave span Unset",
			handler:        &spanStatusWebhookHandler{},
			webhookName:    "status",
			method:         http.MethodGet,
			wantHTTPStatus: http.StatusOK,
			want: wantSpanStatus{
				code: codes.Unset,
			},
		},
		{
			name: "Webhook dynamic 500 should set span Error",
			handler: &spanStatusWebhookHandler{
				statusCode: http.StatusInternalServerError,
			},
			webhookName:    "update",
			method:         http.MethodDelete,
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tp := newSpanStatusTrace(t)
			srv := tt.handler.newServer(t, tp)

			req := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()
			require.True(t, srv.Handle(tt.webhookName, rec, req))
			require.Equal(t, tt.wantHTTPStatus, rec.Code)

			requireServerSpanStatus(t, exp, tt.want)
		})
	}
}

func TestServerSpanStatusResponseEncodeFailures(t *testing.T) {
	tests := []struct {
		name           string
		handler        spanStatusServerHandler
		path           string
		wantHTTPStatus int
		want           wantSpanStatus
	}{
		{
			name: "nil typed response before WriteHeader should set span Error",
			handler: &spanStatusHandler{
				bodyRes: nil,
			},
			path:           "/span-status",
			wantHTTPStatus: http.StatusInternalServerError,
			want: wantSpanStatus{
				code: codes.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tp := newSpanStatusTrace(t)
			srv := tt.handler.newServer(t, tp)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			require.Equal(t, tt.wantHTTPStatus, rec.Code)

			requireServerSpanStatus(t, exp, tt.want)
		})
	}
}

func requireServerSpanStatus(t *testing.T, exp *tracetest.InMemoryExporter, want wantSpanStatus) {
	t.Helper()

	spans := exp.GetSpans()
	require.NotEmpty(t, spans, "no spans recorded")

	var serverSpan *tracetest.SpanStub
	for i := range spans {
		if spans[i].SpanKind == trace.SpanKindServer {
			serverSpan = &spans[i]
			break
		}
	}
	require.NotNil(t, serverSpan, "no server span found")

	assert.Equal(t, want.code, serverSpan.Status.Code,
		"span status code should match OTel HTTP Semantic Conventions")
}
