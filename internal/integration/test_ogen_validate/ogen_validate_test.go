package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/validate"
)

func TestOgenValidation(t *testing.T) {
	// Register validators
	err := validate.RegisterValidator("hasPrefix", func(value any, params any) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		prefix, ok := params.(string)
		if !ok {
			return fmt.Errorf("expected string prefix, got %T", params)
		}
		if !strings.HasPrefix(str, prefix) {
			return fmt.Errorf("string %q does not have prefix %q", str, prefix)
		}
		return nil
	})
	require.NoError(t, err)

	err = validate.RegisterValidator("count", func(value any, params any) error {
		obj, ok := value.(UserOther)
		if !ok {
			return fmt.Errorf("expected UserOther, got %T", value)
		}
		paramsMap, ok := params.(map[string]any)
		if !ok {
			return fmt.Errorf("expected map[string]interface{}, got %T", params)
		}
		fieldsMap, ok := paramsMap["fields"].(map[string]any)
		if !ok {
			return fmt.Errorf("expected fields map, got %T", paramsMap["fields"])
		}
		exact, ok := fieldsMap["exact"].(int)
		if !ok {
			return fmt.Errorf("expected exact int, got %T", fieldsMap["exact"])
		}

		// Count fields in UserOther (it's a map type)
		fieldCount := len(obj)
		if fieldCount != exact {
			return fmt.Errorf("expected exactly %d fields, got %d", exact, fieldCount)
		}
		return nil
	})
	require.NoError(t, err)

	// Object-level validator.
	var minAgeCalls int
	err = validate.RegisterValidator("minAge", func(value any, params any) error {
		minAgeCalls++
		user, ok := value.(User)
		if !ok {
			return fmt.Errorf("expected User, got %T", value)
		}
		minAge, ok := params.(int)
		if !ok {
			return fmt.Errorf("expected int min age, got %T", params)
		}
		if user.Age < minAge {
			return fmt.Errorf("age %d is less than %d", user.Age, minAge)
		}
		return nil
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid user with prefix and exact one field in other",
			input:   `{"name": "Mr. John", "age": 30, "email": "john@example.com", "other": {"field1": "value1"}}`,
			wantErr: false,
		},
		{
			name:    "invalid name without prefix",
			input:   `{"name": "John", "age": 30, "email": "john@example.com", "other": {"field1": "value1"}}`,
			wantErr: true,
			errMsg:  "hasPrefix",
		},
		{
			name:    "invalid other with zero fields",
			input:   `{"name": "Mr. John", "age": 30, "email": "john@example.com", "other": {}}`,
			wantErr: true,
			errMsg:  "count",
		},
		{
			name:    "invalid other with multiple fields",
			input:   `{"name": "Mr. John", "age": 30, "email": "john@example.com", "other": {"field1": "value1", "field2": "value2"}}`,
			wantErr: true,
			errMsg:  "count",
		},
		{
			name:    "valid without other field (optional)",
			input:   `{"name": "Mr. John", "age": 30, "email": "john@example.com"}`,
			wantErr: false,
		},
		{
			name:    "invalid user below min age",
			input:   `{"name": "Mr. John", "age": 17, "email": "john@example.com"}`,
			wantErr: true,
			errMsg:  "minAge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user User
			err := user.Decode(jx.DecodeStr(tt.input))
			require.NoError(t, err, "decode should not fail")

			err = user.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("object-level validator is called once per request", func(t *testing.T) {
		srv, err := NewServer(UnimplementedHandler{})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(
			`{"name": "Mr. John", "age": 30, "email": "john@example.com"}`,
		))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		minAgeCalls = 0
		srv.ServeHTTP(rec, req)

		// Request passed validation and reached the handler.
		require.Equal(t, http.StatusNotImplemented, rec.Code)
		require.Equal(t, 1, minAgeCalls)
	})
}
