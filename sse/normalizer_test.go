package sse

import (
	"io"
	"strings"
	"testing"
)

func TestNewlineNormalizer(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "some_plaintext", want: "some_plaintext"},
		{input: "some\rplaintext", want: "some\nplaintext"},
		{input: "some\r\nplaintext", want: "some\r\nplaintext"},
		{input: "some\nplaintext", want: "some\nplaintext"},
		{input: "a\r\nb\rc\n", want: "a\r\nb\nc\n"},
		{input: "123456\r", want: "123456\n"},
		{input: "123456\rX", want: "123456\nX"},
		{input: "123456\r\nX", want: "123456\r\nX"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			const bufSize = 7
			var r = &newlineNormalizer{r: strings.NewReader(tt.input)}
			got, err := readSmallChunks(r, bufSize)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("Read() = %q, want %q", got, tt.want)
			}
		})
	}
}

func readSmallChunks(r io.Reader, size int) (string, error) {
	var b strings.Builder
	buf := make([]byte, size)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
		}
		if err == io.EOF {
			return b.String(), nil
		}
		if err != nil {
			return b.String(), err
		}
	}
}
