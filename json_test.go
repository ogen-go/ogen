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

func TestJSONArray(t *testing.T) {
	t.Run("Decode", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.ArrayTest
			Error    bool
		}{
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
			})
		}
	})
}

func TestJSONAdditionalProperties(t *testing.T) {
	t.Run("Decode", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.MapWithProperties
			Error    bool
		}{
			{
				`{"required": 1}`,
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
				},
				false,
			},
			{
				`{"required": 1, "optional": 10}`,
				api.MapWithProperties{
					Required:        1,
					Optional:        api.NewOptInt(10),
					AdditionalProps: map[string]string{},
				},
				false,
			},
			{
				`{"required": 1, "runtime_field": "field"}`,
				api.MapWithProperties{
					Required: 1,
					AdditionalProps: map[string]string{
						"runtime_field": "field",
					},
				},
				false,
			},
			{
				`{"required": 1, "sub_map":{"runtime_field": "field"}}`,
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
					SubMap: api.NewOptStringMap(api.StringMap{
						"runtime_field": "field",
					}),
				},
				false,
			},
			{
				`{"required": 1, "inlined_sub_map":{"runtime_field": "field"}}`,
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
					InlinedSubMap: api.NewOptMapWithPropertiesInlinedSubMap(api.MapWithPropertiesInlinedSubMap{
						"runtime_field": "field",
					}),
				},
				false,
			},
			{
				// MapWithProperties expects string for `runtime_field`.
				`{"required": 1, "runtime_field": 10}`,
				api.MapWithProperties{},
				true,
			},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				r := api.MapWithProperties{}
				if err := r.Decode(jx.DecodeStr(tc.Input)); tc.Error {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, tc.Expected, r)
				}
			})
		}
	})
	t.Run("Encode", func(t *testing.T) {
		for i, tc := range []struct {
			Value    json.Marshaler
			Expected string
		}{
			{
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
				},
				`{"required": 1}`,
			},
			{
				api.MapWithProperties{
					Required:        1,
					Optional:        api.NewOptInt(10),
					AdditionalProps: map[string]string{},
				},
				`{"required": 1, "optional": 10}`,
			},
			{
				api.MapWithProperties{
					Required: 1,
					AdditionalProps: map[string]string{
						"runtime_field": "field",
					},
				},
				`{"required": 1, "runtime_field": "field"}`,
			},
			{
				api.MapWithProperties{},
				`{"required": 0}`,
			},
			{
				api.StringStringMap{
					"a": api.StringMap{
						"b": "c",
					},
				},
				`{"a":{"b":"c"}}`,
			},
			{
				api.MapWithProperties{
					Required: 1,
					SubMap: api.NewOptStringMap(api.StringMap{
						"runtime_field": "field",
					}),
				},
				`{"required": 1, "sub_map":{"runtime_field": "field"}}`,
			},
			{
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
					InlinedSubMap: api.NewOptMapWithPropertiesInlinedSubMap(api.MapWithPropertiesInlinedSubMap{
						"runtime_field": "field",
					}),
				},
				`{"required": 1, "inlined_sub_map":{"runtime_field": "field"}}`,
			},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				testEncode(t, tc.Value, tc.Expected)
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
