package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
)

type Report struct {
	File           FileMatch
	Error          string
	NotImplemented []string          `json:",omitempty"`
	Hash           [sha256.Size]byte `json:"-"`
}

type Reporter struct {
	ch      chan Report
	counter int
}

func (r *Reporter) run(ctx context.Context, path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case report, ok := <-r.ch:
			if !ok {
				return nil
			}

			data, err := json.MarshalIndent(report, "", "\t")
			if err != nil {
				return errors.Wrap(err, "encode error")
			}

			writePath := filepath.Join(path, fmt.Sprintf("%x.json", report.Hash))
			if err := os.WriteFile(writePath, data, 0o750); err != nil {
				return err
			}
			r.counter++
		}
	}
}

func (r *Reporter) close() {
	close(r.ch)
}

type Reporters struct {
	reporters [last]*Reporter
}

func (r *Reporters) init(buf int) {
	for i := range r.reporters {
		r.reporters[i] = &Reporter{
			ch: make(chan Report, buf),
		}
	}
}

func (r *Reporters) close() {
	for _, reporter := range r.reporters {
		reporter.close()
	}
}

func (r *Reporters) run(ctx context.Context, clean bool, path string) error {
	g, ctx := errgroup.WithContext(ctx)

	if clean {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(path, 0o750); err != nil {
		return err
	}

	spawn := func(name string, reporter *Reporter) {
		g.Go(func() error {
			if err := reporter.run(ctx, filepath.Join(path, name)); err != nil {
				return errors.Wrap(err, name)
			}
			return nil
		})
	}
	for idx := range r.reporters {
		spawn(Stage(idx).String(), r.reporters[idx])
	}
	return g.Wait()
}

func (r *Reporters) report(ctx context.Context, stage Stage, report Report) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case r.reporters[stage].ch <- report:
		return nil
	}
}

func (r *Reporters) writeStats(output string, total int) error {
	output = filepath.Clean(output)
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	w := tabwriter.NewWriter(f, 0, 0, 1, ' ', 0)
	for idx, reporter := range r.reporters {
		name := Stage(idx).String()
		if _, err := fmt.Fprintf(w, "%s\t%d\n", name, reporter.counter); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%s\t%d\n", "total", total); err != nil {
		return err
	}
	return w.Flush()
}
