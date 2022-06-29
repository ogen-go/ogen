package main

import (
	"net/http"

	"github.com/go-faster/errors"
)

type filterTransport struct {
	next    http.RoundTripper
	allowed map[string]struct{}
}

func (f filterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.Host
	if u := req.URL; host == "" && u != nil {
		host = u.Host
	}
	if _, ok := f.allowed[host]; !ok {
		return nil, errors.Errorf("host %q is not allowed", host)
	}
	return f.next.RoundTrip(req)
}
