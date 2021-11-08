package http

import (
	"net/http"
)

// Client represents http client.
type Client interface {
	Do(r *http.Request) (*http.Response, error)
}
