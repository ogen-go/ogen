package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
	"go.opentelemetry.io/otel/sdk/resource"

	api "github.com/ogen-go/ogen/internal/integration/sample_api"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/otelogen"
)

type metricsTestHandler struct {
	api.UnimplementedHandler
}

func (h metricsTestHandler) PetNameByID(ctx context.Context, params api.PetNameByIDParams) (string, error) {
	return "Fluffy", nil
}

func (h metricsTestHandler) PetGetByName(ctx context.Context, params api.PetGetByNameParams) (*api.Pet, error) {
	return nil, context.DeadlineExceeded
}

func (h metricsTestHandler) HandleAPIKey(ctx context.Context, operationName string, t api.APIKey) (context.Context, error) {
	return ctx, nil
}

func (h metricsTestHandler) APIKey(ctx context.Context, operationName string) (api.APIKey, error) {
	return api.APIKey{
		APIKey: "blah",
	}, nil
}

var _ api.Handler = metricsTestHandler{}
var _ api.SecurityHandler = metricsTestHandler{}
var _ api.SecuritySource = metricsTestHandler{}

func labelerMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	labeler, _ := api.LabelerFromContext(req.Context)

	if id, ok := req.Params.Path("id"); ok {
		labeler.Add(
			attribute.Int("pet_id", id.(int)),
		)
	}

	if name, ok := req.Params.Path("name"); ok {
		labeler.Add(
			attribute.String("pet_name", name.(string)),
		)
	}

	return next(req)
}

func labelerErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	var errType string
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		errType = "timeout"
	case errors.Is(err, context.Canceled):
		errType = "canceled"
	default:
		errType = "other"
	}

	labeler, _ := api.LabelerFromContext(ctx)
	labeler.Add(
		attribute.String("error.type", errType),
	)

	ogenerrors.DefaultErrorHandler(ctx, w, r, err)
}

func TestServerMetrics(t *testing.T) {
	tests := []struct {
		name      string
		mw        middleware.Middleware
		eh        ogenerrors.ErrorHandler
		operation func(*testing.T, *api.Client)
		want      []metricdata.Metrics
	}{
		{
			name: "Success",
			operation: func(t *testing.T, c *api.Client) {
				name, err := c.PetNameByID(context.Background(), api.PetNameByIDParams{
					ID: 1,
				})
				assert.Equal(t, "Fluffy", name)
				assert.NoError(t, err)
			},
			want: []metricdata.Metrics{
				{
					Name:        "ogen.server.request_count",
					Description: "Incoming request count total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petNameByID"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/name/{id}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name:        "ogen.server.duration",
					Description: "Incoming end to end duration",
					Unit:        "ms",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petNameByID"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/name/{id}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
			},
		},
		{
			name: "Error",
			operation: func(t *testing.T, c *api.Client) {
				pet, err := c.PetGetByName(context.Background(), api.PetGetByNameParams{
					Name: "Fluffy",
				})
				assert.Nil(t, pet)
				assert.EqualError(t, err, "decode response: unexpected status code: 500")
			},
			want: []metricdata.Metrics{
				{
					Name:        "ogen.server.request_count",
					Description: "Incoming request count total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name:        "ogen.server.duration",
					Description: "Incoming end to end duration",
					Unit:        "ms",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
				{
					Name:        "ogen.server.errors_count",
					Description: "Incoming errors total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
			},
		},
		{
			name: "Success with Labeler",
			mw:   labelerMiddleware,
			operation: func(t *testing.T, c *api.Client) {
				name, err := c.PetNameByID(context.Background(), api.PetNameByIDParams{
					ID: 1,
				})
				assert.Equal(t, "Fluffy", name)
				assert.NoError(t, err)
			},
			want: []metricdata.Metrics{
				{
					Name:        "ogen.server.request_count",
					Description: "Incoming request count total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petNameByID"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/name/{id}"),
									attribute.Int("pet_id", 1),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name:        "ogen.server.duration",
					Description: "Incoming end to end duration",
					Unit:        "ms",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petNameByID"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/name/{id}"),
									attribute.Int("pet_id", 1),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
			},
		},
		{
			name: "Error with Labeler",
			mw:   labelerMiddleware,
			eh:   labelerErrorHandler,
			operation: func(t *testing.T, c *api.Client) {
				pet, err := c.PetGetByName(context.Background(), api.PetGetByNameParams{
					Name: "Fluffy",
				})
				assert.Nil(t, pet)
				assert.EqualError(t, err, "decode response: unexpected status code: 500")
			},
			want: []metricdata.Metrics{
				{
					Name:        "ogen.server.request_count",
					Description: "Incoming request count total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
									attribute.String("pet_name", "Fluffy"),
									attribute.String("error.type", "timeout"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name:        "ogen.server.duration",
					Description: "Incoming end to end duration",
					Unit:        "ms",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
									attribute.String("pet_name", "Fluffy"),
									attribute.String("error.type", "timeout"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
				{
					Name:        "ogen.server.errors_count",
					Description: "Incoming errors total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
									attribute.String("pet_name", "Fluffy"),
									attribute.String("error.type", "timeout"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			reader := sdkmetric.NewManualReader()
			mp := sdkmetric.NewMeterProvider(
				sdkmetric.WithReader(reader),
			)
			defer func() {
				err := mp.Shutdown(ctx)
				assert.NoError(t, err)
			}()

			handler := metricsTestHandler{}
			h, err := api.NewServer(handler, handler,
				api.WithMeterProvider(mp),
				api.WithMiddleware(tt.mw),
				api.WithErrorHandler(tt.eh),
			)
			require.NoError(t, err)

			s := httptest.NewServer(h)
			defer s.Close()

			client, err := api.NewClient(s.URL, handler,
				api.WithClient(s.Client()),
				api.WithMeterProvider(metricnoop.NewMeterProvider()),
			)
			require.NoError(t, err)

			// Perform the operation
			tt.operation(t, client)

			// Check server metrics
			var rm metricdata.ResourceMetrics
			err = reader.Collect(ctx, &rm)
			require.NoError(t, err)

			expected := metricdata.ResourceMetrics{
				Resource: resource.Default(),
				ScopeMetrics: []metricdata.ScopeMetrics{
					{
						Scope: instrumentation.Scope{
							Name:    otelogen.Name,
							Version: otelogen.SemVersion(),
						},
						Metrics: tt.want,
					},
				},
			}

			metricdatatest.AssertEqual(
				t, expected, rm,
				metricdatatest.IgnoreTimestamp(),
				metricdatatest.IgnoreValue(),
				metricdatatest.IgnoreExemplars(),
			)
		})
	}
}

func TestClientMetrics(t *testing.T) {
	tests := []struct {
		name      string
		operation func(*testing.T, *api.Client)
		want      []metricdata.Metrics
	}{
		{
			name: "Success",
			operation: func(t *testing.T, c *api.Client) {
				name, err := c.PetNameByID(context.Background(), api.PetNameByIDParams{
					ID: 1,
				})
				assert.Equal(t, "Fluffy", name)
				assert.NoError(t, err)
			},
			want: []metricdata.Metrics{
				{
					Name:        "ogen.client.request_count",
					Description: "Outgoing request count total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petNameByID"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/name/{id}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name:        "ogen.client.duration",
					Description: "Outgoing end to end duration",
					Unit:        "ms",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petNameByID"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/name/{id}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
			},
		},
		{
			name: "Error",
			operation: func(t *testing.T, c *api.Client) {
				pet, err := c.PetGetByName(context.Background(), api.PetGetByNameParams{
					Name: "Fluffy",
				})
				assert.Nil(t, pet)
				assert.EqualError(t, err, "decode response: unexpected status code: 500")
			},
			want: []metricdata.Metrics{
				{
					Name:        "ogen.client.request_count",
					Description: "Outgoing request count total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name:        "ogen.client.duration",
					Description: "Outgoing end to end duration",
					Unit:        "ms",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
				{
					Name:        "ogen.client.errors_count",
					Description: "Outgoing errors total",
					Unit:        "{count}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("oas.operation", "petGetByName"),
									attribute.String("http.request.method", "GET"),
									attribute.String("http.route", "/pet/{name}"),
								),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			reader := sdkmetric.NewManualReader()
			mp := sdkmetric.NewMeterProvider(
				sdkmetric.WithReader(reader),
			)
			defer func() {
				err := mp.Shutdown(ctx)
				assert.NoError(t, err)
			}()

			handler := metricsTestHandler{}
			h, err := api.NewServer(handler, handler,
				api.WithMeterProvider(metricnoop.NewMeterProvider()),
			)
			require.NoError(t, err)

			s := httptest.NewServer(h)
			defer s.Close()

			client, err := api.NewClient(s.URL, handler,
				api.WithClient(s.Client()),
				api.WithMeterProvider(mp),
			)
			require.NoError(t, err)

			// Perform the operation
			tt.operation(t, client)

			// Check client metrics
			var rm metricdata.ResourceMetrics
			err = reader.Collect(ctx, &rm)
			require.NoError(t, err)

			expected := metricdata.ResourceMetrics{
				Resource: resource.Default(),
				ScopeMetrics: []metricdata.ScopeMetrics{
					{
						Scope: instrumentation.Scope{
							Name:    otelogen.Name,
							Version: otelogen.SemVersion(),
						},
						Metrics: tt.want,
					},
				},
			}

			metricdatatest.AssertEqual(
				t, expected, rm,
				metricdatatest.IgnoreTimestamp(),
				metricdatatest.IgnoreValue(),
				metricdatatest.IgnoreExemplars(),
			)
		})
	}
}
