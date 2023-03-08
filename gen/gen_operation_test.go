package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ogen-go/ogen/openapi"
)

func Test_vetPathParametersUsed(t *testing.T) {
	observerCore, logs := observer.New(zap.WarnLevel)
	logger := zaptest.NewLogger(t).WithOptions(
		zap.WrapCore(func(orig zapcore.Core) zapcore.Core {
			return zapcore.NewTee(orig, observerCore)
		}),
	)

	a := require.New(t)
	{
		id := &openapi.Parameter{
			Name: "id",
			In:   openapi.LocationPath,
		}
		queryParam := &openapi.Parameter{
			Name: "search",
			In:   openapi.LocationQuery,
		}
		notUsed := &openapi.Parameter{
			Name: "not_used",
			In:   openapi.LocationPath,
		}
		parts := openapi.Path{
			{Raw: "/users/"},
			{Param: id},
		}
		params := []*openapi.Parameter{
			id,
			queryParam,
			notUsed,
		}
		vetPathParametersUsed(logger, parts, params)
		l := logs.TakeAll()
		a.Len(l, 1)
		a.Equal("Path parameter is not used", l[0].Message)
		a.Equal(notUsed.Name, l[0].ContextMap()["name"])
	}
}
