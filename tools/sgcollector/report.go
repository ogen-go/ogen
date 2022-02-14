package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
)

type Report struct {
	File  FileMatch
	Error string
}

type Reporters struct {
	InvalidJSON chan Report
	Parse       chan Report
	Build       chan Report
	Template    chan Report
	Format      chan Report
	Crash       chan Report
}

func (r *Reporters) init() {
	r.InvalidJSON = make(chan Report)
	r.Parse = make(chan Report)
	r.Build = make(chan Report)
	r.Template = make(chan Report)
	r.Format = make(chan Report)
	r.Crash = make(chan Report)
}

func (r *Reporters) close() {
	close(r.InvalidJSON)
	close(r.Parse)
	close(r.Build)
	close(r.Template)
	close(r.Format)
	close(r.Crash)
}

func (r *Reporters) spawn(ctx context.Context, path string) error {
	g, ctx := errgroup.WithContext(ctx)

	mapping := map[string]chan Report{
		"invalidJSON": r.InvalidJSON,
		"parse":       r.Parse,
		"build":       r.Build,
		"template":    r.Template,
		"format":      r.Format,
		"crash":       r.Crash,
	}

	if err := os.MkdirAll(path, 0o666); err != nil {
		return err
	}

	spawn := func(name string, ch chan Report) {
		g.Go(func() error {
			if err := schemasWriter(ctx, filepath.Join(path, name), ch); err != nil {
				return errors.Wrap(err, name)
			}
			return nil
		})
	}
	for name, ch := range mapping {
		spawn(name, ch)
	}
	return g.Wait()
}

func schemasWriter(ctx context.Context, path string, r <-chan Report) error {
	if err := os.MkdirAll(path, 0o666); err != nil {
		return err
	}
	i := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case invalid, ok := <-r:
			if !ok {
				return nil
			}

			data, err := json.MarshalIndent(invalid, "", "\t")
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
