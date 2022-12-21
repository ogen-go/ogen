package ogenreflect

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOperationTypes(t *testing.T) {
	typBool := reflect.TypeOf(true)
	typInt64 := reflect.TypeOf(int64(0))
	tests := []struct {
		typ  reflect.Type
		impl []reflect.Type
		v    any
		want bool
	}{
		{nil, nil, int64(0), false},

		{typInt64, nil, int64(0), true},
		{typInt64, nil, int64(1), true},
		{typInt64, nil, time.Second, false},
		{typInt64, nil, "", false},

		{typInt64, []reflect.Type{typInt64}, int64(0), true},
		{typBool, []reflect.Type{typInt64, typBool}, int64(0), true},
		{typInt64, []reflect.Type{typInt64, typBool}, int64(0), true},
		{typBool, []reflect.Type{typInt64, typBool}, false, true},
		{typInt64, []reflect.Type{typInt64, typBool}, false, true},
		{typInt64, []reflect.Type{typInt64, typBool}, "", false},
		{typInt64, []reflect.Type{typInt64, typBool}, time.Second, false},
	}
	t.Run("IsRequest", func(t *testing.T) {
		for i, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				a := require.New(t)

				ot := OperationTypes{
					Request: RequestType{
						Type:            tt.typ,
						Implementations: tt.impl,
					},
				}
				if tt.want {
					a.True(ot.IsRequest(tt.v))
				} else {
					a.False(ot.IsRequest(tt.v))
				}
			})
		}
	})
	t.Run("IsResponse", func(t *testing.T) {
		for i, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				a := require.New(t)

				ot := OperationTypes{
					Responses: ResponsesType{
						Type:            tt.typ,
						Implementations: tt.impl,
					},
				}
				if tt.want {
					a.True(ot.IsResponse(tt.v))
				} else {
					a.False(ot.IsResponse(tt.v))
				}
			})
		}
	})
	t.Run("IsParam", func(t *testing.T) {
		ot := OperationTypes{
			Params: ParametersType{
				Map: ParameterMap[ParameterType]{
					{"cache", "query"}: {
						Type: typBool,
					},
					{"id", "path"}: {
						Type: typInt64,
					},
				},
			},
		}

		a := require.New(t)
		a.True(ot.IsParam(true))
		a.True(ot.IsParam(int64(0)))
		a.False(ot.IsParam(""))
		a.False(ot.IsParam(time.Second))
	})
}

func TestResponseType_FindResponse(t *testing.T) {
	a := ResponseType{
		Contents: Contents{
			"application/json": nil,
		},
	}
	b := ResponseType{
		Contents: Contents{
			"text/html": nil,
		},
	}
	c := ResponseType{}

	tests := []struct {
		patterns map[string]ResponseType
		code     int
		want     ResponseType
		wantOk   bool
	}{
		// Exact match.
		{map[string]ResponseType{"200": a}, 200, a, true},
		{map[string]ResponseType{"200": a, "201": b}, 200, a, true},

		// Pattern.
		{map[string]ResponseType{"2XX": a}, 200, a, true},
		{map[string]ResponseType{"2XX": a}, 201, a, true},

		// Combined.
		{map[string]ResponseType{"200": a, "2XX": b, "default": c}, 200, a, true},
		{map[string]ResponseType{"200": a, "2XX": b, "default": c}, 201, b, true},
		{map[string]ResponseType{"200": a, "2XX": b, "default": c}, 500, c, true},

		// No match.
		{nil, 0, ResponseType{}, false},
		{nil, 200, ResponseType{}, false},
		{map[string]ResponseType{"200": a}, 201, ResponseType{}, false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			rtyp := ResponsesType{
				PatternMap: tt.patterns,
			}
			got, gotOk := rtyp.FindResponse(tt.code)
			if !tt.wantOk {
				a.False(gotOk)
				return
			}
			a.True(gotOk)
			a.Equal(tt.want, got)
		})
	}
}
