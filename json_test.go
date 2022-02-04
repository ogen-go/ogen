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
	require.True(t, std.Valid(e.Buf), string(e.Buf))
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
				// ValidationStringMap expects maximum 4 property.
				`{"required": 1, "map_validation": {"1": "1", "2": "2"}}`,
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
					MapValidation: api.NewOptValidationStringMap(api.ValidationStringMap{
						"1": "1",
						"2": "2",
					}),
				},
				false,
			},
			{
				// ValidationStringMap expects maximum 4 property.
				`{"required": 1, "map_validation": {"1": "1", "2": "2", "3": "3", "4": "4"}}`,
				api.MapWithProperties{
					Required:        1,
					AdditionalProps: map[string]string{},
					MapValidation: api.NewOptValidationStringMap(api.ValidationStringMap{
						"1": "1",
						"2": "2",
						"3": "3",
						"4": "4",
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
			{
				// MapWithProperties expects maximum 7 properties.
				`{"required": 1, "2": "2", "3": "3", "4": "4", "5": "5", "6": "6", "7": "7", "8":"8"}`,
				api.MapWithProperties{},
				true,
			},
			{
				// MapWithProperties expects maximum 7 properties.
				`{
   "required":1,
   "inlined_sub_map":{
      "runtime_field":"field"
   },
   "sub_map":{
      "runtime_field":"field"
   },
   "4":"4",
   "5":"5",
   "6":"6",
   "7":"7",
   "8":"8"
}`,
				api.MapWithProperties{},
				true,
			},
			{
				// ValidationStringMap expects minimum 1 property.
				`{"required": 1, "map_validation": {}}`,
				api.MapWithProperties{},
				true,
			},
			{
				// ValidationStringMap expects maximum 4 property.
				`{"required": 1, "map_validation": {"1": "1", "2": "2", "3": "3", "4": "4", "5": "5"}}`,
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

func TestJSONPropertiesCount(t *testing.T) {
	for i, tc := range []struct {
		Input string
		Error bool
	}{
		{
			`{"required": 1, "optional_a": 1}`,
			false,
		},
		{
			`{"required": 1, "optional_b": 1}`,
			false,
		},
		{
			`{}`,
			true,
		},
		{
			`{"required": 1}`,
			true,
		},
		{
			`{"optional_a": 1, "optional_b": 1}`,
			true,
		},
		{
			`{"required": 1, "optional_a": 1, "optional_b": 1}`,
			true,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			m := api.MaxPropertiesTest{}
			checker := require.NoError
			if tc.Error {
				checker = require.Error
			}
			checker(t, m.Decode(jx.DecodeStr(tc.Input)))
		})
	}
}

func TestJSONSum(t *testing.T) {
	t.Run("Issue143", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.Issue143Type
			Error    bool
		}{
			{`{"common-1": "abc", "common-2": 1, "unique-1": "unique"}`, api.Issue1430Issue143, false},
			{`{"common-1": "abc", "common-2": 1, "unique-2": "unique"}`, api.Issue1431Issue143, false},
			{`{"common-1": "abc", "common-2": 1, "unique-3": "unique"}`, api.Issue1432Issue143, false},
			{`{"common-1": "abc", "common-2": 1, "unique-4": "unique"}`, api.Issue1433Issue143, false},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.Issue143{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				require.Equal(t, tc.Expected, r.Type)
			})
		}
	})
	t.Run("OneVariantHasNoUniqueFields", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.OneVariantHasNoUniqueFieldsType
			Error    bool
		}{
			{`{"a": "a", "c": "c"}`,
				api.OneVariantHasNoUniqueFields0OneVariantHasNoUniqueFields, false},
			{`{"a": "a", "b": 10, "c": "c"}`,
				api.OneVariantHasNoUniqueFields0OneVariantHasNoUniqueFields, false},
			{`{"a": "a", "c": "c", "d": 10}`,
				api.OneVariantHasNoUniqueFields1OneVariantHasNoUniqueFields, false},
			{`{"a": "a", "b": 10, "c": "c", "d": 10}`,
				api.OneVariantHasNoUniqueFields1OneVariantHasNoUniqueFields, false},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.OneVariantHasNoUniqueFields{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				require.Equal(t, tc.Expected, r.Type)
			})
		}
	})
	t.Run("AnyOfTestSizeLimit", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.AnyOfTestSizeLimitType
			Error    bool
		}{
			{`10`, api.IntAnyOfTestSizeLimit, false},
			{`"10"`, api.StringAnyOfTestSizeLimit, false},
			{`true`, "", true},
			{`null`, "", true},
		} {
			// Make range value copy to prevent data races.
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.AnyOfTestSizeLimit{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				require.Equal(t, tc.Expected, r.Type)
			})
		}
	})
}

func TestJSONAny(t *testing.T) {
	validJSON := []string{
		`null`,
		`true`,
		`false`,
		`10`,
		`10.0`,
		`10.0e1`,
		`{}`,
		`{"a":"b"}`,
		`[{"a":"b"}]`,
		`[{"a":{}}]`,
	}
	templateCase := func(f string) (r []string) {
		for _, val := range validJSON {
			r = append(r, fmt.Sprintf(f, val))
		}
		return r
	}
	type testCases struct {
		Name   string
		Inputs []string
		Error  bool
	}
	var cases []testCases

	for _, template := range []struct {
		Name   string
		Format string
		Error  bool
	}{
		{
			Name:   "Raw",
			Format: `{"empty":%s}`,
			Error:  false,
		},

		{Name: "AnyArray",
			Format: `{"any_array":[%s]}`,
			Error:  false,
		},
		{
			Name:   "AnyMap",
			Format: `{"any_map":{"key": %s}}`,
			Error:  false,
		},
	} {
		cases = append(cases, testCases{
			Name:   template.Name,
			Inputs: templateCase(template.Format),
			Error:  template.Error,
		})
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			for i, input := range tc.Inputs {
				input := input
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					typ := &api.AnyTest{}
					checker := require.NoError
					if tc.Error {
						checker = require.Error
					}
					checker(t, typ.Decode(jx.DecodeStr(input)))
					if !tc.Error {
						testEncode(t, typ, input)
					}
				})
			}
		})
	}
}
