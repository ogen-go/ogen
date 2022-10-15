package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

func TestParseForm(t *testing.T) {
	tests := []struct {
		input   string
		want    url.Values
		wantErr bool
	}{
		{"", url.Values{}, false},
		{"a=b", url.Values{"a": {"b"}}, false},
		{"a=b&c=d", url.Values{"a": {"b"}, "c": {"d"}}, false},
		{"a=b&a=c&d=e", url.Values{"a": {"b", "c"}, "d": {"e"}}, false},
		// Invalid form data.
		{"%", url.Values{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			got, err := ParseForm(&http.Request{
				Body: io.NopCloser(strings.NewReader(tt.input)),
			})
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.Equal(tt.want, got)
		})
	}
	t.Run("Error", func(t *testing.T) {
		testErr := errors.New("test error")
		_, err := ParseForm(&http.Request{
			Body: io.NopCloser(iotest.ErrReader(testErr)),
		})
		require.ErrorIs(t, err, testErr)
	})
	t.Run("PostFormIsSet", func(t *testing.T) {
		a := require.New(t)
		got, err := ParseForm(&http.Request{
			PostForm: url.Values{"a": {"b"}},
		})
		a.NoError(err)
		a.Equal(url.Values{"a": {"b"}}, got)
	})
}
