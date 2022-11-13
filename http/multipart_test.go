package http

import (
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

func TestCreateMultipartBody(t *testing.T) {
	t.Run("WriteError", func(t *testing.T) {
		testErr := errors.New("test error")
		reader, _ := CreateMultipartBody(func(mw *multipart.Writer) error {
			return testErr
		})

		_, err := io.ReadAll(reader)
		require.ErrorIs(t, err, testErr)
	})
	t.Run("Good", func(t *testing.T) {
		reader, boundary := CreateMultipartBody(func(mw *multipart.Writer) error {
			w, err := mw.CreateFormFile("file", "file.txt")
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, "text data")
			return err
		})

		a := require.New(t)
		var s strings.Builder
		_, err := io.Copy(&s, reader)
		a.NoError(err)

		expected := fmt.Sprintf("--%[1]s"+"\r\n"+
			`Content-Disposition: form-data; name="file"; filename="file.txt"`+"\r\n"+
			`Content-Type: application/octet-stream`+"\r\n\r\n"+
			`text data`+"\r\n"+
			`--%[1]s--`+"\r\n", boundary)
		a.Equal(expected, s.String())
	})
}
