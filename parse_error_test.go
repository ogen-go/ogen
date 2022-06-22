package ogen

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_wrapLineOffset(t *testing.T) {
	tests := []struct {
		input        string
		target       func([]byte) error
		line, offset int
	}{
		{"{\n\t\"a\":1,\n\t\"b\":2\n}", func(input []byte) error {
			var target struct {
				A int  `json:"a"`
				B bool `json:"b"`
			}
			return json.Unmarshal(input, &target)
		}, 3, 7},

		{"[\n0,\ntrue\n]", func(input []byte) error {
			var target []int
			return json.Unmarshal(input, &target)
		}, 3, 5},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			input := []byte(tt.input)

			a := require.New(t)
			err := tt.target(input)
			a.Error(err)

			err = wrapLineOffset(input, err)
			a.Error(err)

			msg := err.Error()
			prefix := fmt.Sprintf("line %d:%d", tt.line, tt.offset)
			a.Truef(strings.HasPrefix(msg, prefix), "prefix: %s, msg: %s", prefix, msg)
		})
	}
}
