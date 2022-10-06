package naming

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAfterDot(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Dot is the second character: return the part after the dot.
		{"a.b", "b"},
		// No dot: return the whole string.
		{"", ""},
		{"a", "a"},
		// Dot is the first character: return the whole string.
		{".", "."},
		{".ab", ".ab"},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.want, AfterDot(tt.input))
		})
	}
}
