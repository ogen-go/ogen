package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	e := &Error{
		Fields: []FieldError{
			{
				Name:  "foo",
				Error: ErrFieldRequired,
			},
			{
				Name:  "bar",
				Error: ErrFieldRequired,
			},
		},
	}
	msg := e.Error()
	require.NotEmpty(t, msg)
	for _, f := range e.Fields {
		require.Contains(t, msg, f.Name)
	}
}
