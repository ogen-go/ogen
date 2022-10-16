package ogen_test

import (
	"reflect"
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

func TestComponents_Init(t *testing.T) {
	a := require.New(t)

	c := ogen.Components{}
	c.Init()

	val := reflect.ValueOf(c)
	a.Equalf(10+1, val.NumField(), "update this test if you add new fields to Components")
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		// Init doesn't set Common and it's ok.
		if val.Type().Field(i).Name == "Common" {
			continue
		}
		a.NotNil(f.Interface())
	}
}
