// Code generated by ogen, DO NOT EDIT.

package api

import (
	"github.com/go-faster/jx"

	std "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoard_EncodeDecode(t *testing.T) {
	var typ Board
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Board
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestBoardIconsItem_EncodeDecode(t *testing.T) {
	var typ BoardIconsItem
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 BoardIconsItem
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestBoards_EncodeDecode(t *testing.T) {
	var typ Boards
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Boards
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestCaptcha_EncodeDecode(t *testing.T) {
	var typ Captcha
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Captcha
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestCaptchaType_EncodeDecode(t *testing.T) {
	var typ CaptchaType
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 CaptchaType
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestError_EncodeDecode(t *testing.T) {
	var typ Error
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Error
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestErrorCode_EncodeDecode(t *testing.T) {
	var typ ErrorCode
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 ErrorCode
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestFile_EncodeDecode(t *testing.T) {
	var typ File
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 File
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestFileType_EncodeDecode(t *testing.T) {
	var typ FileType
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 FileType
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestLike_EncodeDecode(t *testing.T) {
	var typ Like
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Like
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestMobilePost_EncodeDecode(t *testing.T) {
	var typ MobilePost
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 MobilePost
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestMobileThreadLastInfo_EncodeDecode(t *testing.T) {
	var typ MobileThreadLastInfo
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 MobileThreadLastInfo
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestMobileThreadLastInfoThread_EncodeDecode(t *testing.T) {
	var typ MobileThreadLastInfoThread
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 MobileThreadLastInfoThread
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestMobileThreadPostsAfter_EncodeDecode(t *testing.T) {
	var typ MobileThreadPostsAfter
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 MobileThreadPostsAfter
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestPasscode_EncodeDecode(t *testing.T) {
	var typ Passcode
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Passcode
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestPasscodePasscode_EncodeDecode(t *testing.T) {
	var typ PasscodePasscode
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 PasscodePasscode
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestPost_EncodeDecode(t *testing.T) {
	var typ Post
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Post
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestPostingNewPost_EncodeDecode(t *testing.T) {
	var typ PostingNewPost
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 PostingNewPost
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestPostingNewThread_EncodeDecode(t *testing.T) {
	var typ PostingNewThread
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 PostingNewThread
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestReport_EncodeDecode(t *testing.T) {
	var typ Report
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 Report
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
func TestUserPostingPostOK_EncodeDecode(t *testing.T) {
	var typ UserPostingPostOK
	typ.SetFake()

	e := jx.Encoder{}
	typ.Encode(&e)
	data := e.Bytes()
	require.True(t, std.Valid(data), "Encoded: %s", data)

	var typ2 UserPostingPostOK
	require.NoError(t, typ2.Decode(jx.DecodeBytes(data)))
}
