package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/middleware"
)

func Logging(logger *zap.Logger) middleware.Middleware {
	return func(
		req middleware.Request,
		next func(req middleware.Request) (middleware.Response, error),
	) (middleware.Response, error) {
		logger := logger.With(
			zap.String("operation", req.OperationName),
			zap.String("operationId", req.OperationID),
		)
		logger.Info("Handling request")
		resp, err := next(req)
		if err != nil {
			logger.Error("Fail", zap.Error(err))
		} else {
			logger.Info("Success")
		}
		return resp, err
	}
}

func ModifyRequest(logger *zap.Logger) middleware.Middleware {
	return func(
		req middleware.Request,
		next func(req middleware.Request) (middleware.Response, error),
	) (middleware.Response, error) {
		switch body := req.Body.(type) {
		case api.PetUploadAvatarByIDReq:
			if v, ok := req.Params["petID"].(int64); ok {
				logger.Info("Modifying request", zap.Int64("petID", v))
				req.Body = api.PetUploadAvatarByIDReq{
					Data: io.MultiReader(strings.NewReader("prefix"), body.Data),
				}
				req.Params["petID"] = v + 1
			}
		default:
			logger.Info("Skipping request modification")
		}
		return next(req)
	}
}

type testMiddleware struct {
	*sampleAPIServer
}

func (s *testMiddleware) PetUploadAvatarByID(
	ctx context.Context,
	req api.PetUploadAvatarByIDReq,
	params api.PetUploadAvatarByIDParams,
) (api.PetUploadAvatarByIDRes, error) {
	avatar, err := io.ReadAll(req)
	if err != nil {
		return &api.ErrorStatusCode{
			StatusCode: http.StatusInternalServerError,
			Response:   api.Error{Message: err.Error()},
		}, nil
	}

	if expected := petExistingID + 1; params.PetID != expected {
		return &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response:   api.Error{Message: fmt.Sprintf("expected %d, got %d", expected, params.PetID)},
		}, nil
	}

	// check that prefix was added.
	expected := append([]byte("prefix"), petAvatar...)
	if !bytes.Equal(avatar, expected) {
		return &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response:   api.Error{Message: fmt.Sprintf("expected %q, got %q", expected, avatar)},
		}, nil
	}

	return &api.PetUploadAvatarByIDOK{}, nil
}

func TestMiddleware(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	observerCore, logs := observer.New(zap.DebugLevel)
	logger := zaptest.NewLogger(t).WithOptions(
		zap.WrapCore(func(orig zapcore.Core) zapcore.Core {
			return zapcore.NewTee(orig, observerCore)
		}),
	)

	handler := &testMiddleware{}
	h, err := api.NewServer(handler, handler,
		api.WithMiddleware(
			Logging(logger.Named("logger")),
			ModifyRequest(logger.Named("modify")),
		),
	)
	a.NoError(err)

	s := httptest.NewServer(h)
	defer s.Close()

	client, err := api.NewClient(s.URL, handler, api.WithClient(s.Client()))
	a.NoError(err)

	stream := api.PetUploadAvatarByIDReq{
		Data: io.NopCloser(bytes.NewReader(petAvatar)),
	}
	got, err := client.PetUploadAvatarByID(ctx, stream, api.PetUploadAvatarByIDParams{
		PetID: petExistingID,
	})
	a.NoError(err)
	a.Equal(&api.PetUploadAvatarByIDOK{}, got)

	entries := logs.All()
	a.Len(entries, 3)
	a.Equal("Handling request", entries[0].Message)
}
