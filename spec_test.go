package ogen_test

import (
	"testing"

	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func TestExtensionParsing(t *testing.T) {
	a := require.New(t)

	{
		var (
			input = `{"url": "/api/v1", "x-ogen-name": "foo"}`
			s     ogen.Server
		)
		a.NoError(yaml.Unmarshal([]byte(input), &s))
		a.Equal("foo", s.Common.Extensions["x-ogen-name"].Value)
		// FIXME(tdakkota): encodeDecode doesn't work for this type
	}

	{
		var (
			input = `{"description": "foo", "x-ogen-extension": "bar"}`
			s     ogen.Response
		)
		a.NoError(yaml.Unmarshal([]byte(input), &s))
		a.Equal("bar", s.Common.Extensions["x-ogen-extension"].Value)
		// FIXME(tdakkota): encodeDecode doesn't work for this type
	}
}
