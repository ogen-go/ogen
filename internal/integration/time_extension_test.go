package integration

import (
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_time_extension"
)

func TestTimeExtension(t *testing.T) {
	input := `{ "date": "04/03/2001", "time": "1:23AM", "dateTime": "2001-03-04T01:23:45.123456789-07:00", "alias": "04/03/2001 1:23:45AM" }`

	t.Run("Required", func(t *testing.T) {
		expected := api.RequiredOK{
			Date:     time.Date(2001, 3, 4, 0, 0, 0, 0, time.UTC),
			Time:     time.Date(0, 1, 1, 1, 23, 0, 0, time.UTC),
			DateTime: time.Date(2001, 3, 4, 1, 23, 45, 123456789, time.FixedZone("", -7*60*60)),
			Alias:    api.Alias(time.Date(2001, 3, 4, 1, 23, 45, 0, time.UTC)),
		}

		a := require.New(t)
		var p api.RequiredOK
		a.NoError(p.Decode(jx.DecodeStr(input)))
		a.Equal(p, expected)

		out, err := p.MarshalJSON()
		a.NoError(err)
		a.JSONEq(input, string(out))
	})

	t.Run("Optional", func(t *testing.T) {
		expected := api.OptionalOK{
			Date:     api.NewOptDate(time.Date(2001, 3, 4, 0, 0, 0, 0, time.UTC)),
			Time:     api.NewOptTime(time.Date(0, 1, 1, 1, 23, 0, 0, time.UTC)),
			DateTime: api.NewOptDateTime(time.Date(2001, 3, 4, 1, 23, 45, 123456789, time.FixedZone("", -7*60*60))),
			Alias:    api.NewOptAlias(api.Alias(time.Date(2001, 3, 4, 1, 23, 45, 0, time.UTC))),
		}

		a := require.New(t)
		var p api.OptionalOK
		a.NoError(p.Decode(jx.DecodeStr(input)))
		a.Equal(p, expected)

		out, err := p.MarshalJSON()
		a.NoError(err)
		a.JSONEq(input, string(out))
	})

	t.Run("Defaults", func(t *testing.T) {
		expected := api.OptionalOK{
			Date:     api.NewOptDate(time.Date(2001, 3, 4, 0, 0, 0, 0, time.UTC)),
			Time:     api.NewOptTime(time.Date(0, 1, 1, 1, 23, 0, 0, time.UTC)),
			DateTime: api.NewOptDateTime(time.Date(2001, 3, 4, 1, 23, 45, 123456789, time.FixedZone("", -7*60*60))),
			Alias:    api.NewOptAlias(api.Alias(time.Date(2001, 3, 4, 1, 23, 45, 0, time.UTC))),
		}

		a := require.New(t)
		var p api.OptionalOK
		a.NoError(p.Decode(jx.DecodeStr(`{}`)))
		a.Equal(p, expected)
	})
}
