package ogen

import (
	std "encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
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
	e := &jx.Writer{}
	e.ObjStart()
	if settable, ok := v.(json.Settable); ok && !settable.IsSet() {
		e.ObjEnd()
		return e.Buf
	}
	e.FieldStart("key")
	v.Encode(e)
	e.ObjEnd()
	return e.Buf
}

func testEncode(t *testing.T, encoder json.Marshaler, expected string) {
	e := jx.GetWriter()
	defer jx.PutWriter(e)

	encoder.Encode(e)
	if expected == "" {
		require.Empty(t, e.Buf)
		return
	}
	require.True(t, std.Valid(e.Buf))
	require.JSONEq(t, expected, string(e.Buf), "encoding result mismatch")
}

func TestJSONGenerics(t *testing.T) {
	t.Parallel()

	t.Run("EncodeDecodeEncode", func(t *testing.T) {
		for _, tc := range []struct {
			Name   string
			Value  api.OptNilString
			Result string
		}{
			{
				Name:   "Zero",
				Value:  api.OptNilString{},
				Result: "{}",
			},
			{
				Name:   "Set",
				Value:  api.NewOptNilString("foo"),
				Result: `{"key":"foo"}`,
			},
			{
				Name:   "Nil",
				Value:  api.OptNilString{Null: true, Set: true},
				Result: `{"key":null}`,
			},
		} {
			// Make range value copy to prevent data races.
			tc := tc
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
	})
	t.Run("Encode", func(t *testing.T) {
		for _, tc := range []struct {
			Name     string
			Value    json.Marshaler
			Expected string
		}{
			{
				Name:     "ZeroPrimitive",
				Value:    api.OptNilString{},
				Expected: "",
			},
			{
				Name:     "SetPrimitive",
				Value:    api.NewOptNilString("foo"),
				Expected: `"foo"`,
			},
			{
				Name:     "NilPrimitive",
				Value:    api.OptNilString{Null: true, Set: true},
				Expected: `null`,
			},
			{
				Name:     "ZeroAlias",
				Value:    api.OptPetName{},
				Expected: "",
			},
			{
				Name:     "SetAlias",
				Value:    api.NewOptPetName("foo"),
				Expected: `"foo"`,
			},
			{
				Name:     "ZeroEnum",
				Value:    api.OptPetType{},
				Expected: "",
			},
			{
				Name:     "SetEnum",
				Value:    api.NewOptPetType(api.PetTypeFifa),
				Expected: strconv.Quote(string(api.PetTypeFifa)),
			},
			{
				Name:     "ZeroSum",
				Value:    api.OptID{},
				Expected: "",
			},
			{
				Name:     "SetSum",
				Value:    api.NewOptID(api.NewIntID(10)),
				Expected: `10`,
			},
			{
				Name: "SetArray",
				Value: api.NewOptNilStringArray([]string{
					"aboba",
				}),
				Expected: `["aboba"]`,
			},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()

				testEncode(t, tc.Value, tc.Expected)
			})
		}
	})
}

<<<<<<< HEAD
func TestJSONArray(t *testing.T) {
	t.Run("Decode", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.ArrayTest
			Error    bool
		}{
=======
func TestJSONExample(t *testing.T) {
	t.Parallel()
	date := time.Date(2011, 10, 10, 7, 12, 34, 4125, time.UTC)
	stringMap := api.StringMap{
		"i\nhate": "openapi specification",
	}
	pet := api.Pet{
		Friends:  []api.Pet{},
		Birthday: conv.Date(date),
		ID:       42,
		Name:     "SomePet",
		TestArray1: [][]string{
>>>>>>> cdaeadd (feat(gen): initial additionalProperties support)
			{
				`{"required": [], "nullable_required": []}`,
				api.ArrayTest{},
				false,
			},
			{
				`{"required": [], "optional": [], "nullable_required": [], "nullable_optional": []}`,
				api.ArrayTest{
					NullableOptional: api.OptNilStringArray{
						Set: true,
					},
				},
				false,
			},
			{
				`{"required": [], "nullable_required": null}`,
				api.ArrayTest{},
				false,
			},
			{
				`{"required": [], "nullable_required": null, "nullable_optional": null}`,
				api.ArrayTest{
					NullableOptional: api.OptNilStringArray{
						Set:  true,
						Null: true,
					},
				},
				false,
			},
<<<<<<< HEAD
=======
		},
		TestMap:      api.NewOptStringMap(stringMap),
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
>>>>>>> cdaeadd (feat(gen): initial additionalProperties support)

			// Negative tests
			{
				`{"required": [], "nullable_required": null, "optional": null}`,
				api.ArrayTest{},
				true,
			},
			{
				`{"required": null, "nullable_required": null}`,
				api.ArrayTest{},
				true,
			},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				r := api.ArrayTest{}
				if err := r.Decode(jx.DecodeStr(tc.Input)); tc.Error {
					require.Error(t, err)
				} else {
					require.Equal(t, tc.Expected, r)
					require.NoError(t, err)
				}
			})
		}
<<<<<<< HEAD
	})
	t.Run("Encode", func(t *testing.T) {
		for i, tc := range []struct {
			Value    api.ArrayTest
			Expected string
		}{
			{
				Value:    api.ArrayTest{},
				Expected: `{"required":[],"nullable_required":null}`,
			},
			{
				Value: api.ArrayTest{
					Optional: []string{},
				},
				Expected: `{"required":[],"optional":[],"nullable_required":null}`,
			},
			{
				Value: api.ArrayTest{
					NullableOptional: api.OptNilStringArray{
						Set: true,
					},
				},
				Expected: `{"required":[],"nullable_required":null,"nullable_optional":[]}`,
			},
			{
				Value: api.ArrayTest{
					NullableOptional: api.OptNilStringArray{
						Set:  true,
						Null: true,
					},
				},
				Expected: `{"required":[],"nullable_required":null,"nullable_optional":null}`,
			},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				testEncode(t, tc.Value, tc.Expected)
=======
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
		{
			Name:     "StringMap",
			Value:    stringMap,
			Expected: `{"i\nhate": "openapi specification"}`,
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
	e := &jx.Writer{}
	hw.Encode(e)
	var parsed techempower.WorldObject
	d := jx.GetDecoder()
	d.ResetBytes(e.Buf)
	t.Log(e)
	require.NoError(t, parsed.Decode(d))
	require.Equal(t, hw, parsed)
}

func TestValidateRequired(t *testing.T) {
	data := func() json.Unmarshaler {
		return &api.Data{}
	}
	required := func(fields ...string) (r []validate.FieldError) {
		for _, f := range fields {
			r = append(r, validate.FieldError{
				Name:  f,
				Error: validate.ErrFieldRequired,
>>>>>>> cdaeadd (feat(gen): initial additionalProperties support)
			})
		}
	})
}

func TestJSONNullableEnum(t *testing.T) {
	for _, tc := range []struct {
		Type  json.Unmarshaler
		Error bool
	}{
		{new(api.NullableEnumsOnlyNullable), true},
		{new(api.NilNullableEnumsOnlyNullValue), false},
		{new(api.NilNullableEnumsBoth), false},
	} {
		t.Run(fmt.Sprintf("%T", tc.Type), func(t *testing.T) {
			checker := require.NoError
			if tc.Error {
				checker = require.Error
			}
			checker(t, tc.Type.Decode(jx.DecodeStr(`null`)))
			require.NoError(t, tc.Type.Decode(jx.DecodeStr(`"asc"`)))
		})
	}
}
