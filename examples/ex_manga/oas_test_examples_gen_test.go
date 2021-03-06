// Code generated by ogen, DO NOT EDIT.

package api

import (
	"github.com/go-faster/jx"

	std "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBook_EncodeDecode(t *testing.T) {
	var typ Book
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Book
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestImage_EncodeDecode(t *testing.T) {
	var typ Image
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Image
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestImages_EncodeDecode(t *testing.T) {
	var typ Images
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Images
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestSearchByTagIDOKApplicationJSON_EncodeDecode(t *testing.T) {
	var typ SearchByTagIDOKApplicationJSON
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 SearchByTagIDOKApplicationJSON
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestSearchOKApplicationJSON_EncodeDecode(t *testing.T) {
	var typ SearchOKApplicationJSON
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 SearchOKApplicationJSON
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestSearchResponse_EncodeDecode(t *testing.T) {
	var typ SearchResponse
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 SearchResponse
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestTag_EncodeDecode(t *testing.T) {
	var typ Tag
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Tag
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestTagType_EncodeDecode(t *testing.T) {
	var typ TagType
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 TagType
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestTitle_EncodeDecode(t *testing.T) {
	var typ Title
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Title
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
