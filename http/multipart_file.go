package http

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"

	"golang.org/x/exp/slices"
)

// MultipartFile is multipart form file.
type MultipartFile struct {
	Name   string
	File   io.Reader
	Header textproto.MIMEHeader
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// headers generates headers for multipart form file, similar to CreateFormFile, but this function does not
// overwrite Content-Type if it is already set.
func (m MultipartFile) headers(fieldName string) (h textproto.MIMEHeader) {
	h = make(textproto.MIMEHeader, len(m.Header)+2)
	for k, v := range m.Header {
		h[k] = slices.Clone(v)
	}

	disposition := fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		escapeQuotes(fieldName), escapeQuotes(m.Name))
	h.Set("Content-Disposition", disposition)
	if _, ok := h["Content-Type"]; !ok {
		h.Set("Content-Type", "application/octet-stream")
	}
	return h
}

// WriteMultipart writes data from reader to given multipart.Writer as a form file.
func (m MultipartFile) WriteMultipart(fieldName string, w *multipart.Writer) error {
	p, err := w.CreatePart(m.headers(fieldName))
	if err != nil {
		return err
	}
	_, err = io.Copy(p, m.File)
	return err
}
