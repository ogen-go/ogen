package json

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLines_Collect(t *testing.T) {
	tests := []struct {
		data  string
		lines []int64
	}{
		{"", nil},
		{"a", nil},
		{"abcd", nil},
		{"\n", []int64{0}},
		{"a\n", []int64{1}},
		{"a\n\n", []int64{1, 2}},
		{"\n\n\n", []int64{0, 1, 2}},
		{"a\nb\n", []int64{1, 3}},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := assert.New(t)
			data := []byte(tt.data)

			var l Lines
			l.Collect(data)
			a.Equal(data, l.data)
			for _, offset := range l.lines {
				a.Equal(byte('\n'), data[offset])
			}
		})
	}
}

func TestLines_Line(t *testing.T) {
	tests := []struct {
		input string
		lines []string
	}{
		{"", []string{""}},
		{"\n", []string{""}},
		{"abc", []string{"abc"}},
		{"abc\n", []string{"abc"}},
		{"abc\ndef", []string{"abc", "def"}},
		{"abc\ndef\n", []string{"abc", "def"}},
		{"abc\n" + "def\n" + "ghi\n" + "jkl", []string{"abc", "def", "ghi", "jkl"}},
		{"abc\n" + "def\n" + "ghi\n" + "jkl\n", []string{"abc", "def", "ghi", "jkl"}},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			var lines Lines
			lines.Collect([]byte(tt.input))

			for line, val := range tt.lines {
				start, end := lines.Line(line + 1)
				assert.Equal(t, val, strings.Trim(tt.input[start:end], "\r\n"))
			}
		})
	}
}

func TestLocation_Field(t *testing.T) {
	input := `{
  "a": 1,
  "b": {
    "c": 2
  }
}`
	a := require.New(t)

	var node yaml.Node
	a.NoError(yaml.Unmarshal([]byte(input), &node))

	var loc Location
	loc.fromNode(&node)
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

func TestLocation_Index(t *testing.T) {
	input := `[
  1,
  2.125
]`
	a := require.New(t)

	var node yaml.Node
	a.NoError(yaml.Unmarshal([]byte(input), &node))

	var loc Location
	loc.fromNode(&node)
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
