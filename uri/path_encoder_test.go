package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathEncoder(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  string
			Style   PathStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "foo",
				Expect:  "foo",
				Style:   PathStyleSimple,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "f%o",
				Expect:  "f%25o",
				Style:   PathStyleSimple,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  "foo",
				Style:   PathStyleSimple,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  "foo",
				Style:   PathStyleSimple,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  ".foo",
				Style:   PathStyleLabel,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "f%o",
				Expect:  ".f%25o",
				Style:   PathStyleLabel,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  ".foo",
				Style:   PathStyleLabel,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  ";id=foo",
				Style:   PathStyleMatrix,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "f%o",
				Expect:  ";id=f%25o",
				Style:   PathStyleMatrix,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  ";id=foo",
				Style:   PathStyleMatrix,
				Explode: true,
			},
		}

		for i, test := range tests {
			e := NewPathEncoder(PathEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			})
			require.NoError(t, e.EncodeValue(test.Input))
			require.Equal(t, test.Expect, e.Result(), fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Array", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []string
			Expect  string
			Style   PathStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   []string{"foo", "bar"},
				Expect:  "foo,bar",
				Style:   PathStyleSimple,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"f%o", "b?r"},
				Expect:  "f%25o,b%3Fr",
				Style:   PathStyleSimple,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"foo", "bar"},
				Expect:  "foo,bar",
				Style:   PathStyleSimple,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"foo", "bar"},
				Expect:  ".foo,bar",
				Style:   PathStyleLabel,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"f%o", "b?r"},
				Expect:  ".f%25o,b%3Fr",
				Style:   PathStyleLabel,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"foo", "bar"},
				Expect:  ".foo.bar",
				Style:   PathStyleLabel,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"foo", "bar"},
				Expect:  ";id=foo,bar",
				Style:   PathStyleMatrix,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"f%o", "b?r"},
				Expect:  ";id=f%25o,b%3Fr",
				Style:   PathStyleMatrix,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"foo", "bar"},
				Expect:  ";id=foo;id=bar",
				Style:   PathStyleMatrix,
				Explode: true,
			},
		}

		for i, test := range tests {
			e := NewPathEncoder(PathEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			})
			err := e.EncodeArray(func(e Encoder) error {
				for _, item := range test.Input {
					if err := e.EncodeValue(item); err != nil {
						return err
					}
				}
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, test.Expect, e.Result(), fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Object", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []Field
			Expect  string
			Style   PathStyle
			Explode bool
		}{
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   PathStyleSimple,
				Explode: false,
				Expect:  "role,admin,firstName,Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"r?le", "%dmin"},
					{"f?rstName", "Alex"},
				},
				Style:   PathStyleSimple,
				Explode: false,
				Expect:  "r%3Fle,%25dmin,f%3FrstName,Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   PathStyleSimple,
				Explode: true,
				Expect:  "role=admin,firstName=Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   PathStyleLabel,
				Explode: false,
				Expect:  ".role,admin,firstName,Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"r?le", "%dmin"},
					{"f?rstName", "Alex"},
				},
				Style:   PathStyleLabel,
				Explode: false,
				Expect:  ".r%3Fle,%25dmin,f%3FrstName,Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   PathStyleLabel,
				Explode: true,
				Expect:  ".role=admin.firstName=Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   PathStyleMatrix,
				Explode: false,
				Expect:  ";id=role,admin,firstName,Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"r?le", "%dmin"},
					{"f?rstName", "Alex"},
				},
				Style:   PathStyleMatrix,
				Explode: false,
				Expect:  ";id=r%3Fle,%25dmin,f%3FrstName,Alex",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   PathStyleMatrix,
				Explode: true,
				Expect:  ";role=admin;firstName=Alex",
			},
		}

		for i, test := range tests {
			e := NewPathEncoder(PathEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			})
			for _, field := range test.Input {
				err := e.EncodeField(field.Name, func(e Encoder) error {
					return e.EncodeValue(field.Value)
				})
				require.NoError(t, err)
			}
			require.Equal(t, test.Expect, e.Result(), fmt.Sprintf("Test %d", i+1))
		}
	})
}

func BenchmarkPathEncoder(b *testing.B) {
	b.Run("Array", func(b *testing.B) {
		var (
			cfg = PathEncoderConfig{
				Param:   "tags",
				Style:   PathStyleSimple,
				Explode: true,
			}
			elems = []string{
				"S&P500",
				"100%",
				"1_000_000",
				"among us",
				"foo",
				"bar",
			}

			sink string
		)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			e := NewPathEncoder(cfg)
			if err := e.EncodeArray(func(e Encoder) error {
				for _, elem := range elems {
					if err := e.EncodeValue(elem); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				b.Fatal(err)
			}
			sink = e.Result()
		}
		if sink == "" {
			b.Fatal("sink is empty")
		}
	})
	b.Run("Object", func(b *testing.B) {
		var (
			cfg = PathEncoderConfig{
				Param:   "user",
				Style:   PathStyleSimple,
				Explode: true,
			}
			fields = []Field{
				{"username", "%dmin"},
				{"name", "Dark&Brandon"},
			}
			sink string
		)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			e := NewPathEncoder(cfg)
			for _, field := range fields {
				if err := e.EncodeField(field.Name, func(e Encoder) error {
					return e.EncodeValue(field.Value)
				}); err != nil {
					b.Fatal(err)
				}
			}
			sink = e.Result()
		}
		if sink == "" {
			b.Fatal("sink is empty")
		}
	})
}
