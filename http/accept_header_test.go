package http_test

import (
	"errors"
	"testing"

	"github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/uri"
	"github.com/stretchr/testify/require"
)

func TestAcceptHeaderNew(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		require.Equal(t, http.AcceptHeader(nil), actual)
	})
	t.Run("one", func(t *testing.T) {
		actual := http.AcceptHeaderNew("application/json")
		require.Equal(t, http.AcceptHeader([]string{"application/json"}), actual)
	})
	t.Run("two", func(t *testing.T) {
		actual := http.AcceptHeaderNew("application/json", "text/*")
		require.Equal(t, http.AcceptHeader([]string{"application/json", "text/*"}), actual)
	})
}

func TestMarshalText(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		in := http.AcceptHeaderNew()
		actual, err := in.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "", string(actual))
	})
	t.Run("one", func(t *testing.T) {
		in := http.AcceptHeaderNew("application/json")
		actual, err := in.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "application/json", string(actual))
	})
	t.Run("two", func(t *testing.T) {
		in := http.AcceptHeaderNew("application/json", "text/*")
		actual, err := in.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "application/json, text/*", string(actual))
	})
}

func TestUnmarshalText(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		err := actual.UnmarshalText([]byte{})
		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader(nil), actual)
	})
	t.Run("only whitespace", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		err := actual.UnmarshalText([]byte("  "))
		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader(nil), actual)
	})
	t.Run("one", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		err := actual.UnmarshalText([]byte("application/json"))
		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader([]string{"application/json"}), actual)
	})
	t.Run("two", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		err := actual.UnmarshalText([]byte("application/json,text/*"))
		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader([]string{"application/json", "text/*"}), actual)
	})
	t.Run("optional whitespace is ignored", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		err := actual.UnmarshalText([]byte("application/json, text/*"))
		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader([]string{"application/json", "text/*"}), actual)
	})
	t.Run("q-factor weighting is ignored", func(t *testing.T) {
		actual := http.AcceptHeaderNew()
		err := actual.UnmarshalText([]byte("application/json, text/*;q=0.7"))
		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader([]string{"application/json", "text/*"}), actual)
	})
}

type testEncoderDecoder struct {
	value string
	err   error
}

func (d *testEncoderDecoder) DecodeValue() (string, error) {
	return d.value, d.err
}

func (d testEncoderDecoder) DecodeArray(f func(uri.Decoder) error) error {
	panic("not implemented")
}

func (d testEncoderDecoder) DecodeFields(f func(string, uri.Decoder) error) error {
	panic("not implemented")
}

func (d *testEncoderDecoder) EncodeValue(v string) error {
	if d.err == nil {
		d.value = v
		return nil
	}
	return d.err
}
func (d testEncoderDecoder) EncodeArray(f func(e uri.Encoder) error) error {
	panic("not implemented")
}
func (d testEncoderDecoder) EncodeField(name string, f func(e uri.Encoder) error) error {
	panic("not implemented")
}

func TestDecodeURI(t *testing.T) {
	// No detailed tests on the actual decoding, that is covered by the test for UnmarshalText
	t.Run("value is decoded", func(t *testing.T) {
		decoder := &testEncoderDecoder{value: "application/json, text/*", err: nil}

		actual := http.AcceptHeaderNew()
		err := actual.DecodeURI(decoder)

		require.NoError(t, err)
		require.Equal(t, http.AcceptHeader([]string{"application/json", "text/*"}), actual)
	})
	t.Run("error is wrapped", func(t *testing.T) {
		expectedErr := errors.New("unit test error")
		decoder := &testEncoderDecoder{value: "application/json, text/*", err: expectedErr}

		actual := http.AcceptHeaderNew()
		err := actual.DecodeURI(decoder)

		require.ErrorIs(t, err, expectedErr)
		require.Empty(t, actual)
	})
}

func TestEncodeURI(t *testing.T) {
	// No detailed tests on the actual encoding, that is covered by the test for MarshalText
	t.Run("value is encoded", func(t *testing.T) {
		encoder := &testEncoderDecoder{value: "", err: nil}

		actual := http.AcceptHeaderNew("application/json", "text/*")
		err := actual.EncodeURI(encoder)

		require.NoError(t, err)
		require.Equal(t, "application/json, text/*", encoder.value)
	})
	t.Run("error is wrapped", func(t *testing.T) {
		expectedErr := errors.New("unit test error")
		encoder := &testEncoderDecoder{value: "", err: expectedErr}

		actual := http.AcceptHeaderNew("application/json", "text/*")
		err := actual.EncodeURI(encoder)

		require.ErrorIs(t, err, expectedErr)
		require.Empty(t, encoder.value)
	})
}

func TestMatchesContentType(t *testing.T) {
	t.Run("empty matches nothing", func(t *testing.T) {
		in := http.AcceptHeaderNew()
		actual := in.MatchesContentType("application/json")
		require.False(t, actual)
	})
	t.Run("single match", func(t *testing.T) {
		in := http.AcceptHeaderNew("application/json")
		actual := in.MatchesContentType("application/json")
		require.True(t, actual)
	})
	t.Run("single mismatch", func(t *testing.T) {
		in := http.AcceptHeaderNew("application/octet-stream")
		actual := in.MatchesContentType("application/json")
		require.False(t, actual)
	})
	t.Run("multiple match", func(t *testing.T) {
		in := http.AcceptHeaderNew("application/json", "text/*")
		actual := in.MatchesContentType("text/plain")
		require.True(t, actual)
	})
	t.Run("multiple mismatch", func(t *testing.T) {
		in := http.AcceptHeaderNew("application/octet-stream", "text/*")
		actual := in.MatchesContentType("application/json")
		require.False(t, actual)
	})
}
