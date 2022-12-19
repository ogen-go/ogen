package rgbatype

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func TestRGBA(t *testing.T) {
	for i, tt := range []struct {
		input   string
		want    RGBA
		wantErr bool
	}{
		{`"rgba(1,1,1,1)"`, RGBA{R: 1, G: 1, B: 1, A: 1}, false},
		{`"rgba(255,0,255,255)"`, RGBA{R: 255, G: 0, B: 255, A: 255}, false},

		// Wrong number format.
		{`"rgba(1,1,1,-1)"`, RGBA{}, true},
		{`"rgba(1,1,1,1.0)"`, RGBA{}, true},
		{`"rgba(1,1,1,256)"`, RGBA{}, true},
		{`"rgba(1,1,1,a)"`, RGBA{}, true},
		// Wrong name of the function.
		{`"rgb(1,1,1,1)"`, RGBA{}, true},
		// Wrong number of arguments.
		{`"rgba(1,1,1)"`, RGBA{}, true},
		// Invalid syntax.
		{`"rgba(1,1,1,1"`, RGBA{}, true},
		// Wrong JSON type.
		{`null`, RGBA{}, true},
		{`true`, RGBA{}, true},
		{`0`, RGBA{}, true},
		// Invalid JSON.
		{`"`, RGBA{}, true},
	} {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var enc JSONRGBAEncoding
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
