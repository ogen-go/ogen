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
