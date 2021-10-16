package ogen

import (
	"bytes"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/encoding/v2/json"
	api "github.com/ogen-go/ogen/internal/sample_api"
)

func decodeObject(t testing.TB, data []byte, v json.Unmarshaler) {
	i := jsoniter.NewIterator(jsoniter.ConfigDefault)
	i.ResetBytes(data)
	if rs, ok := v.(json.Resettable); ok {
		rs.Reset()
	}
	i.ReadMapCB(func(iterator *jsoniter.Iterator, s string) bool {
		require.NoError(t, v.ReadJSON(i))
		return true
	})
	require.NoError(t, i.Error)
}

func encodeObject(v json.Marshaler) []byte {
	buf := new(bytes.Buffer)
	s := jsoniter.NewStream(jsoniter.ConfigDefault, buf, 1024)
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
		Value  api.OptionalNilInt64
		Result string
	}{
		{
			Name:   "Zero",
			Result: "{}",
		},
		{
			Name:   "Set",
			Result: `{"key":10}`,
			Value:  api.NewOptionalNilInt64(10),
		},
		{
			Name:   "Nil",
			Result: `{"key":null}`,
			Value:  api.OptionalNilInt64{Nil: true, Set: true},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			result := encodeObject(tc.Value)
			require.Equal(t, tc.Result, string(result), "encoding result mismatch")
			var v api.OptionalNilInt64
			decodeObject(t, result, &v)
			require.Equal(t, tc.Value, v)
			require.Equal(t, tc.Result, string(encodeObject(v)))
		})
	}
}
