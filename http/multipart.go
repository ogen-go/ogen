package http

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
)

func randomBoundary() string {
	var buf [30]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

// CreateMultipartBody is helper for streaming multipart/form-data.
func CreateMultipartBody(cb func(mw *multipart.Writer) error) (body io.ReadCloser, boundary string) {
	boundary = randomBoundary()
	body = CreateBodyWriter(func(w io.Writer) error {
		mw := multipart.NewWriter(w)
		if err := mw.SetBoundary(boundary); err != nil {
			return err
		}
		defer func() {
			_ = mw.Close()
		}()
		return cb(mw)
	})
	return body, boundary
}
