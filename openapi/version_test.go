package openapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion_UnmarshalText(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"3", Version{Major: 3}, false},
		{"3.0", Version{Major: 3, Minor: 0}, false},
		{"3.1", Version{Major: 3, Minor: 1}, false},
		{"3.1.0", Version{Major: 3, Minor: 1, Patch: 0}, false},
		{"3.1.1", Version{Major: 3, Minor: 1, Patch: 1}, false},
		{"3.2.0", Version{Major: 3, Minor: 2, Patch: 0}, false},

		{"", Version{}, true},
		{" ", Version{}, true},
		{".", Version{}, true},
		{"..", Version{}, true},
		{"3.", Version{}, true},
		{"3.1.", Version{}, true},
		{".1.", Version{}, true},
		{"3..0", Version{}, true},
		{"3.3 .0", Version{}, true},
		{"3.3.0.4", Version{}, true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var v Version
			err := v.UnmarshalText([]byte(tt.input))
			if tt.wantErr {
				a.Error(err)
				a.Zero(v)
				return
			}
			a.NoError(err)
			a.Equal(tt.want, v)

			// Ensure that the version can be marshaled back to the same string.
			r, err := v.MarshalText()
			a.NoError(err)
			var v2 Version
			a.NoError(v2.UnmarshalText(r))
			a.Equal(v, v2)
		})
	}
}
