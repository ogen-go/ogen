package http

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchContentType(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"*/*", "application/json", true},
		{"*/**", "application/json", true},
		{"application/*", "application/json", true},
		{"application/*", "application/xml", true},
		{"*/json", "application/json", true},
		{"*/json", "text/json", true},
		{"application/*", "text/json", false},
		{"text/*", "application/json", false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.want, MatchContentType(tt.pattern, tt.value))
		})
	}
}
