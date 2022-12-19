package phonetype

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func TestPhone(t *testing.T) {
	for i, tt := range []struct {
		input   string
		want    Phone
		wantErr bool
	}{
		{`"+1234567890"`, Phone("+1234567890"), false},
		{`"+9876543210"`, Phone("+9876543210"), false},

		// Wrong number format.
		{`"+"`, Phone(""), true},
		{`"123456789"`, Phone(""), true},
		{`"+123456789a"`, Phone(""), true},
		// Wrong JSON type.
		{`null`, Phone(""), true},
		{`true`, Phone(""), true},
		{`0`, Phone(""), true},
		// Invalid JSON.
		{`"`, Phone(""), true},
	} {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var enc JSONPhoneEncoding
			v, err := enc.DecodeJSON(jx.DecodeStr(tt.input))
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.Equal(tt.want, v)

			e := jx.GetEncoder()
			enc.EncodeJSON(e, tt.want)
			a.JSONEq(tt.input, e.String())
		})
	}
}
