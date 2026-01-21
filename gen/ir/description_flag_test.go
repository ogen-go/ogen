package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettyDocEnabledFlag(t *testing.T) {
	// Save original state to restore later
	originalState := prettyDocEnabled

	// Test with pretty documentation enabled
	SetPrettyDoc(true)
	require.True(t, IsPrettyDocEnabled())
	doc := "   this is a test.   "
	prettyResult := prettyDoc(doc, "")
	require.Equal(t, "This is a test.", prettyResult[0])

	// Test with pretty documentation disabled
	SetPrettyDoc(false)
	require.False(t, IsPrettyDocEnabled())
	verbatimResult := prettyDoc(doc, "")
	require.Equal(t, "   this is a test.   ", verbatimResult[0])

	// Restore original state
	SetPrettyDoc(originalState)
}
