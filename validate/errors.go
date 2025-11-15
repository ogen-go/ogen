package validate

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ogen-go/ogen/ogenregex"

	"github.com/go-faster/errors"
)

// ErrFieldRequired reports that a field is required, but not found.
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
func (e *InvalidContentTypeError) Error() string {
	return fmt.Sprintf("unexpected Content-Type: %s", e.ContentType)
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
	Payload    *http.Response
}

// UnexpectedStatusCodeWithResponse creates new UnexpectedStatusCode.
func UnexpectedStatusCodeWithResponse(response *http.Response) error {
	return &UnexpectedStatusCodeError{
		StatusCode: response.StatusCode,
		Payload:    response,
	}
}

// UnexpectedStatusCode creates new UnexpectedStatusCode.
//
// Deprecated: client codes generated a while ago used this function.
// Kept here solely for backward compatibility to them.
func UnexpectedStatusCode(statusCode int) error {
	return &UnexpectedStatusCodeError{
		StatusCode: statusCode,
	}
}

// UnexpectedStatusCodeError implements error.
func (e *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}

// ErrNilPointer reports that use Validate, but receiver pointer is nil.
var ErrNilPointer = errors.New("nil pointer")

// MinLengthError reports that len less than minimum.
type MinLengthError struct {
	Len       int
	MinLength int
}

// MinLengthError implements error.
func (e *MinLengthError) Error() string {
	return fmt.Sprintf("len %d less than minimum %d", e.Len, e.MinLength)
}

// MaxLengthError reports that len greater than maximum.
type MaxLengthError struct {
	Len       int
	MaxLength int
}

// MaxLengthError implements error.
func (e *MaxLengthError) Error() string {
	return fmt.Sprintf("len %d greater than maximum %d", e.Len, e.MaxLength)
}

// NoRegexMatchError reports that value have no regexp match.
type NoRegexMatchError struct {
	Pattern ogenregex.Regexp
}

// MaxLengthError implements error.
func (e *NoRegexMatchError) Error() string {
	return fmt.Sprintf("no regex match: %s", e.Pattern.String())
}

// DuplicateItemsError indicates duplicate items in a uniqueItems array.
type DuplicateItemsError struct {
	// Indices contains all indices where duplicates were found.
	// First element is the original, subsequent are duplicates.
	Indices []int
}

// Error implements error.
func (e *DuplicateItemsError) Error() string {
	if len(e.Indices) < 2 {
		return "duplicate items found"
	}
	return fmt.Sprintf("duplicate item found at indices %d and %d",
		e.Indices[0], e.Indices[1])
}

// DepthLimitError indicates nesting depth limit was exceeded.
type DepthLimitError struct {
	// MaxDepth is the configured maximum depth.
	MaxDepth int

	// TypeName is the type being compared when limit was hit.
	TypeName string
}

// Error implements error.
func (e *DepthLimitError) Error() string {
	return fmt.Sprintf("equality check depth limit exceeded for type %s (max: %d)",
		e.TypeName, e.MaxDepth)
}
