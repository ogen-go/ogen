package validate

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"
)

// ErrFieldRequired reports that field is required, but not found.
var ErrFieldRequired = errors.New("field required")

// Error represents validation error.
type Error struct {
	Fields []FieldError
}

// Error implements error.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString("invalid:")
	for i, f := range e.Fields {
		if i != 0 {
			b.WriteRune(',')
		}
		b.WriteRune(' ')
		b.WriteString(f.Name)
		b.WriteString(" (")
		b.WriteString(f.Error.Error())
		b.WriteString(")")
	}

	return b.String()
}

// FieldError is failed validation on field.
type FieldError struct {
	Name  string
	Error error
}

// ErrBodyRequired reports that request body is required but server got empty request.
var ErrBodyRequired = errors.New("body required")

// InvalidContentTypeError reports that decoder got unexpected content type.
type InvalidContentTypeError struct {
	ContentType string
}

// InvalidContentTypeError implements error.
func (i *InvalidContentTypeError) Error() string {
	return fmt.Sprintf("unexpected Content-Type: %s", i.ContentType)
}

// InvalidContentType creates new InvalidContentTypeError.
func InvalidContentType(contentType string) error {
	return &InvalidContentTypeError{
		ContentType: contentType,
	}
}

// UnexpectedStatusCodeError reports that client got unexpected status code.
type UnexpectedStatusCodeError struct {
	StatusCode int
}

// UnexpectedStatusCode creates new UnexpectedStatusCode.
func UnexpectedStatusCode(statusCode int) error {
	return &UnexpectedStatusCodeError{
		StatusCode: statusCode,
	}
}

// UnexpectedStatusCodeError implements error.
func (i *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", i.StatusCode)
}
