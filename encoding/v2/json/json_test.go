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

func (h *Helper) Write(t testing.TB, v Marshaler) {
	t.Helper()
	h.s.WriteObjectStart()
	require.NoError(t, v.WriteFieldJSON("key", h.s))
	h.s.WriteObjectEnd()

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
		return assert.NoError(t, v.ReadJSON(i))
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
