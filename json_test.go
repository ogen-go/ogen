package ogen

import (
	"bytes"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/jx"

	"github.com/ogen-go/ogen/conv"
	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/internal/techempower"
	"github.com/ogen-go/ogen/json"
)

func decodeObject(t testing.TB, data []byte, v json.Unmarshaler) {
	r := json.GetReader()
	r.ResetBytes(data)
	defer json.PutReader(r)
	if rs, ok := v.(json.Resettable); ok {
		rs.Reset()
	}
	require.NoError(t, r.ObjBytes(func(r *json.Reader, _ []byte) error {
		return v.ReadJSON(r)
	}))
}

func encodeObject(v json.Marshaler) []byte {
	buf := new(bytes.Buffer)
	s := json.GetWriter()
	s.Reset(buf)
	s.ObjStart()
	if settable, ok := v.(json.Settable); ok && !settable.IsSet() {
		s.ObjEnd()
		_ = s.Flush()
		return buf.Bytes()
	}
	s.ObjField("key")
	v.WriteJSON(s)
	s.ObjEnd()
	_ = s.Flush()
	return buf.Bytes()
}

func TestJSONGenerics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name   string
		Value  api.OptNilString
		Result string
	}{
		{
			Name:   "Zero",
			Result: "{}",
		},
		{
			Name:   "Set",
			Result: `{"key":"foo"}`,
			Value:  api.NewOptNilString("foo"),
		},
		{
			Name:   "Nil",
			Result: `{"key":null}`,
			Value:  api.OptNilString{Null: true, Set: true},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			result := encodeObject(tc.Value)
			require.Equal(t, tc.Result, string(result), "encoding result mismatch")
			var v api.OptNilString
			decodeObject(t, result, &v)
			require.Equal(t, tc.Value, v)
			require.Equal(t, tc.Result, string(encodeObject(v)))
		})
	}
}

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
			Description: api.NewOptString("Foo"),
			ID:          api.NewIntID(10),
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
	buf := new(bytes.Buffer)
	s := json.GetWriter()
	s.Reset(buf)
	hw.WriteJSON(s)
	require.NoError(t, s.Flush())
	var parsed techempower.WorldObject
	i := json.GetReader()
	i.ResetBytes(buf.Bytes())
	t.Log(buf.String())
	require.NoError(t, parsed.ReadJSON(i))
	require.Equal(t, hw, parsed)
}
