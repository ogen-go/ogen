package http

import (
	"io"
	"mime/multipart"
)

// MultipartFile is multipart form file.
type MultipartFile struct {
	Name string
	File io.Reader
}

// WriteMultipart writes data from reader to given multipart.Writer as a form file.
func (m MultipartFile) WriteMultipart(fieldName string, w *multipart.Writer) error {
	fw, err := w.CreateFormFile(fieldName, m.Name)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, m.File)
	return err
}
