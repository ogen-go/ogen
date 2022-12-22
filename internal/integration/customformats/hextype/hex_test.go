package hextype

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func TestHex(t *testing.T) {
	for i, tt := range []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{`"1"`, 1, false},
		{`"a"`, 10, false},
		{`"ff"`, 255, false},

		// Wrong number format.
		{`"-"`, 0, true},
		{`"x"`, 0, true},
		{`"fg"`, 0, true},
		{`"0.0"`, 0, true},
		// Wrong JSON type.
		{`null`, 0, true},
		{`true`, 0, true},
		{`0`, 0, true},
		// Invalid JSON.
		{`"`, 0, true},
	} {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var enc JSONHexEncoding
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
