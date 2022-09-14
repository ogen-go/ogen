package ogen

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/location"
)

func TestLocator(t *testing.T) {
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
		return yaml.Unmarshal(input, &target)
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
			return yaml.Unmarshal(input, &target)
		}, 3, 6},

		{"[\n0,\ntrue\n]", func(input []byte) error {
			var target []int
			return yaml.Unmarshal(input, &target)
		}, 3, 1},

		{testdata, func(input []byte) error {
			var target *Spec
			return yaml.Unmarshal(input, &target)
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
			prefix := fmt.Sprintf("line %d:", tt.line)
			a.Truef(strings.Contains(msg, prefix), "input: %q,\ncontains: %q,\nmsg: %q", tt.input, prefix, msg)
		})
	}
}

var (
	//go:embed _testdata/location/location_spec.json
	locationSpecJSON string
	//go:embed _testdata/location/location_spec.yml
	locationSpecYAML string
)

func TestLocation(t *testing.T) {
	createEqualLoc := func(a *assert.Assertions, data []byte) func(l location.Locatable, line, column int) {
		var lines location.Lines
		lines.Collect(data)
		return func(l location.Locatable, line, column int) {
			t.Helper()

			loc, ok := l.Location()
			a.True(ok)
			getLine := func(n int) string {
				start, end := lines.Line(n)
				// Offset points exactly to the newline, trim it.
				return strings.Trim(string(data[start:end]), "\n\r")
			}

			type compareLoc struct {
				Line, Column int
				Data         string
			}
			a.Equal(
				compareLoc{line, column, getLine(line)},
				compareLoc{loc.Line, loc.Column, getLine(loc.Line)},
			)
		}
	}

	t.Run("JSON", func(t *testing.T) {
		a := assert.New(t)
		equalLoc := createEqualLoc(a, []byte(locationSpecJSON))

		locationSpec, err := Parse([]byte(locationSpecJSON))
		require.NoError(t, err)

		var (
			foo    = locationSpec.Paths["/foo"]
			post   = foo.Post
			get    = foo.Get
			body   = post.RequestBody
			media  = body.Content["application/json"]
			schema = media.Schema
		)
		// Compare PathItem.
		equalLoc(&foo.Common.Locator, 8, 13)

		// Compare post
		equalLoc(&post.Common.Locator, 9, 15)

		// Compare Parameters.
		equalLoc(&post.Parameters[0].Common.Locator, 11, 11)
		equalLoc(&post.Parameters[1].Common.Locator, 18, 11)

		// Compare RequestBody.
		equalLoc(&body.Common.Locator, 26, 24)
		equalLoc(&media.Common.Locator, 28, 33)
		equalLoc(&schema.Common.Locator, 29, 25)

		// Compare get.
		equalLoc(&get.Common.Locator, 48, 14)

		var user = locationSpec.Components.Schemas["User"]
		equalLoc(&user.Common.Locator, 59, 15)
		equalLoc(&user.Properties[0].Schema.Common.Locator, 62, 19)
	})
	t.Run("YAML", func(t *testing.T) {
		a := assert.New(t)
		equalLoc := createEqualLoc(a, []byte(locationSpecYAML))

		locationSpec, err := Parse([]byte(locationSpecYAML))
		require.NoError(t, err)

		var (
			foo           = locationSpec.Paths["/foo"]
			post          = foo.Post
			body          = post.RequestBody
			requestMedia  = body.Content["application/json"]
			requestSchema = requestMedia.Schema
		)
		// FIXME(tdakkota): parser sets map/seq location to the first element.
		// Compare PathItem and Operation.
		equalLoc(&foo.Common.Locator, 7, 5)
		equalLoc(&post.Common.Locator, 8, 7)

		// Compare Parameters.
		equalLoc(&post.Parameters[0].Common.Locator, 9, 11)
		equalLoc(&post.Parameters[1].Common.Locator, 13, 11)

		// Compare RequestBody.
		equalLoc(&body.Common.Locator, 18, 9)
		equalLoc(&requestMedia.Common.Locator, 20, 13)
		equalLoc(&requestSchema.Common.Locator, 21, 15)

		var user = locationSpec.Components.Schemas["User"]
		equalLoc(&user.Common.Locator, 36, 7)
		equalLoc(&user.Properties[0].Schema.Common.Locator, 39, 11)
	})
}
