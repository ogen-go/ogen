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

func Test_vetHeaderParameterName(t *testing.T) {
	observerCore, logs := observer.New(zap.WarnLevel)
	logger := zaptest.NewLogger(t).WithOptions(
		zap.WrapCore(func(orig zapcore.Core) zapcore.Core {
			return zapcore.NewTee(orig, observerCore)
		}),
	)

	a := require.New(t)
	{
		vetHeaderParameterName(logger, "content-type", &openapi.Header{}, "Content-Type")
		l := logs.TakeAll()
		a.Len(l, 2)
		a.Equal("Header name is not canonical, canonical name will be used", l[0].Message)
		a.Equal("Content-Type is described separately and will be ignored in this section", l[1].Message)
	}
	{
		vetHeaderParameterName(logger, "x-foo-bar", &openapi.Header{}, "Content-Type")
		l := logs.TakeAll()
		a.Len(l, 1)

		row := l[0]
		a.Equal("Header name is not canonical, canonical name will be used", row.Message)
		keys := row.ContextMap()
		a.Equal("x-foo-bar", keys["original_name"])
		a.Equal("X-Foo-Bar", keys["canonical_name"])
	}
}
