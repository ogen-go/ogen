package jsonpointer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotFoundError_Error(t *testing.T) {
	e := &NotFoundError{
		Pointer: "foobar",
	}
	msg := e.Error()
	require.NotEmpty(t, msg)
	require.Contains(t, msg, e.Pointer)
}
