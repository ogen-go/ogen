package jsonpointer

import (
	"fmt"
	"testing"

	helperyaml "github.com/ghodss/yaml"
	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"
)

func getNode(t require.TestingT, input []byte) *yaml.Node {
	var n yaml.Node
	require.NoError(t, yaml.Unmarshal(input, &n))
	return &n
}

func getOutput(t require.TestingT, got *yaml.Node) string {
	data, err := yaml.Marshal(got)
	require.NoError(t, err)

	data, err = helperyaml.YAMLToJSON(data)
	require.NoError(t, err)

	return string(data)
}

func TestSpecification(t *testing.T) {
	specExample := []byte(`{
  "foo": ["bar", "baz"],
  "": 0,
  "a/b": 1,
  "c%d": 2,
  "e^f": 3,
  "g|h": 4,
  "i\\j": 5,
  "k\"l": 6,
  " ": 7,
  "m~n": 8
}`)

	tests := []struct {
		ptr     string
		want    string
		wantErr bool
	}{
		// Tests from https://datatracker.ietf.org/doc/html/rfc6901#section-5.
		{"", string(specExample), false},
		{"/foo", `["bar", "baz"]`, false},
		{"/foo/0", `"bar"`, false},
		{"/", "0", false},
		{"/a~1b", "1", false},
		{"/c%d", "2", false},
		{"/e^f", "3", false},
		{"/g|h", "4", false},
		{"/i\\j", "5", false},
		{"/k\"l", "6", false},
		{"/ ", "7", false},
		{"/m~0n", "8", false},

		// Tests from https://datatracker.ietf.org/doc/html/rfc6901#section-6.
		{"#", string(specExample), false},
		{"#/foo", `["bar", "baz"]`, false},
		{"#/foo/0", `"bar"`, false},
		{"#/", "0", false},
		{"#/a~1b", "1", false},
		{"#/c%25d", "2", false},
		{"#/e%5Ef", "3", false},
		{"#/g%7Ch", "4", false},
		{"#/i%5Cj", "5", false},
		{"#/k%22l", "6", false},
		{"#/%20", "7", false},
		{"#/m~0n", "8", false},

		// Test URL pointer.
		{"https://example.com#/m~0n", "8", false},

		// Invalid URL.
		{"\x00", "", true},

		// Invalid path.
		{"#foo/unknown", "", true},
		{"#%2", "", true},

		// Path does not exist.
		{"/foo/unknown", "", true},
		{"/foo/3", "", true},
		{"/foo/0/3", "", true},
		{"/foo/0/-3", "", true},
		{"/foo/-3", "", true},
		{"/bar/baz", "", true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, err := Resolve(tt.ptr, getNode(t, specExample))
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.JSONEq(tt.want, getOutput(t, got))
		})
	}
}

func TestResolveNilNode(t *testing.T) {
	a := require.New(t)
	var err error
	a.NotPanics(func() {
		_, err = Resolve("", nil)
	})
	a.EqualError(err, "root is nil")
}

func BenchmarkResolve(b *testing.B) {
	specExample := getNode(b, []byte(`{
  "openapi": "3.0.3",
  "components": {
    "schemas": {
      "Error": {
        "description": "Represents error object",
        "type": "object",
        "properties": {
          "code": {
            "type": "integer",
            "format": "int64"
          },
          "message": {
            "type": "string"
          }
        },
        "required": [
          "code",
          "message"
        ]
      }
    }
  }
}`))
	var (
		n   *yaml.Node
		err error
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n, err = Resolve("#/components/schemas/Error/properties/code/type", specExample)
	}

	if err != nil {
		b.Fatal(err)
	}
	if n.Value != `integer` {
		b.Fatal("unexpected result", n)
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		ptr     string
		input   string
		want    string
		wantErr bool
	}{
		{"/foo/0/0", `{"foo":[["foo"]]}`, `"foo"`, false},
		// Invalid path.
		{"/foo/0/-3/0", `{"foo":[["foo"]]}`, "", true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			got, err := Resolve(tt.ptr, getNode(t, []byte(tt.input)))
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.JSONEq(tt.want, getOutput(t, got))
		})
	}
}
