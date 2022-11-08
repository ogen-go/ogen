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
					Response: ResponseType{
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
			Params: ParameterMap[ParameterType]{
				{"cache", "query"}: {
					Type: typBool,
				},
				{"id", "path"}: {
					Type: typInt64,
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

func TestResponseType_FindContents(t *testing.T) {
	a := Contents{"application/json": nil}
	b := Contents{"text/html": nil}
	c := Contents{}

	tests := []struct {
		patterns map[string]Contents
		code     int
		want     Contents
		wantOk   bool
	}{
		// Exact match.
		{map[string]Contents{"200": a}, 200, a, true},
		{map[string]Contents{"200": a, "201": b}, 200, a, true},

		// Pattern.
		{map[string]Contents{"2XX": a}, 200, a, true},
		{map[string]Contents{"2XX": a}, 201, a, true},

		// Combined.
		{map[string]Contents{"200": a, "2XX": b, "default": c}, 200, a, true},
		{map[string]Contents{"200": a, "2XX": b, "default": c}, 201, b, true},
		{map[string]Contents{"200": a, "2XX": b, "default": c}, 500, c, true},

		// No match.
		{nil, 0, nil, false},
		{map[string]Contents{"200": a}, 201, nil, false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			rtyp := ResponseType{
				PatternMap: tt.patterns,
			}
			got, gotOk := rtyp.FindContents(tt.code)
			if !tt.wantOk {
				a.False(gotOk)
				return
			}
			a.True(gotOk)
			a.Equal(tt.want, got)
		})
	}
}
