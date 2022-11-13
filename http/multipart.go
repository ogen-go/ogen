package http

import (
	"io"
	"mime/multipart"

	"golang.org/x/sync/errgroup"
)

// CreateMultipartBody is helper for streaming multipart/form-data.
func CreateMultipartBody(cb func(mw *multipart.Writer) error) (body io.ReadCloser, boundary string) {
	piper, pipew := io.Pipe()

	mw := multipart.NewWriter(pipew)
	wg := new(errgroup.Group)
	wg.Go(func() (rerr error) {
		defer func() {
			_ = mw.Close()
			if rerr != nil {
				_ = pipew.CloseWithError(rerr)
			} else {
				_ = pipew.Close()
			}
		}()
		return cb(mw)
	})

	body = bodyReader{
		mw: mw,
		r:  piper,
		w:  pipew,
		wg: wg,
	}
	boundary = mw.Boundary()
	return body, boundary
}

type bodyReader struct {
	mw *multipart.Writer
	r  *io.PipeReader
	w  *io.PipeWriter
	wg *errgroup.Group
}

func (r bodyReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r bodyReader) Close() (rerr error) {
	rerr = r.r.Close()
	_ = r.wg.Wait()
	return rerr
}
