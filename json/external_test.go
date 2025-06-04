package json_test

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/_testdata/testtypes"
	"github.com/ogen-go/ogen/json"
)

func TestEncodeNative(t *testing.T) {
	var e jx.Encoder
	json.EncodeNative(&e, testtypes.NumberOgen{Value: 123})
	require.Equal(t, []byte("123"), e.Bytes())
}

func TestDecodeNative(t *testing.T) {
	d := jx.DecodeBytes([]byte("123"))
	val, err := json.DecodeNative[testtypes.NumberOgen](d)
	require.NoError(t, err)
	require.Equal(t, testtypes.NumberOgen{Value: 123}, val)
}

func TestEncodeStringNative(t *testing.T) {
	var e jx.Encoder
	json.EncodeStringNative(&e, testtypes.StringOgen{Value: "test"})
	require.Equal(t, []byte(`"test"`), e.Bytes())
}

func TestDecodeStringNative(t *testing.T) {
	d := jx.DecodeBytes([]byte(`"test"`))
	val, err := json.DecodeStringNative[testtypes.StringOgen](d)
	require.NoError(t, err)
	require.Equal(t, testtypes.StringOgen{Value: "test"}, val)
}

func TestEncodeText(t *testing.T) {
	var e jx.Encoder
	json.EncodeText(&e, testtypes.Text{Value: "123"})
	require.Equal(t, []byte("123"), e.Bytes())
}

func TestDecodeText(t *testing.T) {
	d := jx.DecodeBytes([]byte("123"))
	val, err := json.DecodeText[testtypes.Text](d)
	require.NoError(t, err)
	require.Equal(t, testtypes.Text{Value: "123"}, val)
}

func TestEncodeStringText(t *testing.T) {
	var e jx.Encoder
	json.EncodeStringText(&e, testtypes.Text{Value: "test"})
	require.Equal(t, []byte(`"test"`), e.Bytes())
}

func TestDecodeStringText(t *testing.T) {
	d := jx.DecodeBytes([]byte(`"test"`))
	val, err := json.DecodeStringText[testtypes.Text](d)
	require.NoError(t, err)
	require.Equal(t, testtypes.Text{Value: "test"}, val)
}

func TestEncodeJSON(t *testing.T) {
	var e jx.Encoder
	json.EncodeJSON(&e, testtypes.NumberJSON{Value: 123})
	require.JSONEq(t, "123", string(e.Bytes()))
}

func TestDecodeJSON(t *testing.T) {
	d := jx.DecodeBytes([]byte("123"))
	val, err := json.DecodeJSON[testtypes.NumberJSON](d)
	require.NoError(t, err)
	require.Equal(t, float64(123), val.Value)
}

func TestEncodeStringJSON(t *testing.T) {
	var e jx.Encoder
	json.EncodeStringJSON(&e, testtypes.NumberJSON{Value: 123})
	require.JSONEq(t, "123", string(e.Bytes()))
}

func TestDecodeStringJSON(t *testing.T) {
	d := jx.DecodeBytes([]byte("123"))
	val, err := json.DecodeStringJSON[testtypes.NumberJSON](d)
	require.NoError(t, err)
	require.Equal(t, float64(123), val.Value)
}

func TestEncodeExternal(t *testing.T) {
	var e jx.Encoder
	json.EncodeExternal(&e, testtypes.Number(123))
	require.Equal(t, []byte("123"), e.Bytes())
}

func TestDecodeExternal(t *testing.T) {
	d := jx.DecodeBytes([]byte("123"))
	val, err := json.DecodeExternal[testtypes.Number](d)
	require.NoError(t, err)
	require.Equal(t, testtypes.Number(123), val)
}

func TestEncodeStringExternal(t *testing.T) {
	var e jx.Encoder
	json.EncodeStringExternal(&e, testtypes.String("test"))
	require.Equal(t, []byte(`"test"`), e.Bytes())
}

func TestDecodeStringExternal(t *testing.T) {
	d := jx.DecodeBytes([]byte(`"test"`))
	val, err := json.DecodeStringExternal[testtypes.String](d)
	require.NoError(t, err)
	require.Equal(t, testtypes.String("test"), val)
}
