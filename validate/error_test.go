package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	a := require.New(t)

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
	a.NotEmpty(msg)
	for _, f := range e.Fields {
		a.Contains(msg, f.Name)
	}
}

func TestInvalidContentType(t *testing.T) {
	a := require.New(t)
	err := InvalidContentType("application/json")
	var ctErr *InvalidContentTypeError
	a.EqualError(err, "unexpected Content-Type: application/json")
	a.ErrorAs(err, &ctErr)
	a.Equal("application/json", ctErr.ContentType)
}

func TestUnexpectedStatusCode(t *testing.T) {
	a := require.New(t)
	err := UnexpectedStatusCode(500)
	var ctErr *UnexpectedStatusCodeError
	a.EqualError(err, "unexpected status code: 500")
	a.ErrorAs(err, &ctErr)
	a.Equal(500, ctErr.StatusCode)
}
