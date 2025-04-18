package uri

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryDecoder_HasParam(t *testing.T) {
	tests := []struct {
		Input   url.Values
		Cfg     QueryParameterDecodingConfig
		WantErr string
	}{
		// QueryStyleDeepObject + Explode
		{
			Input: url.Values{},
			Cfg: QueryParameterDecodingConfig{
				Name:    "object",
				Style:   QueryStyleDeepObject,
				Explode: true,
				Fields: []QueryParameterObjectField{
					{"foo", true},
				},
			},
			WantErr: "invalid: object[foo] (field required)",
		},
		{
			Input: url.Values{},
			Cfg: QueryParameterDecodingConfig{
				Name:    "object",
				Style:   QueryStyleDeepObject,
				Explode: true,
				Fields: []QueryParameterObjectField{
					{"foo", false},
				},
			},
			WantErr: "none of parameters ([{Name:foo Required:false}]) are set",
		},
		{
			Input: url.Values{
				"object[foo]": []string{"bar"},
			},
			Cfg: QueryParameterDecodingConfig{
				Name:    "object",
				Style:   QueryStyleDeepObject,
				Explode: true,
				Fields: []QueryParameterObjectField{
					{"foo", true},
				},
			},
		},

		// QueryStyleForm + Explode
		{
			Input: url.Values{},
			Cfg: QueryParameterDecodingConfig{
				Name:    "foo",
				Style:   QueryStyleForm,
				Explode: true,
				Fields: []QueryParameterObjectField{
					{"foo", true},
				},
			},
			WantErr: "invalid: foo (field required)",
		},
		{
			Input: url.Values{},
			Cfg: QueryParameterDecodingConfig{
				Name:    "foo",
				Style:   QueryStyleForm,
				Explode: true,
				Fields: []QueryParameterObjectField{
					{"foo", false},
				},
			},
			WantErr: "none of parameters ([{Name:foo Required:false}]) are set",
		},
		{
			Input: url.Values{
				"foo": []string{"bar"},
			},
			Cfg: QueryParameterDecodingConfig{
				Name:    "foo",
				Style:   QueryStyleForm,
				Explode: true,
				Fields: []QueryParameterObjectField{
					{"foo", true},
				},
			},
		},

		// Other
		{
			Cfg: QueryParameterDecodingConfig{
				Name:  "foo",
				Style: QueryStyleForm,
			},
			WantErr: "query parameter \"foo\" not set",
		},
		{
			Input: url.Values{
				"foo": []string{"bar"},
			},
			Cfg: QueryParameterDecodingConfig{
				Name:  "foo",
				Style: QueryStyleForm,
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			d := NewQueryDecoder(tt.Input)

			err := d.HasParam(tt.Cfg)
			if tt.WantErr != "" {
				require.EqualError(t, err, tt.WantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
