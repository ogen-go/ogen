package integration

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/sample_api"
	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/validate"
)

func TestValidateRequired(t *testing.T) {
	data := func() json.Unmarshaler {
		return &api.Data{}
	}
	arrayTest := func() json.Unmarshaler {
		return &api.ArrayTest{}
	}
	mapWithProperties := func() json.Unmarshaler {
		return &api.MapWithProperties{}
	}
	required := func(fields ...string) (r []validate.FieldError) {
		for _, f := range fields {
			r = append(r, validate.FieldError{
				Name:  f,
				Error: validate.ErrFieldRequired,
			})
		}
		return r
	}
	for i, tc := range []struct {
		Input   string
		Decoder func() json.Unmarshaler
		Error   []validate.FieldError
	}{
		{
			`{}`,
			data,
			required("id", "description", "email", "hostname", "format"),
		},
		{
			`{"email": "aboba"}`,
			data,
			required("id", "description", "hostname", "format"),
		},
		{
			`{"id":10, "email": "aboba"}`,
			data,
			required("description", "hostname", "format"),
		},
		{
			`{}`,
			arrayTest,
			required("required", "nullable_required"),
		},
		{
			`{"required": []}`,
			arrayTest,
			required("nullable_required"),
		},
		{
			`{"nullable_required": []}`,
			arrayTest,
			required("required"),
		},
		{
			`{}`,
			mapWithProperties,
			required("required"),
		},
		{
			`{"random_field": "string"}`,
			mapWithProperties,
			required("required"),
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			err := tc.Decoder().Decode(jx.DecodeStr(tc.Input))
			if len(tc.Error) > 0 {
				var e *validate.Error
				require.ErrorAs(t, err, &e)
				require.Equal(t, tc.Error, e.Fields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMap(t *testing.T) {
	for i, tc := range []struct {
		Input string
		Error bool
	}{
		{
			`{}`,
			false,
		},
		{
			`{"a": ""}`,
			true,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			m := api.StringMap{}
			require.NoError(t, m.Decode(jx.DecodeStr(tc.Input)))

			checker := require.NoError
			if tc.Error {
				checker = require.Error
			}
			checker(t, m.Validate())
		})
	}
}

func TestValidateFloat(t *testing.T) {
	for i, tc := range []struct {
		Input string
		Error bool
	}{
		{
			`{"minmax": 1.0, "multipleOf": 5.0}`,
			true,
		},
		{
			`{"minmax": 1.4, "multipleOf": 5.0}`,
			true,
		},
		{
			`{"minmax": 2.7, "multipleOf": 5.0}`,
			true,
		},
		{
			`{"minmax": 2.0, "multipleOf": 5.0}`,
			false,
		},
		{
			`{"minmax": 2.0, "multipleOf": 15.0}`,
			false,
		},
		{
			`{"minmax": 2.0, "multipleOf": 0.1}`,
			true,
		},
		{
			`{"minmax": 2.0, "multipleOf": 10.1}`,
			true,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			m := api.TestFloatValidation{}
			require.NoError(t, m.Decode(jx.DecodeStr(tc.Input)))

			checker := require.NoError
			if tc.Error {
				checker = require.Error
			}
			checker(t, m.Validate())
		})
	}
}

func TestValidateUniqueItems(t *testing.T) {
	for i, tc := range []struct {
		Input string
		Error bool
	}{
		{
			`{"required_unique": []}`,
			false,
		},
		{
			`{"required_unique": ["a"]}`,
			false,
		},
		{
			`{"required_unique": ["a", "b"]}`,
			false,
		},
		{
			`{"required_unique": ["a", "a"]}`,
			true,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			m := api.UniqueItemsTest{}
			require.NoError(t, m.Decode(jx.DecodeStr(tc.Input)))

			checker := require.NoError
			if tc.Error {
				checker = require.Error
			}
			checker(t, m.Validate())
		})
	}
}

func TestArrayLengthValidation(t *testing.T) {
	decodeValidate := func(input string, r *api.Issue1461) error {
		if err := r.Decode(jx.DecodeStr(input)); err != nil {
			return err
		}
		if err := r.Validate(); err != nil {
			return err
		}
		return nil
	}

	for i, tc := range []struct {
		Input string
		Error string
	}{
		// Required test cases
		{
			`{
				"requiredTest": {}
			}`,
			`decode Issue1461: callback: decode field "requiredTest": invalid: banana (field required)`,
		},
		{
			`{
				"requiredTest": {"banana": []}
			}`,
			`invalid: requiredTest (invalid: banana (array: len 0 less than minimum 2))`,
		},
		{
			`{
				"requiredTest": {"banana": ["a"]}
			}`,
			`invalid: requiredTest (invalid: banana (array: len 1 less than minimum 2))`,
		},
		{
			`{
				"requiredTest": {"banana": ["a", "b"]}
			}`,
			``,
		},

		// Optional test cases
		{
			`{
				"optionalTest": {}
			}`,
			``,
		},
		{
			`{
				"optionalTest": {"banana": []}
			}`,
			`invalid: optionalTest (invalid: banana (array: len 0 less than minimum 2))`,
		},
		{
			`{
				"optionalTest": {"banana": ["a"]}
			}`,
			`invalid: optionalTest (invalid: banana (array: len 1 less than minimum 2))`,
		},
		{
			`{
				"optionalTest": {"banana": ["a", "b"]}
			}`,
			``,
		},

		// Nullable test cases
		{
			`{
				"nullableTest": {}
			}`,
			`decode Issue1461: callback: decode field "nullableTest": invalid: banana (field required)`,
		},
		{
			`{
				"nullableTest": {"banana": []}
			}`,
			`invalid: nullableTest (invalid: banana (array: len 0 less than minimum 2))`,
		},
		{
			`{
				"nullableTest": {"banana": ["a"]}
			}`,
			`invalid: nullableTest (invalid: banana (array: len 1 less than minimum 2))`,
		},
		{
			`{
				"nullableTest": {"banana": null}
			}`,
			``,
		},
		{
			`{
				"nullableTest": {"banana": ["a", "b"]}
			}`,
			``,
		},

		// Nullable optional test cases
		{
			`{
				"nullableOptionalTest": {}
			}`,
			``,
		},
		{
			`{
				"nullableOptionalTest": {"banana": []}
			}`,
			`invalid: nullableOptionalTest (invalid: banana (array: len 0 less than minimum 2))`,
		},
		{
			`{
				"nullableOptionalTest": {"banana": ["a"]}
			}`,
			`invalid: nullableOptionalTest (invalid: banana (array: len 1 less than minimum 2))`,
		},
		{
			`{
				"nullableOptionalTest": {"banana": null}
			}`,
			``,
		},
		{
			`{
				"nullableOptionalTest": {"banana": ["a", "b"]}
			}`,
			``,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil || t.Failed() {
					t.Logf("Input:\n%s", tc.Input)
				}
			}()

			m := api.Issue1461{}
			err := decodeValidate(tc.Input, &m)

			if e := tc.Error; e != "" {
				require.EqualError(t, err, e)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
