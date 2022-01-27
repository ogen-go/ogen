package ogen

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
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
