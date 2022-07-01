package http

import (
	"io"
	"mime/multipart"

	"golang.org/x/sync/errgroup"
)

// CreateMultipartBody is helper for streaming multipart/form-data.
func CreateMultipartBody(cb func(mw *multipart.Writer) error) (body io.ReadCloser, contentType string) {
	piper, pipew := io.Pipe()

	mw := multipart.NewWriter(pipew)
	wg := new(errgroup.Group)
	wg.Go(func() (rerr error) {
		defer func() {
			_ = mw.Close()
			_ = pipew.Close()
		}()
		return cb(mw)
	})

	body = bodyReader{
		mw: mw,
		r:  piper,
		w:  pipew,
		wg: wg,
	}
	contentType = mw.FormDataContentType()
	return body, contentType
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
