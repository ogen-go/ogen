package ogen

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/conv"
	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/internal/techempower"
	"github.com/ogen-go/ogen/json"
)

func decodeObject(t testing.TB, data []byte, v json.Unmarshaler) {
	d := jx.GetDecoder()
	d.ResetBytes(data)
	defer jx.PutDecoder(d)
	if rs, ok := v.(json.Resettable); ok {
		rs.Reset()
	}
	require.NoError(t, d.ObjBytes(func(d *jx.Decoder, _ []byte) error {
		return v.Decode(d)
	}))
}

func encodeObject(v json.Marshaler) []byte {
	e := jx.GetEncoder()
	e.ObjStart()
	if settable, ok := v.(json.Settable); ok && !settable.IsSet() {
		e.ObjEnd()
		return e.Bytes()
	}
	e.FieldStart("key")
	v.Encode(e)
	e.ObjEnd()
	return e.Bytes()
}

// func TestJSONGenerics(t *testing.T) {
// 	t.Parallel()

// 	for _, tc := range []struct {
// 		Name   string
// 		Value  api.OptNilString
// 		Result string
// 	}{
// 		{
// 			Name:   "Zero",
// 			Result: "{}",
// 		},
// 		{
// 			Name:   "Set",
// 			Result: `{"key":"foo"}`,
// 			Value:  api.NewOptNilString("foo"),
// 		},
// 		{
// 			Name:   "Nil",
// 			Result: `{"key":null}`,
// 			Value:  api.OptNilString{Null: true, Set: true},
// 		},
// 	} {
// 		t.Run(tc.Name, func(t *testing.T) {
// 			t.Parallel()

// 			result := encodeObject(tc.Value)
// 			require.Equal(t, tc.Result, string(result), "encoding result mismatch")
// 			var v api.OptNilString
// 			decodeObject(t, result, &v)
// 			require.Equal(t, tc.Value, v)
// 			require.Equal(t, tc.Result, string(encodeObject(v)))
// 		})
// 	}
// }

func TestJSONExample(t *testing.T) {
	t.Parallel()

	date := time.Date(2011, 10, 10, 7, 12, 34, 4125, time.UTC)
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
		Nickname:     api.NewNilString("Nick"),
		NullStr:      api.NewOptNilString("Bar"),
		Rate:         time.Second,
		Tag:          api.NewOptUUID(uuid.New()),
		TestDate:     api.NewOptTime(conv.Date(date)),
		TestDateTime: api.NewOptTime(conv.DateTime(date)),
		TestDuration: api.NewOptDuration(time.Minute),
		TestFloat1:   api.NewOptFloat64(1.0),
		TestInteger1: api.NewOptInt(10),
		TestTime:     api.NewOptTime(conv.Time(date)),
		UniqueID:     uuid.New(),
		URI:          url.URL{Scheme: "s3", Host: "foo", Path: "bar"},
		IP:           net.IPv4(127, 0, 0, 1),
		IPV4:         net.IPv4(127, 0, 0, 1),
		IPV6:         net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		Next: api.NewOptData(api.Data{
			Description: api.NewDescriptionSimpleDataDescription(api.DescriptionSimple{
				Description: "foo",
			}),
			ID: api.NewIntID(10),
		}),
	}
	t.Logf("%s", json.Encode(pet))
	require.True(t, jx.Valid(json.Encode(pet)), "invalid json")
}

func TestTechEmpowerJSON(t *testing.T) {
	hw := techempower.WorldObject{
		ID:           10,
		RandomNumber: 2134,
	}
	e := jx.GetEncoder()
	hw.Encode(e)
	var parsed techempower.WorldObject
	d := jx.GetDecoder()
	d.ResetBytes(e.Bytes())
	t.Log(e)
	require.NoError(t, parsed.Decode(d))
	require.Equal(t, hw, parsed)
}
