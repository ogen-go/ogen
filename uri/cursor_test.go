package uri

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_cursor_readUntil(t *testing.T) {
	a := require.New(t)
	c := &cursor{
		src: `abc,def,ghi`,
	}

	r, err := c.readUntil(',')
	a.NoError(err)
	a.Equal("abc", r)

	r, err = c.readUntil(',')
	a.NoError(err)
	a.Equal("def", r)

	_, err = c.readUntil(',')
	a.ErrorIs(err, io.EOF)
}

func Test_cursor_readValue(t *testing.T) {
	type state struct {
		r       string
		hasNext bool
		err     error
	}
	for i, tt := range []struct {
		input  string
		sep    byte
		states []state
	}{
		{
			input: `abc`,
			sep:   ',',
			states: []state{
				{r: "abc", hasNext: false},
				{err: io.EOF},
			},
		},
		{
			input: `abc,def`,
			sep:   ',',
			states: []state{
				{r: "abc", hasNext: true},
				{r: "def", hasNext: false},
				{err: io.EOF},
			},
		},
		{
			input: `abc,def,ghi`,
			sep:   ',',
			states: []state{
				{r: "abc", hasNext: true},
				{r: "def", hasNext: true},
				{r: "ghi", hasNext: false},
				{err: io.EOF},
			},
		},
	} {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			c := &cursor{
				src: tt.input,
			}

			for _, s := range tt.states {
				r, hasNext, err := c.readValue(tt.sep)
				if s.err != nil {
					a.ErrorIs(err, s.err)
					continue
				}
				a.NoError(err)
				a.Equal(s.r, r)
				a.Equal(s.hasNext, hasNext)
			}
		})
	}
}

func BenchmarkCursor(b *testing.B) {
	input := `abc,def,ghi`

	var (
		sinkErr error
		sink    = make([]string, 0, 3)
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink = sink[:0]
		c := &cursor{src: input}
		sinkErr = parseArray(c, ',', func(d Decoder) error {
			v, err := d.DecodeValue()
			if err != nil {
				return err
			}
			sink = append(sink, v)
			return nil
		})
	}

	if sinkErr != nil {
		b.Fatal(sinkErr)
	}
	if len(sink) != 3 {
		b.Fatalf("unexpected length: %d", len(sink))
	}
}
