package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func TestPathParser(t *testing.T) {
	var (
		bar = &openapi.Parameter{
			Name:   "bar",
			Schema: &jsonschema.Schema{Type: jsonschema.Integer},
			In:     openapi.LocationPath,
		}
		baz = &openapi.Parameter{
			Name:   "baz",
			Schema: &jsonschema.Schema{Type: jsonschema.String},
			In:     openapi.LocationPath,
		}
	)

	tests := []struct {
		Name      string
		Path      string
		Params    []*openapi.Parameter
		Expect    []openapi.PathPart
		ExpectErr string
	}{
		{
			Name:   "test1",
			Path:   "/foo/{bar}",
			Params: []*openapi.Parameter{bar},
			Expect: []openapi.PathPart{
				{Raw: "/foo/"},
				{Param: bar},
			},
		},
		{
			Name:   "test2",
			Path:   "/foo.{bar}",
			Params: []*openapi.Parameter{bar},
			Expect: []openapi.PathPart{
				{Raw: "/foo."},
				{Param: bar},
			},
		},
		{
			Name:   "test3",
			Path:   "/foo.{bar}.{baz}abc/def",
			Params: []*openapi.Parameter{bar, baz},
			Expect: []openapi.PathPart{
				{Raw: "/foo."},
				{Param: bar},
				{Raw: "."},
				{Param: baz},
				{Raw: "abc/def"},
			},
		},
		{
			Name:      "test4",
			Path:      "/foo/{bar}/{baz}",
			Params:    []*openapi.Parameter{bar},
			ExpectErr: `path parameter not specified: "baz"`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, err := parsePath(test.Path, test.Params)
			if test.ExpectErr != "" {
				require.EqualError(t, err, test.ExpectErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.Expect, result)
		})
	}
}
