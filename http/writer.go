package http

import (
	"io"
	"mime/multipart"

	"golang.org/x/sync/errgroup"
)

// CreateMultipartBody is helper for streaming multipart/form-data.
func CreateMultipartBody(cb func(mw *multipart.Writer) error) (
	getBody func() (io.ReadCloser, error),
	contentType string,
) {
	r, w := io.Pipe()
	mw := multipart.NewWriter(w)
	getBody = func() (io.ReadCloser, error) {
		wg := new(errgroup.Group)
		wg.Go(func() (rerr error) {
			defer w.Close()
			defer mw.Close()
			return cb(mw)
		})
		return bodyReader{
			mw: mw,
			r:  r,
			w:  w,
			wg: wg,
		}, nil
	}
	return getBody, mw.FormDataContentType()
}

type bodyReader struct {
	mw *multipart.Writer
	r  *io.PipeReader
	w  *io.PipeWriter
	wg *errgroup.Group
}

func (w bodyReader) Read(p []byte) (int, error) {
	return w.r.Read(p)
}

func (w bodyReader) Close() (rerr error) {
	rerr = w.r.Close()
	_ = w.wg.Wait()
	return rerr
}
