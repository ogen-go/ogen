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

func BenchmarkParseForm(b *testing.B) {
	bench := func(body string, parse func(r *http.Request) error) func(*testing.B) {
		return func(b *testing.B) {
			sr := strings.NewReader(body)
			r := &http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(sr),
				Header: http.Header{},
			}
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			b.SetBytes(int64(len(body)))
			b.ReportAllocs()
			b.ResetTimer()

			var sink error
			for i := 0; i < b.N; i++ {
				r.Form = nil
				r.PostForm = nil
				sr.Reset(body)

				sink = parse(r)
			}
			if sink != nil {
				b.Fatal(sink)
			}
		}
	}

	// ~12KB of form data.
	body := func() string {
		var sb strings.Builder
		for i := 0; i < 1024; i++ {
			if i > 0 {
				sb.WriteString("&")
			}
			_, _ = fmt.Fprintf(&sb, "a%04d=b%04d", i, i)
		}
		return sb.String()
	}()

	b.Run("Custom", bench(body, func(r *http.Request) error {
		_, err := ParseForm(r)
		return err
	}))
	b.Run("Std", bench(body, (*http.Request).ParseForm))
}
