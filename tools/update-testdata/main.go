// Binary update-schema is simple script to update ogen testdata.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
)

type Schema struct {
	File       string
	Link       string
	Output     string
	SkipReason string
}

func (s Schema) OutputDir() string {
	if o := s.Output; o != "" {
		return o
	}
	return fmt.Sprintf("ex_%s", strings.TrimSuffix(strings.ToLower(s.File), ".json"))
}

func (s Schema) GoGenerate(w io.Writer) {
	fmt.Fprintf(w, `//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/%s --target %s`,
		s.File, s.OutputDir())
	fmt.Fprintln(w, ` --infer-types --debug.noerr --clean`)
}

type skipBOMReader struct {
	met bool
	r   io.Reader
}

var bomPrefix = []byte{0xEF, 0xBB, 0xBF}

func (s *skipBOMReader) Read(p []byte) (n int, err error) {
	if s.met {
		return s.r.Read(p)
	}

	n, err = s.r.Read(p)
	if n == 0 {
		return
	}
	cut := bytes.TrimPrefix(p, bomPrefix)
	n = copy(p, cut)

	s.met = true
	return
}

func get(ctx context.Context, s Schema) error {
	dir := filepath.Dir(s.File)
	if err := os.MkdirAll(dir, 0o666); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Link, http.NoBody)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	f, err := os.Create(s.File)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer func() {
		_ = f.Close()
	}()

	r := skipBOMReader{
		r: resp.Body,
	}
	if _, err := io.Copy(f, &r); err != nil {
		return errors.Wrap(err, "copy")
	}

	return nil
}

func downloadSchemas(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	spawn := func(s Schema) {
		g.Go(func() error {
			s.File = filepath.Join("_testdata", "positive", s.File)
			if err := get(ctx, s); err != nil {
				return errors.Wrapf(err, "schema: %s", s.Output)
			}
			return nil
		})
	}
	for _, links := range linkSets {
		for _, s := range links {
			if s.SkipReason != "" {
				_ = os.Remove(filepath.Join("_testdata", "positive", s.File))
				continue
			}
			// Argument is copied.
			spawn(s)
		}
	}
	return g.Wait()
}

func run(ctx context.Context) error {
	if err := downloadSchemas(ctx); err != nil {
		return errors.Wrap(err, "download schemas")
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
