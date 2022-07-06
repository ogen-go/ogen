package ogen

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	ogenjson "github.com/ogen-go/ogen/json"
)

func Test_unmarshalJSON(t *testing.T) {
	const testdata = `{
  "openapi": "3.1.0",
  "info": {
    "title": "API",
    "version": "0.1.0"
  },
  "components": {
    "schemas": {
      "User": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "example": "Thomas A. Anderson",
            "required": true
          }
        }
      }
    }
  }
}`
	ubool := func(input []byte) error {
		var target bool
		return unmarshalJSON(input, &target)
	}
	tests := []struct {
		input        string
		target       func([]byte) error
		line, column int
	}{
		{"{}", ubool, 1, 1},
		{"{}\n", ubool, 1, 1},
		{"\x20{}", ubool, 1, 2},
		{"\x20{}\n", ubool, 1, 2},
		{"\n{}", ubool, 2, 1},
		{"\n{}\n", ubool, 2, 1},
		{"\n\n{}", ubool, 3, 1},
		{"\n\n{}\n", ubool, 3, 1},
		{"\x20\n{}", ubool, 2, 1},
		{"\x20\n{}\n", ubool, 2, 1},

		{"{\n\t\"a\":1,\n\t\"b\":2\n}", func(input []byte) error {
			var target struct {
				A int  `json:"a"`
				B bool `json:"b"`
			}
			return unmarshalJSON(input, &target)
		}, 3, 6},

		{"[\n0,\ntrue\n]", func(input []byte) error {
			var target []int
			return unmarshalJSON(input, &target)
		}, 3, 1},

		{testdata, func(input []byte) error {
			var target *Spec
			return unmarshalJSON(input, &target)
		}, 15, 25},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			input := []byte(tt.input)

			a := require.New(t)
			err := tt.target(input)
			a.Error(err)

			msg := err.Error()
			prefix := fmt.Sprintf("line %d:%d", tt.line, tt.column)
			a.Truef(strings.HasPrefix(msg, prefix), "input: %q,\nprefix: %q,\nmsg: %q", tt.input, prefix, msg)
		})
	}
}
