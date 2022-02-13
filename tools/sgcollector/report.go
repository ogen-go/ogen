package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
)

type Report struct {
	File  FileMatch
	Error string
	Data  json.RawMessage
}

func schemasWriter(ctx context.Context, path string, invalidSchema <-chan Report) error {
	if err := os.MkdirAll(path, 0o666); err != nil {
		return err
	}
	i := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case invalid, ok := <-invalidSchema:
			if !ok {
				return nil
			}

			data, err := json.Marshal(invalid)
			if err != nil {
				return errors.Wrap(err, "encode error")
			}

			writePath := filepath.Join(path, fmt.Sprintf("%d.json", i))
			if err := os.WriteFile(writePath, data, 0o666); err != nil {
				return err
			}
			i++
		}
	}
}
