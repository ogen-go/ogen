package gen

import (
	"testing"

	"github.com/ogen-go/ogen/internal/ast"
	"github.com/stretchr/testify/require"
)

func TestPathParser(t *testing.T) {
	var (
		bar = &ast.Parameter{
			Name:       "Bar",
			SourceName: "bar",
			Schema:     ast.Primitive("int"),
			In:         ast.LocationPath,
		}
		baz = &ast.Parameter{
			Name:       "Baz",
			SourceName: "baz",
			Schema:     ast.Primitive("string"),
			In:         ast.LocationPath,
		}
	)

	tests := []struct {
		Name      string
		Path      string
		Params    []*ast.Parameter
		Expect    []ast.PathPart
		ExpectErr string
	}{
		{
			Name:   "test1",
			Path:   "/foo/{bar}",
			Params: []*ast.Parameter{bar},
			Expect: []ast.PathPart{
				{Raw: "/foo/"},
				{Param: bar},
			},
		},
		{
			Name:   "test2",
			Path:   "/foo.{bar}",
			Params: []*ast.Parameter{bar},
			Expect: []ast.PathPart{
				{Raw: "/foo."},
				{Param: bar},
			},
		},
		{
			Name:   "test3",
			Path:   "/foo.{bar}.{baz}abc/def",
			Params: []*ast.Parameter{bar, baz},
			Expect: []ast.PathPart{
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
			Params:    []*ast.Parameter{bar},
			ExpectErr: "path parameter 'baz' not found in parameters",
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
