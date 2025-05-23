package http

import (
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/uri"
)

// Represents the content of an HTTP Accept Header.
// Supports multiple content types (comma separated) and wild cards.
// Does NOT support q-factor weighting (these values are stripped and ignored).
type AcceptHeader []string

// Create a new accept header from the given media types.
// Expects each passed string to be a single media type, e.g. application/json
// or a wildcard of acceptable formats, e.g. application/*.
func AcceptHeaderNew(mediaType ...string) AcceptHeader {
	return AcceptHeader(mediaType)
}

// MarshalText implements encoding.TextMarshaler.
func (s AcceptHeader) MarshalText() ([]byte, error) {
	return []byte(strings.Join(s, ", ")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *AcceptHeader) UnmarshalText(data []byte) error {
	*s = strings.Split(string(data), ",")
	for i, segment := range *s {
		// Remove q-factor weighting
		if semicolonIndex := strings.IndexByte(segment, ';'); semicolonIndex >= 0 {
			segment = segment[:semicolonIndex]
		}
		// Trim spaces to clean up leftovers from comma separation above (spaces are optional there)
		(*s)[i] = strings.TrimSpace(segment)
	}
	return nil
}

func (s AcceptHeader) MatchesContentType(contentType string) bool {
	return slices.ContainsFunc(s, func(pattern string) bool {
		return MatchContentType(pattern, contentType)
	})
}

func (s *AcceptHeader) DecodeURI(d uri.Decoder) error {
	val, err := d.DecodeValue()
	if err != nil {
		return errors.Wrap(err, "decode accept header")
	}
	return s.UnmarshalText([]byte(val))
}

func (s *AcceptHeader) EncodeURI(e uri.Encoder) error {
	val, err := s.MarshalText()
	if err != nil {
		return errors.Wrap(err, "encode accept header")
	}
	return e.EncodeValue(string(val))
}
