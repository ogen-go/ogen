package internal

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
)

func TestDefault(t *testing.T) {
	type mutator func(*api.DefaultTest)

	value := func(cb mutator) api.DefaultTest {
		defaultValue := api.DefaultTest{
			Str: api.OptString{
				Value: "str",
				Set:   true,
			},
			NullStr: api.OptNilString{
				Value: "",
				Set:   false,
				Null:  true,
			},
			Enum: api.OptDefaultTestEnum{
				Value: "big",
				Set:   true,
			},
			UUID: api.OptUUID{
				Value: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Set:   true,
			},
			IP: api.OptIP{
				Value: netip.AddrFrom4([4]byte{1, 1, 1, 1}),
				Set:   true,
			},
			IPV4: api.OptIPv4{
				Value: netip.AddrFrom4([4]byte{1, 1, 1, 1}),
				Set:   true,
			},
			IPV6: api.OptIPv6{
				Value: netip.MustParseAddr("2001:db8:85a3::8a2e:370:7334"),
				Set:   true,
			},
			URI: api.OptURI{
				Value: url.URL{
					Scheme: "s3",
					Host:   "foo",
					Path:   "/baz",
				},
				Set: true,
			},
			Birthday: api.OptDate{
				Value: time.Date(2011, time.October, 10, 0, 0, 0, 0, time.UTC),
				Set:   true,
			},
			Rate: api.OptDuration{
				Value: 5 * time.Second,
				Set:   true,
			},
			Email: api.OptString{
				Value: "foo@example.com",
				Set:   true,
			},
			Hostname: api.OptString{
				Value: "example.org",
				Set:   true,
			},
			Format: api.OptString{
				Value: "1-2",
				Set:   true,
			},
			Base64: func() []byte {
				b, err := jx.DecodeStr(`"aGVsbG8sIHdvcmxkIQ=="`).Base64()
				if err != nil {
					panic(err)
				}
				return b
			}(),
		}

		cb(&defaultValue)
		return defaultValue
	}

	for i, tc := range []struct {
		Input    string
		Expected mutator
		Error    bool
	}{
		{
			Input: `{"required":"required"}`,
			Expected: func(test *api.DefaultTest) {
				test.Required = "required"
			},
			Error: false,
		},
		{
			Input: `{"required":"required", "enum": "smol"}`,
			Expected: func(test *api.DefaultTest) {
				test.Required = "required"
				test.Enum.SetTo(api.DefaultTestEnumSmol)
			},
			Error: false,
		},
		{
			Input: `{"required":"required", "ip": "8.8.8.8", "ip_v4": "8.8.8.8"}`,
			Expected: func(test *api.DefaultTest) {
				test.Required = "required"
				test.IP.SetTo(netip.AddrFrom4([4]byte{8, 8, 8, 8}))
				test.IPV4.SetTo(test.IP.Value)
			},
			Error: false,
		},
		{
			Input:    `{}`,
			Expected: nil,
			Error:    true,
		},
	} {
		// Make range value copy to prevent data races.
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			r := api.DefaultTest{}
			if err := r.Decode(jx.DecodeStr(tc.Input)); tc.Error {
				require.Error(t, err)
			} else {
				require.Equal(t, value(tc.Expected), r)
				require.NoError(t, err)
			}
		})
	}
}
