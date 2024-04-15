package openapi

import (
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/location"
)

// Operation is an OpenAPI Operation.
type Operation struct {
	OperationID string // optional
	Summary     string // optional
	Description string // optional
	Deprecated  bool   // optional

	HTTPMethod  string
	Path        Path
	Parameters  []*Parameter
	RequestBody *RequestBody // optional

	// Security requirements.
	Security SecurityRequirements

	// Operation responses.
	Responses Responses

	XOgenOperationGroup string // Extension field for operation grouping.

	location.Pointer `json:"-" yaml:"-"`
}

// RequestBody of an OpenAPI Operation.
type RequestBody struct {
	Ref         Ref
	Description string
	Required    bool
	Content     map[string]*MediaType

	location.Pointer `json:"-" yaml:"-"`
}

// Header is an OpenAPI Header definition.
type Header = Parameter

// Response is an OpenAPI Response definition.
type Response struct {
	Ref         Ref
	Description string
	Headers     map[string]*Header
	Content     map[string]*MediaType
	// Links map[string]*Link

	location.Pointer `json:"-" yaml:"-"`
}

// Responses contains a list of parsed OpenAPI Responses.
type Responses struct {
	StatusCode map[int]*Response
	Pattern    [5]*Response
	Default    *Response

	location.Pointer `json:"-" yaml:"-"`
}

// Add adds a response to the Responses.
func (r *Responses) Add(pattern string, resp *Response) error {
	switch pattern {
	case "default":
		r.Default = resp
	case "1XX":
		r.Pattern[0] = resp
	case "2XX":
		r.Pattern[1] = resp
	case "3XX":
		r.Pattern[2] = resp
	case "4XX":
		r.Pattern[3] = resp
	case "5XX":
		r.Pattern[4] = resp
	default:
		code, err := strconv.Atoi(pattern)
		if err != nil {
			// Do not return parsing error, it could be a bit confusing.
			return errors.Errorf("invalid response pattern %q", pattern)
		}
		if code < 100 || code > 599 {
			return errors.Errorf("invalid status code: %d", code)
		}
		if r.StatusCode == nil {
			r.StatusCode = make(map[int]*Response)
		}
		r.StatusCode[code] = resp
	}
	return nil
}
