package naming

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"\x87", "\x87"},
		{"int8", "Int8"},
		{"Int8", "Int8"},
		{"хлеб", "Хлеб"},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.want, Capitalize(tt.input))
		})
	}
}
