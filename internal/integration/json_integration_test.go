package integration

import (
	"net/netip"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/conv"
	api "github.com/ogen-go/ogen/internal/integration/sample_api"
	"github.com/ogen-go/ogen/internal/integration/techempower"
	singleEndpoint "github.com/ogen-go/ogen/internal/integration/test_single_endpoint"
	"github.com/ogen-go/ogen/json"
)

// Ensure that convenient errors triggered on single endpoint.
//
// See https://github.com/ogen-go/ogen/issues/63.
var (
	_                      = singleEndpoint.ErrorStatusCode{}
	_ singleEndpoint.Error = singleEndpoint.ErrorStatusCode{}.Response
	_                      = singleEndpoint.Handler.NewError
)

func TestExampleJSON(t *testing.T) {
	t.Parallel()
	date := time.Date(2011, 10, 10, 7, 12, 34, 4125, time.UTC)
	stringMap := api.StringStringMap{
		"prop": api.StringMap{
			"i\nhate": "OpenAPI specification",
		},
	}
	pet := api.Pet{
		Friends:  []api.Pet{},
		Birthday: conv.Date(date),
		ID:       42,
		Name:     "SomePet",
		TestArray1: [][]string{
			{
				"Foo", "Bar",
			},
			{
				"Baz",
			},
		},
		TestMap: api.NewOptStringStringMap(stringMap),
		TestMapWithProps: api.NewOptMapWithProperties(api.MapWithProperties{
			Required: 10,
			AdditionalProps: map[string]string{
				"data": "data",
			},
		}),
		Nickname:     api.NewNilString("Nick"),
		NullStr:      api.NewOptNilString("Bar"),
		Rate:         time.Second,
		Tag:          api.NewOptUUID(uuid.New()),
		TestDate:     api.NewOptDate(conv.Date(date)),
		TestDateTime: api.NewOptDateTime(conv.DateTime(date)),
		TestDuration: api.NewOptDuration(time.Minute),
		TestFloat1:   api.NewOptFloat64(1.0),
		TestInteger1: api.NewOptInt(10),
		TestTime:     api.NewOptTime(conv.Time(date)),
		UniqueID:     uuid.New(),
		URI:          url.URL{Scheme: "s3", Host: "foo", Path: "bar"},
		IP:           netip.AddrFrom4([4]byte{127, 0, 0, 1}),
		IPV4:         netip.AddrFrom4([4]byte{127, 0, 0, 1}),
		IPV6:         netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		Next: api.NewOptData(api.Data{
			Description: api.NewDescriptionSimpleDataDescription(api.DescriptionSimple{
				Description: "foo",
			}),
			ID: api.NewIntID(10),
		}),
	}

	for _, tc := range []struct {
		Name  string
		Value interface {
			json.Marshaler
		}
		Expected string
	}{
		{
			Name:  "Pet",
			Value: pet,
		},
		{
			Name: "PetWithPrimary",
			Value: func(input api.Pet) (r api.Pet) {
				r = input
				r.Primary = &input
				return r
			}(pet),
		},
		{
			Name:  "OptPetSet",
			Value: api.NewOptPet(pet),
		},
		{
			Name:     "PetName",
			Value:    api.PetName("boba"),
			Expected: `"boba"`,
		},
		{
			Name:     "OptPetName",
			Value:    api.NewOptPetName("boba"),
			Expected: `"boba"`,
		},
		{
			Name:     "PetType",
			Value:    api.PetTypeFifa,
			Expected: strconv.Quote(string(api.PetTypeFifa)),
		},
		{
			Name:     "OptPetType",
			Value:    api.NewOptPetType(api.PetTypeFifa),
			Expected: strconv.Quote(string(api.PetTypeFifa)),
		},
	} {
		// Make range value copy to prevent data races.
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			encode := json.Encode(tc.Value)
			t.Logf("%s", encode)
			require.True(t, jx.Valid(encode), "invalid json")
			if tc.Expected != "" {
				require.JSONEq(t, tc.Expected, string(encode))
			}
		})
	}
}

func TestTechEmpowerJSON(t *testing.T) {
	hw := techempower.WorldObject{
		ID:           10,
		RandomNumber: 2134,
	}
	e := &jx.Encoder{}
	hw.Encode(e)
	var parsed techempower.WorldObject
	d := jx.GetDecoder()
	d.ResetBytes(e.Bytes())
	t.Log(e)
	require.NoError(t, parsed.Decode(d))
	require.Equal(t, hw, parsed)
}
