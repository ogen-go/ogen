package http

import (
	"fmt"

	"github.com/go-faster/errors"
)

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
