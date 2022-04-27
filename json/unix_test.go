package json

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func TestStringUnix(t *testing.T) {
	for _, format := range []struct {
		name    string
		encoder func(e *jx.Encoder, t time.Time)
		decoder func(d *jx.Decoder) (time.Time, error)
		creator func(val int64) time.Time
	}{
		{"Seconds", EncodeUnixSeconds, DecodeUnixSeconds, func(val int64) time.Time {
			return time.Unix(val, 0)
		}},
		{"Nano", EncodeUnixNano, DecodeUnixNano, func(val int64) time.Time {
			return time.Unix(0, val)
		}},
		{"Micro", EncodeUnixMicro, DecodeUnixMicro, time.UnixMicro},
		{"Milli", EncodeUnixMilli, DecodeUnixMilli, time.UnixMilli},
	} {
		format := format
		t.Run(format.name, func(t *testing.T) {
			tests := []struct {
				input   string
				wantVal int64
				wantErr bool
			}{
				{`"0"`, 0, false},
				{`"1"`, 1, false},
				{`"10"`, 10, false},
				{`"10000"`, 10000, false},

				{"1", 0, true},
				{`"foo"`, 0, true},
			}
			for i, tt := range tests {
				tt := tt
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					a := require.New(t)
					d := jx.DecodeStr(tt.input)

					got, err := format.decoder(d)
					if tt.wantErr {
						a.Error(err)
						return
					}
					a.NoError(err)
					wantVal := format.creator(tt.wantVal)
					a.Equal(wantVal, got)

					e := jx.GetEncoder()
					format.encoder(e, wantVal)

					d.ResetBytes(e.Bytes())
					got2, err := format.decoder(d)
					a.NoError(err)
					a.Equal(wantVal, got2)
				})
			}
		})
	}
}

func TestDecodeUnixNano(t *testing.T) {
	got, err := DecodeUnixNano(jx.DecodeStr(`"1586960586000000000"`))
	require.NoError(t, err)
	want := time.Date(2020, 04, 15, 14, 23, 06, 0, time.UTC)
	require.True(t, want.Equal(got))
}
