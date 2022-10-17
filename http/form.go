package http

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ParseForm is optimized version of http.Request.ParseForm.
//
// Difference from http.Request.ParseForm:
//   - This function does not modify any fields of http.Request. The only copy of the form values is returned.
//   - This function does not check Content-Type header.
func ParseForm(r *http.Request) (url.Values, error) {
	if f := r.PostForm; f != nil {
		return f, nil
	}
	// TODO(tdakkota): implement streaming parser?
	var sb strings.Builder
	if _, err := io.Copy(&sb, r.Body); err != nil {
		return nil, err
	}
	return url.ParseQuery(sb.String())
}
