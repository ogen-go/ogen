package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_custom_validation"
	"github.com/ogen-go/ogen/validate"
)

func TestCELValidation(t *testing.T) {
	// Register a custom validator for testing
	err := validate.RegisterValidator("custom", func(value any, params string) error {
		if params == "test-string-validator" {
			if str, ok := value.(string); ok {
				if len(str) < 3 {
					return fmt.Errorf("string too short (custom validator)")
				}
				if strings.Contains(str, "forbidden") {
					return fmt.Errorf("contains forbidden word (custom validator)")
				}
			}
		}
		return nil
	})
	require.NoError(t, err)

	// Register CEL validator for testing
	err = validate.RegisterValidator("cel", func(value any, params string) error {
		cel := CEL{}
		if err := cel.SetExpression(params); err != nil {
			return err
		}
		return cel.Validate(value)
	})
	require.NoError(t, err)
	for i, tc := range []struct {
		Name    string
		Input   string
		Decoder func() interface{ Decode(*jx.Decoder) error; Validate() error }
		Error   bool
	}{
		{
			Name:  "Valid User",
			Input: `{"name": "John Doe", "age": 25, "email": "john@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: false,
		},
		{
			Name:  "User with short name",
			Input: `{"name": "Jo", "age": 25, "email": "john@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "User with age too young",
			Input: `{"name": "John Doe", "age": 16, "email": "john@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "User with age too old",
			Input: `{"name": "John Doe", "age": 130, "email": "john@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "User with invalid email (no @)",
			Input: `{"name": "John Doe", "age": 25, "email": "johnexample.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "User with invalid email (no .)",
			Input: `{"name": "John Doe", "age": 25, "email": "john@example"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "Young user with short name (object-level validation)",
			Input: `{"name": "Bob", "age": 20, "email": "bob@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "Young user with long name (object-level validation passes)",
			Input: `{"name": "Robert", "age": 20, "email": "robert@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: false,
		},
		{
			Name:  "User with forbidden word (custom validator)",
			Input: `{"name": "forbidden", "age": 25, "email": "test@example.com"}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.User{}
			},
			Error: true,
		},
		{
			Name:  "Valid Product",
			Input: `{"name": "Laptop", "price": 999.99, "categories": ["electronics", "computers"]}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: false,
		},
		{
			Name:  "Product with short name",
			Input: `{"name": "PC", "price": 999.99, "categories": ["electronics"]}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: true,
		},
		{
			Name:  "Product with negative price",
			Input: `{"name": "Laptop", "price": -100.0, "categories": ["electronics"]}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: true,
		},
		{
			Name:  "Product with price too high",
			Input: `{"name": "Laptop", "price": 15000.0, "categories": ["electronics"]}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: true,
		},
		{
			Name:  "Product with empty categories",
			Input: `{"name": "Laptop", "price": 999.99, "categories": []}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: true,
		},
		{
			Name:  "Product with too many categories",
			Input: `{"name": "Laptop", "price": 999.99, "categories": ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: true,
		},
		{
			Name:  "Product with empty category name",
			Input: `{"name": "Laptop", "price": 999.99, "categories": [""]}`,
			Decoder: func() interface{ Decode(*jx.Decoder) error; Validate() error } {
				return &api.Product{}
			},
			Error: true,
		},
	} {
		t.Run(fmt.Sprintf("%s_%d", tc.Name, i+1), func(t *testing.T) {
			decoder := tc.Decoder()
			err := decoder.Decode(jx.DecodeStr(tc.Input))
			require.NoError(t, err, "JSON decoding should not fail")

			err = decoder.Validate()
			if tc.Error {
				require.Error(t, err, "Validation should fail")
				t.Logf("Validation error: %s", err.Error())
			} else {
				require.NoError(t, err, "Validation should pass")
			}
		})
	}
}
