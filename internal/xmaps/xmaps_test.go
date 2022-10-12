package xmaps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]struct{}
		want []string
	}{
		{"Nil", nil, []string{}},
		{"Empty", map[string]struct{}{}, []string{}},
		{"One", map[string]struct{}{"a": {}}, []string{"a"}},
		{"Two", map[string]struct{}{"a": {}, "b": {}}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, SortedKeys(tt.m))
		})
	}
}
