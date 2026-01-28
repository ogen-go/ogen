package validate

import (
	"io"
	"net/http"
	"strings"
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

func TestUnexpectedStatusCodeWithResponse(t *testing.T) {
	a := require.New(t)
	body := io.NopCloser(strings.NewReader("error"))
	resp := http.Response{
		StatusCode: 500,
		Body:       body,
	}
	err := UnexpectedStatusCodeWithResponse(&resp)
	body.Close() // emulate deferred close
	var ctErr *UnexpectedStatusCodeError
	a.EqualError(err, "unexpected status code: 500")
	a.ErrorAs(err, &ctErr)
	a.Equal(500, ctErr.StatusCode)

	var sb strings.Builder
	_, err = io.Copy(&sb, ctErr.Payload.Body)
	a.NoError(err)
	a.Equal("error", sb.String())
}
