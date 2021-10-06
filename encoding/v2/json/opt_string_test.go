package json

import (
	"bytes"
	"testing"

	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Helper struct {
	w *bytes.Buffer
	s *json.Stream
	i *json.Iterator
}

func writeSimpleObject(s *json.Stream, v Marshaler) error {
	s.WriteObjectStart()
	v.WriteFieldJSON("key", s)
	s.WriteObjectEnd()
	return s.Error
}

func readSimpleObject(i *json.Iterator, v Unmarshaler) error {
	i.ReadObjectCB(func(i *json.Iterator, s string) bool {
		return v.ReadJSON(i)
	})
	return i.Error
}

func (h *Helper) Write(t testing.TB, v Marshaler) {
	t.Helper()
	require.NoError(t, writeSimpleObject(h.s, v))
	require.NoError(t, h.s.Flush())
}

func (h *Helper) Reset() {
	h.w.Reset()
	h.s.Reset(h.w)
}

func (h *Helper) Read(t testing.TB, v Unmarshaler) {
	t.Helper()
	h.i.ResetBytes(h.w.Bytes())
	h.i.ReadObjectCB(func(i *json.Iterator, s string) bool {
		assert.Equal(t, "key", s)
		assert.NoError(t, i.Error)
		return v.ReadJSON(i)
	})
	require.NoError(t, h.i.Error)
}

func (h *Helper) Check(t testing.TB, v Value) {
	t.Helper()
	h.Field(t, v, v)
}

func (h *Helper) Field(t testing.TB, in Marshaler, out Unmarshaler) {
	t.Helper()
	h.Reset()
	h.Write(t, in)
	h.Read(t, out)
	require.Equal(t, in, out)
}

func New() *Helper {
	buf := new(bytes.Buffer)
	s := json.NewStream(json.ConfigDefault, buf, 1024)
	i := json.NewIterator(json.ConfigDefault)
	return &Helper{
		w: buf,
		s: s,
		i: i,
	}
}

func TestOptionalNullableString_ReadJSON(t *testing.T) {
	for _, tc := range []struct {
		Name string
		In   OptionalNullableString
	}{
		{
			Name: "NotSet",
		},
		{
			Name: "SetNil",
			In: OptionalNullableString{
				Set: true,
				NullableString: NullableString{
					Nil: true,
				},
			},
		},
		{
			Name: "SetValue",
			In: OptionalNullableString{
				Set: true,
				NullableString: NullableString{
					Value: "Value",
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			var out OptionalNullableString
			New().Field(t, &tc.In, &out)
			require.Equal(t, tc.In, out)
		})
	}
}

func BenchmarkOptionalNullableString(b *testing.B) {
	v := &OptionalNullableString{
		Set: true,
		NullableString: NullableString{
			Value: "Value",
		},
	}
	buf := new(bytes.Buffer)
	s := json.NewStream(json.ConfigFastest, buf, 1024)
	require.NoError(b, writeSimpleObject(s, v))
	require.NoError(b, s.Flush())

	b.Run("Write", func(b *testing.B) {
		b.Run("Single", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(buf.Len()))

			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := writeSimpleObject(s, v); err != nil {
					b.Fatal(err)
				}
				if err := s.Flush(); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("Multi", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(buf.Len()))

			b.RunParallel(func(pb *testing.PB) {
				buf := new(bytes.Buffer)
				s := json.NewStream(json.ConfigFastest, buf, 1024)

				for pb.Next() {
					buf.Reset()
					if err := writeSimpleObject(s, v); err != nil {
						b.Fatal(err)
					}
					if err := s.Flush(); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	})
	b.Run("Read", func(b *testing.B) {
		b.Run("Single", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(buf.Len()))
			iter := json.NewIterator(json.ConfigFastest)
			var out OptionalNullableString

			for i := 0; i < b.N; i++ {
				iter.ResetBytes(buf.Bytes())
				if err := readSimpleObject(iter, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("Multi", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(buf.Len()))

			b.RunParallel(func(pb *testing.PB) {
				iter := json.NewIterator(json.ConfigFastest)
				var out OptionalNullableString

				for pb.Next() {
					iter.ResetBytes(buf.Bytes())
					if err := readSimpleObject(iter, &out); err != nil {
						b.Error(err)
					}
				}
			})
		})
	})
}
