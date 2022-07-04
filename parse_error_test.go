package ogen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_wrapLineOffset(t *testing.T) {
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
	tests := []struct {
		input        string
		target       func([]byte) error
		line, column int
	}{
		{"{\n\t\"a\":1,\n\t\"b\":2\n}", func(input []byte) error {
			var target struct {
				A int  `json:"a"`
				B bool `json:"b"`
			}
			return unmarshal(input, &target)
		}, 3, 6},

		{"[\n0,\ntrue\n]", func(input []byte) error {
			var target []int
			return unmarshal(input, &target)
		}, 3, 1},
		{testdata, func(input []byte) error {
			var target *Spec
			return unmarshal(input, &target)
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
			a.Truef(strings.HasPrefix(msg, prefix), "prefix: %q, msg: %q", prefix, msg)
		})
	}
}
