// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/go-faster/errors"
	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/otelogen"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func trimTrailingSlashes(u *url.URL) {
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = strings.TrimRight(u.RawPath, "/")
}

// Invoker invokes operations described by OpenAPI v3 specification.
type Invoker interface {
	// NullableStrings invokes nullableStrings operation.
	//
	// Nullable strings.
	//
	// POST /nullableStrings
	NullableStrings(ctx context.Context, request NilString) error
	// ObjectsWithConflictingArrayProperty invokes objectsWithConflictingArrayProperty operation.
	//
	// Objects with conflicting array property.
	//
	// POST /objectsWithConflictingArrayProperty
	ObjectsWithConflictingArrayProperty(ctx context.Context, request *ObjectsWithConflictingArrayPropertyReq) error
	// ObjectsWithConflictingProperties invokes objectsWithConflictingProperties operation.
	//
	// Objects with conflicting properties.
	//
	// POST /objectsWithConflictingProperties
	ObjectsWithConflictingProperties(ctx context.Context, request *ObjectsWithConflictingPropertiesReq) error
	// ReferencedAllOfNullable invokes referencedAllOfNullable operation.
	//
	// Referenced allOf, but requestBody contains nullable refs.
	//
	// POST /referencedAllOfNullable
	ReferencedAllOfNullable(ctx context.Context, request ReferencedAllOfNullableReq) error
	// ReferencedAllof invokes referencedAllof operation.
	//
	// Referenced allOf.
	//
	// POST /referencedAllof
	ReferencedAllof(ctx context.Context, request ReferencedAllofReq) error
	// ReferencedAllofOptional invokes referencedAllofOptional operation.
	//
	// Referenced allOf, but requestBody is not required.
	//
	// POST /referencedAllofOptional
	ReferencedAllofOptional(ctx context.Context, request ReferencedAllofOptionalReq) error
	// SimpleInteger invokes simpleInteger operation.
	//
	// Simple integers with validation.
	//
	// POST /simpleInteger
	SimpleInteger(ctx context.Context, request int) error
	// SimpleObjects invokes simpleObjects operation.
	//
	// Simple objects.
	//
	// POST /simpleObjects
	SimpleObjects(ctx context.Context, request *SimpleObjectsReq) error
	// StringsNotype invokes stringsNotype operation.
	//
	// POST /stringsNotype
	StringsNotype(ctx context.Context, request NilString) error
}

// Client implements OAS client.
type Client struct {
	serverURL *url.URL
	baseClient
}

var _ Handler = struct {
	*Client
}{}

// NewClient initializes new Client defined by OAS.
func NewClient(serverURL string, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	trimTrailingSlashes(u)

	c, err := newClientConfig(opts...).baseClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		serverURL:  u,
		baseClient: c,
	}, nil
}

type serverURLKey struct{}

// WithServerURL sets context key to override server URL.
func WithServerURL(ctx context.Context, u *url.URL) context.Context {
	return context.WithValue(ctx, serverURLKey{}, u)
}

func (c *Client) requestURL(ctx context.Context) *url.URL {
	u, ok := ctx.Value(serverURLKey{}).(*url.URL)
	if !ok {
		return c.serverURL
	}
	return u
}

// NullableStrings invokes nullableStrings operation.
//
// Nullable strings.
//
// POST /nullableStrings
func (c *Client) NullableStrings(ctx context.Context, request NilString) error {
	_, err := c.sendNullableStrings(ctx, request)
	return err
}

func (c *Client) sendNullableStrings(ctx context.Context, request NilString) (res *NullableStringsOK, err error) {
	// Validate request before sending.
	if err := func() error {
		if value, ok := request.Get(); ok {
			if err := func() error {
				if err := (validate.String{
					MinLength:    0,
					MinLengthSet: false,
					MaxLength:    0,
					MaxLengthSet: false,
					Email:        false,
					Hostname:     false,
					Regex:        regexMap["^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$"],
				}).Validate(string(value)); err != nil {
					return errors.Wrap(err, "string")
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return res, errors.Wrap(err, "validate")
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("nullableStrings"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/nullableStrings"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, NullableStringsOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/nullableStrings"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeNullableStringsRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeNullableStringsResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// ObjectsWithConflictingArrayProperty invokes objectsWithConflictingArrayProperty operation.
//
// Objects with conflicting array property.
//
// POST /objectsWithConflictingArrayProperty
func (c *Client) ObjectsWithConflictingArrayProperty(ctx context.Context, request *ObjectsWithConflictingArrayPropertyReq) error {
	_, err := c.sendObjectsWithConflictingArrayProperty(ctx, request)
	return err
}

func (c *Client) sendObjectsWithConflictingArrayProperty(ctx context.Context, request *ObjectsWithConflictingArrayPropertyReq) (res *ObjectsWithConflictingArrayPropertyOK, err error) {
	// Validate request before sending.
	if err := func() error {
		if err := request.Validate(); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return res, errors.Wrap(err, "validate")
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("objectsWithConflictingArrayProperty"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/objectsWithConflictingArrayProperty"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, ObjectsWithConflictingArrayPropertyOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/objectsWithConflictingArrayProperty"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeObjectsWithConflictingArrayPropertyRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeObjectsWithConflictingArrayPropertyResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// ObjectsWithConflictingProperties invokes objectsWithConflictingProperties operation.
//
// Objects with conflicting properties.
//
// POST /objectsWithConflictingProperties
func (c *Client) ObjectsWithConflictingProperties(ctx context.Context, request *ObjectsWithConflictingPropertiesReq) error {
	_, err := c.sendObjectsWithConflictingProperties(ctx, request)
	return err
}

func (c *Client) sendObjectsWithConflictingProperties(ctx context.Context, request *ObjectsWithConflictingPropertiesReq) (res *ObjectsWithConflictingPropertiesOK, err error) {
	// Validate request before sending.
	if err := func() error {
		if err := request.Validate(); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return res, errors.Wrap(err, "validate")
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("objectsWithConflictingProperties"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/objectsWithConflictingProperties"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, ObjectsWithConflictingPropertiesOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/objectsWithConflictingProperties"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeObjectsWithConflictingPropertiesRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeObjectsWithConflictingPropertiesResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// ReferencedAllOfNullable invokes referencedAllOfNullable operation.
//
// Referenced allOf, but requestBody contains nullable refs.
//
// POST /referencedAllOfNullable
func (c *Client) ReferencedAllOfNullable(ctx context.Context, request ReferencedAllOfNullableReq) error {
	_, err := c.sendReferencedAllOfNullable(ctx, request)
	return err
}

func (c *Client) sendReferencedAllOfNullable(ctx context.Context, request ReferencedAllOfNullableReq) (res *ReferencedAllOfNullableOK, err error) {
	// Validate request before sending.
	switch request := request.(type) {
	case *ReferencedAllOfNullableReqEmptyBody:
		// Validation is not needed for the empty body type.
	case *ReferencedAllOfNullable:
		if err := func() error {
			if err := request.Validate(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, errors.Wrap(err, "validate")
		}
	case *ReferencedAllOfNullableMultipart:
		if err := func() error {
			if err := request.Validate(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, errors.Wrap(err, "validate")
		}
	default:
		return res, errors.Errorf("unexpected request type: %T", request)
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("referencedAllOfNullable"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/referencedAllOfNullable"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, ReferencedAllOfNullableOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/referencedAllOfNullable"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeReferencedAllOfNullableRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeReferencedAllOfNullableResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// ReferencedAllof invokes referencedAllof operation.
//
// Referenced allOf.
//
// POST /referencedAllof
func (c *Client) ReferencedAllof(ctx context.Context, request ReferencedAllofReq) error {
	_, err := c.sendReferencedAllof(ctx, request)
	return err
}

func (c *Client) sendReferencedAllof(ctx context.Context, request ReferencedAllofReq) (res *ReferencedAllofOK, err error) {
	// Validate request before sending.
	switch request := request.(type) {
	case *Robot:
		if err := func() error {
			if err := request.Validate(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, errors.Wrap(err, "validate")
		}
	case *RobotMultipart:
		if err := func() error {
			if err := request.Validate(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, errors.Wrap(err, "validate")
		}
	default:
		return res, errors.Errorf("unexpected request type: %T", request)
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("referencedAllof"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/referencedAllof"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, ReferencedAllofOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/referencedAllof"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeReferencedAllofRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeReferencedAllofResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// ReferencedAllofOptional invokes referencedAllofOptional operation.
//
// Referenced allOf, but requestBody is not required.
//
// POST /referencedAllofOptional
func (c *Client) ReferencedAllofOptional(ctx context.Context, request ReferencedAllofOptionalReq) error {
	_, err := c.sendReferencedAllofOptional(ctx, request)
	return err
}

func (c *Client) sendReferencedAllofOptional(ctx context.Context, request ReferencedAllofOptionalReq) (res *ReferencedAllofOptionalOK, err error) {
	// Validate request before sending.
	switch request := request.(type) {
	case *ReferencedAllofOptionalReqEmptyBody:
		// Validation is not needed for the empty body type.
	case *Robot:
		if err := func() error {
			if err := request.Validate(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, errors.Wrap(err, "validate")
		}
	case *RobotMultipart:
		if err := func() error {
			if err := request.Validate(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, errors.Wrap(err, "validate")
		}
	default:
		return res, errors.Errorf("unexpected request type: %T", request)
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("referencedAllofOptional"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/referencedAllofOptional"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, ReferencedAllofOptionalOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/referencedAllofOptional"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeReferencedAllofOptionalRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeReferencedAllofOptionalResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// SimpleInteger invokes simpleInteger operation.
//
// Simple integers with validation.
//
// POST /simpleInteger
func (c *Client) SimpleInteger(ctx context.Context, request int) error {
	_, err := c.sendSimpleInteger(ctx, request)
	return err
}

func (c *Client) sendSimpleInteger(ctx context.Context, request int) (res *SimpleIntegerOK, err error) {
	// Validate request before sending.
	if err := func() error {
		if err := (validate.Int{
			MinSet:        true,
			Min:           -5,
			MaxSet:        true,
			Max:           5,
			MinExclusive:  false,
			MaxExclusive:  false,
			MultipleOfSet: false,
			MultipleOf:    0,
		}).Validate(int64(request)); err != nil {
			return errors.Wrap(err, "int")
		}
		return nil
	}(); err != nil {
		return res, errors.Wrap(err, "validate")
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("simpleInteger"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/simpleInteger"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, SimpleIntegerOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/simpleInteger"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeSimpleIntegerRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeSimpleIntegerResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// SimpleObjects invokes simpleObjects operation.
//
// Simple objects.
//
// POST /simpleObjects
func (c *Client) SimpleObjects(ctx context.Context, request *SimpleObjectsReq) error {
	_, err := c.sendSimpleObjects(ctx, request)
	return err
}

func (c *Client) sendSimpleObjects(ctx context.Context, request *SimpleObjectsReq) (res *SimpleObjectsOK, err error) {
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("simpleObjects"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/simpleObjects"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, SimpleObjectsOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/simpleObjects"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeSimpleObjectsRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeSimpleObjectsResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// StringsNotype invokes stringsNotype operation.
//
// POST /stringsNotype
func (c *Client) StringsNotype(ctx context.Context, request NilString) error {
	_, err := c.sendStringsNotype(ctx, request)
	return err
}

func (c *Client) sendStringsNotype(ctx context.Context, request NilString) (res *StringsNotypeOK, err error) {
	// Validate request before sending.
	if err := func() error {
		if value, ok := request.Get(); ok {
			if err := func() error {
				if err := (validate.String{
					MinLength:    0,
					MinLengthSet: false,
					MaxLength:    15,
					MaxLengthSet: true,
					Email:        false,
					Hostname:     false,
					Regex:        nil,
				}).Validate(string(value)); err != nil {
					return errors.Wrap(err, "string")
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return res, errors.Wrap(err, "validate")
	}
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("stringsNotype"),
		semconv.HTTPRequestMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/stringsNotype"),
	}
	otelAttrs = append(otelAttrs, c.cfg.Attributes...)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(elapsedDuration)/float64(time.Millisecond), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, StringsNotypeOperation,
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/stringsNotype"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodeStringsNotypeRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodeStringsNotypeResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}
