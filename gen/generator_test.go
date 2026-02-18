package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
)

func TestCommentLineLimitPropagation(t *testing.T) {
	// Create a simple spec for testing
	spec := &ogen.Spec{
		OpenAPI: "3.0.0",
		Info: ogen.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// Test cases with different comment line limit values
	tests := []struct {
		name        string
		lineLimit   int
		expectLimit int
	}{
		{
			name:        "default_value",
			lineLimit:   0,
			expectLimit: 100, // Should use default value
		},
		{
			name:        "custom_value",
			lineLimit:   50,
			expectLimit: 50,
		},
		{
			name:        "disable_wrapping",
			lineLimit:   -1,
			expectLimit: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original line limit to restore after the test
			originalLimit := ir.GetLineLimit()
			defer ir.SetLineLimit(originalLimit)

			// Create options with the test case's line limit
			opts := Options{
				Generator: GenerateOptions{
					CommentLineLimit: tt.lineLimit,
				},
			}

			// Create a new generator
			_, err := NewGenerator(spec, opts)
			require.NoError(t, err)

			// Verify the line limit was set correctly
			actualLimit := ir.GetLineLimit()
			require.Equal(t, tt.expectLimit, actualLimit,
				"Line limit should be set to %d, but got %d", tt.expectLimit, actualLimit)
		})
	}
}
