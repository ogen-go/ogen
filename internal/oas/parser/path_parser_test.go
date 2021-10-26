package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/oas"
)

func TestPathParser(t *testing.T) {
	var (
		bar = &oas.Parameter{
			Name:   "bar",
			Schema: &oas.Schema{Type: oas.Integer},
			In:     oas.LocationPath,
		}
		baz = &oas.Parameter{
			Name:   "baz",
			Schema: &oas.Schema{Type: oas.String},
			In:     oas.LocationPath,
		}
	)

	tests := []struct {
		Name      string
		Path      string
		Params    []*oas.Parameter
		Expect    []oas.PathPart
		ExpectErr string
	}{
		{
			Name:   "test1",
			Path:   "/foo/{bar}",
			Params: []*oas.Parameter{bar},
			Expect: []oas.PathPart{
				{Raw: "/foo/"},
				{Param: bar},
			},
		},
		{
			Name:   "test2",
			Path:   "/foo.{bar}",
			Params: []*oas.Parameter{bar},
			Expect: []oas.PathPart{
				{Raw: "/foo."},
				{Param: bar},
			},
		},
		{
			Name:   "test3",
			Path:   "/foo.{bar}.{baz}abc/def",
			Params: []*oas.Parameter{bar, baz},
			Expect: []oas.PathPart{
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
			Params:    []*oas.Parameter{bar},
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
