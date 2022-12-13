package ogenerrors

import (
	"fmt"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

func TestDecodeBodyError(t *testing.T) {
	innerErr := errors.New("inner error")
	decErr := &DecodeBodyError{
		ContentType: "application/json",
		Body:        []byte(`{"foo": "bar"}`),
		Err:         innerErr,
	}

	a := require.New(t)
	a.Equal(innerErr, decErr.Unwrap())
	a.EqualError(decErr, "decode application/json: inner error")
	detailed := fmt.Sprintf("%+v", decErr)
	a.Contains(detailed, `body: {"foo": "bar"}`)
}
