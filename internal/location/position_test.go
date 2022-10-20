package location

import (
	"testing"

	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"
)

func TestPosition_Key(t *testing.T) {
	input := `{
  "a": 1,
  "b": {
    "c": 2
  }
}`
	a := require.New(t)

	var node yaml.Node
	a.NoError(yaml.Unmarshal([]byte(input), &node))

	var loc Position
	loc.FromNode(&node)
	a.Equal(1, loc.Line)
	a.Equal(1, loc.Column)
	a.Equal(loc, loc.Key("abc"))

	{
		loc := loc.Key("b")
		a.Equal("!!str", loc.Node.ShortTag())
		a.Equal("b", loc.Node.Value)
		a.Equal(3, loc.Line)
		a.Equal(3, loc.Column)
	}

	{
		loc := loc.Field("b").Key("c")
		a.Equal("!!str", loc.Node.ShortTag())
		a.Equal("c", loc.Node.Value)
		a.Equal(4, loc.Line)
		a.Equal(5, loc.Column)
	}
}

func TestPosition_Field(t *testing.T) {
	input := `{
  "a": 1,
  "b": {
    "c": 2
  }
}`
	a := require.New(t)

	var node yaml.Node
	a.NoError(yaml.Unmarshal([]byte(input), &node))

	var loc Position
	loc.FromNode(&node)
	a.Equal(1, loc.Line)
	a.Equal(1, loc.Column)
	a.Equal(loc, loc.Field("abc"))

	loc = loc.Field("b")
	a.Equal("!!map", loc.Node.ShortTag())
	a.Equal(3, loc.Line)
	a.Equal(8, loc.Column)

	loc = loc.Field("c")
	a.Equal("2", loc.Node.Value)
	a.Equal(4, loc.Line)
	a.Equal(10, loc.Column)
}

func TestPosition_Index(t *testing.T) {
	input := `[
  1,
  2.125
]`
	a := require.New(t)

	var node yaml.Node
	a.NoError(yaml.Unmarshal([]byte(input), &node))

	var loc Position
	loc.FromNode(&node)
	a.Equal(1, loc.Line)
	a.Equal(1, loc.Column)
	a.Equal(loc, loc.Index(-10))
	a.Equal(loc, loc.Index(2))

	{
		loc := loc.Index(0)
		a.Equal("!!int", loc.Node.ShortTag())
		a.Equal("1", loc.Node.Value)
		a.Equal(2, loc.Line)
		a.Equal(3, loc.Column)
	}

	{
		loc := loc.Index(1)
		a.Equal("!!float", loc.Node.ShortTag())
		a.Equal("2.125", loc.Node.Value)
		a.Equal(3, loc.Line)
		a.Equal(3, loc.Column)
	}
}
