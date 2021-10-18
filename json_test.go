package ogen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/json"
)

func decodeObject(t testing.TB, data []byte, v json.Unmarshaler) {
	i := json.NewIterator()
	i.ResetBytes(data)
	if rs, ok := v.(json.Resettable); ok {
		rs.Reset()
	}
	i.ReadMapCB(func(iterator *json.Iterator, s string) bool {
		require.NoError(t, v.ReadJSON(i))
		return true
	})
	require.NoError(t, i.Error)
}

func encodeObject(v json.Marshaler) []byte {
	buf := new(bytes.Buffer)
	s := json.NewStream(buf)
	s.WriteObjectStart()
	if settable, ok := v.(json.Settable); ok && !settable.IsSet() {
		s.WriteObjectEnd()
		_ = s.Flush()
		return buf.Bytes()
	}
	s.WriteObjectField("key")
	v.WriteJSON(s)
	s.WriteObjectEnd()
	_ = s.Flush()
	return buf.Bytes()
}

func TestJSONGenerics(t *testing.T) {
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
			result := encodeObject(tc.Value)
			require.Equal(t, tc.Result, string(result), "encoding result mismatch")
			var v api.OptNilString
			decodeObject(t, result, &v)
			require.Equal(t, tc.Value, v)
			require.Equal(t, tc.Result, string(encodeObject(v)))
		})
	}
}
