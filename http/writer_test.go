package http

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureMultipartClose(t *testing.T) {
	a := require.New(t)
	getBody, _ := CreateMultipartBody(func(mw *multipart.Writer) error {
		return mw.WriteField("key", "value")
	})

	r, err := getBody()
	a.NoError(err)
	a.NoError(r.Close())
}
