package ogenerrors

import (
	"fmt"

	"github.com/go-faster/errors"
)

var _ interface {
	errors.Wrapper
	errors.Formatter
	fmt.Formatter
	error
} = (*DecodeBodyError)(nil)

// DecodeBodyError occurs when request or response body cannot be decoded.
type DecodeBodyError struct {
	ContentType string
	Body        []byte
	Err         error
}

// Unwrap returns child error.
func (d *DecodeBodyError) Unwrap() error {
	return d.Err
}

// FormatError implements errors.Formatter.
func (d *DecodeBodyError) FormatError(p errors.Printer) (next error) {
	p.Printf("decode %s", d.ContentType)
	if p.Detail() {
		p.Printf("body: %s", d.Body)
	}
	return d.Err
}

// Format implements fmt.Formatter.
func (d *DecodeBodyError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *DecodeBodyError) Error() string {
	return fmt.Sprintf("decode %s: %s", d.ContentType, d.Err)
}
