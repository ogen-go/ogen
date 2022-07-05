package json

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
