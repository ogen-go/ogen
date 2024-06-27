package ogen_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-faster/yaml"
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

func TestSwaggerExtraction(t *testing.T) {
	a := require.New(t)

	{
		var (
			input = `{"swagger": "2.0.0"}`
			s     ogen.Spec
		)
		a.NoError(yaml.Unmarshal([]byte(input), &s))
		a.Equal("2.0.0", s.Swagger)
	}
}

func TestExtensionsMarshal(t *testing.T) {
	a := require.New(t)

	extensionValueFunc := func(key, value string) ogen.OpenAPICommon {
		return ogen.OpenAPICommon{
			Extensions: ogen.Extensions{
				key: yaml.Node{Kind: yaml.ScalarNode, Value: value},
			},
		}
	}

	extentionKey := "x-ogen-extension"
	extenstionValue := "handler"

	{
		pathItem := ogen.NewPathItem()
		pathItem.Common = extensionValueFunc(extentionKey, extenstionValue)

		pathItemJSON, err := json.Marshal(pathItem)
		a.NoError(err)

		var output map[string]interface{}
		err = json.Unmarshal(pathItemJSON, &output)
		a.NoError(err)

		v, ok := output[extentionKey]
		a.True(ok)

		a.Equal(extenstionValue, v)
	}

	{
		op := ogen.NewOperation()
		op.Common = extensionValueFunc(extentionKey, extenstionValue)

		pathItemJSON, err := json.Marshal(op)
		a.NoError(err)

		var output map[string]interface{}
		err = json.Unmarshal(pathItemJSON, &output)
		a.NoError(err)

		v, ok := output[extentionKey]
		a.True(ok)

		a.Equal(extenstionValue, v)
	}
}
