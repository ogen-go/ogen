// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/ogenerrors"
)

// SecurityHandler is handler for security parameters.
type SecurityHandler interface {
	// HandleAPIKey handles api_key security.
	HandleAPIKey(ctx context.Context, operationName OperationName, t APIKey) (context.Context, error)
}

func findAuthorization(h http.Header, prefix string) (string, bool) {
	v, ok := h["Authorization"]
	if !ok {
		return "", false
	}
	for _, vv := range v {
		scheme, value, ok := strings.Cut(vv, " ")
		if !ok || !strings.EqualFold(scheme, prefix) {
			continue
		}
		return value, true
	}
	return "", false
}

var operationRolesAPIKey = map[string][]string{
	SecurityTestOperation: []string{},
}

func (s *Server) securityAPIKey(ctx context.Context, operationName OperationName, req *http.Request) (context.Context, bool, error) {
	var t APIKey
	const parameterName = "Api_key"
	value := req.Header.Get(parameterName)
	if value == "" {
		return ctx, false, nil
	}
	t.APIKey = value
	t.Roles = operationRolesAPIKey[operationName]
	rctx, err := s.sec.HandleAPIKey(ctx, operationName, t)
	if errors.Is(err, ogenerrors.ErrSkipServerSecurity) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return rctx, true, err
}
