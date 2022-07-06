package ogen

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

var (
	//go:embed _testdata/location/location_spec.json
	locationSpecJSON string
	//go:embed _testdata/location/location_spec.yml
	locationSpecYAML string
)

func TestLocation(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		a := assert.New(t)
		equalLoc := func(l ogenjson.Locatable, line, column int64, ptr string) {
			t.Helper()

			loc, ok := l.Location()
			a.Truef(ok, "ptr: %s", ptr)
			type location struct {
				Line, Column int64
				Ptr          string
			}
			a.Equalf(
				location{line, column, ptr},
				location{loc.Line, loc.Column, loc.JSONPointer},
				"ptr: %s", ptr,
			)
		}

		var locationSpec Spec
		require.NoError(t, unmarshalJSON([]byte(locationSpecJSON), &locationSpec))
		var (
			foo         = locationSpec.Paths["/foo"]
			fooLocation = "/paths/~1foo"

			post         = foo.Post
			postLocation = fooLocation + "/post"

			body         = post.RequestBody
			bodyLocation = postLocation + "/requestBody"

			media         = body.Content["application/json"]
			mediaLocation = bodyLocation + "/content/application~1json"

			schema         = media.Schema
			schemaLocation = mediaLocation + "/schema"
		)
		// Compare PathItem and Operation.
		equalLoc(foo, 8, 13, fooLocation)
		equalLoc(post, 9, 15, postLocation)

		// Compare Parameters.
		// FIXME(tdakkota): For some reason, go-json-experiment does not add array index.
		equalLoc(post.Parameters[0], 11, 11, postLocation+"/parameters")

		// Compare RequestBody.
		equalLoc(body, 19, 24, bodyLocation)
		equalLoc(&media, 21, 33, mediaLocation)
		equalLoc(schema, 22, 25, schemaLocation)

		var user = locationSpec.Components.Schemas["User"]
		equalLoc(user, 45, 15, "/components/schemas/User")
		equalLoc(user.Properties[0].Schema, 48, 19, "/components/schemas/User/properties/name")
	})
	t.Run("YAML", func(t *testing.T) {
		a := assert.New(t)
		equalLoc := func(l ogenjson.Locatable, line, column int64) {
			t.Helper()

			loc, ok := l.Location()
			a.True(ok)
			type location struct {
				Line, Column int64
			}
			a.Equal(
				location{line, column},
				location{loc.Line, loc.Column},
			)
		}

		var locationSpec Spec
		require.NoError(t, unmarshalYAML([]byte(locationSpecYAML), &locationSpec))
		var (
			foo    = locationSpec.Paths["/foo"]
			post   = foo.Post
			body   = post.RequestBody
			media  = body.Content["application/json"]
			schema = media.Schema
		)
		// Compare PathItem and Operation.
		equalLoc(foo, 6, 3)
		equalLoc(post, 7, 5)

		// Compare Parameters.
		equalLoc(post.Parameters[0], 9, 11)

		// Compare RequestBody.
		equalLoc(body, 13, 7)
		equalLoc(&media, 15, 11)
		equalLoc(schema, 16, 13)

		var user = locationSpec.Components.Schemas["User"]
		equalLoc(user, 27, 5)
		equalLoc(user.Properties[0].Schema, 30, 9)
	})
}
