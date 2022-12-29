package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/ogen-go/ogen/internal/ogenzap"
)

func run(ctx context.Context) error {
	var (
		output = flag.String("output", "./corpus", "Path to corpus output")
		clean  = flag.Bool("clean", false, "Clean generated files before generation")
		stats  = flag.String("stats", "", "Path to stats output")

		generateYaml = flag.Bool("yaml", false, "Query yaml files")
		q            = flag.String("query", "", "Sourcegraph query")
		filter       = flag.String("filter", "", "Additional filter to concatenate to the query")

		workers = flag.Int("workers", runtime.GOMAXPROCS(-1), "Number of generator workers to spawn")

		cpuProfile     = flag.String("cpuprofile", "", "Write cpu profile to file")
		memProfile     = flag.String("memprofile", "", "Write memory profile to file")
		memProfileRate = flag.Int64("memprofilerate", 0, "Set runtime.MemProfileRate")

		logOptions ogenzap.Options
	)
	logOptions.RegisterFlags(flag.CommandLine)
	flag.Parse()

	if val := *cpuProfile; val != "" {
		f, err := os.Create(val)
		if err != nil {
			return errors.Wrap(err, "create CPU profile")
		}
		defer func() {
			_ = f.Close()
		}()

		if err := pprof.StartCPUProfile(f); err != nil {
			return errors.Wrap(err, "start CPU profile")
		}
		defer pprof.StopCPUProfile()
	}
	if val := *memProfile; val != "" {
		if *memProfileRate != 0 {
			runtime.MemProfileRate = int(*memProfileRate)
		}

		f, err := os.Create(val)
		if err != nil {
			return errors.Wrap(err, "create memory profile")
		}
		defer func() {
			_ = f.Close()
		}()

		defer func() {
			runtime.GC() // get up-to-date statistics
			_ = pprof.WriteHeapProfile(f)
		}()
	}

	logger, err := ogenzap.Create(logOptions)
	if err != nil {
		return err
	}
	defer func() {
		_ = logger.Sync()
	}()

	var queries []string
	if *q != "" {
		queries = []string{*q}
	} else {
		queries = []string{
			`"openapi":"3 file:.*\.json$`,
			`"openapi":\s+"3 file:.*\.json$`,
		}
		if *generateYaml {
			queries = append(queries,
				`(openapi|"openapi"):\s?"3 file:.*\.yml$`,
				`(openapi|"openapi"):\s+3 file:.*\.yml$`,
				`(openapi|"openapi"):\s?"3 file:.*\.yaml$`,
				`(openapi|"openapi"):\s+3 file:.*\.yaml$`,
			)
		}
		for i := range queries {
			queries[i] += ` count:all -repo:^github\.com/ogen-go/corpus$`
		}
	}
	if f := *filter; f != "" {
		for i := range queries {
			queries[i] += " " + f
		}
	}

	var (
		links     = make(chan FileMatch, *workers)
		reporters = &Reporters{}
		total     int
	)
	reporters.init(*workers)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(links)

		for _, q := range queries {
			logger := logger.With(zap.String("query", q))

			logger.Info("Start query")
			counter := 0
			if err := search(ctx, q, func(match FileMatch) error {
				select {
				case links <- match:
					total++
					counter++
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}); err != nil {
				return err
			}
			logger.Info("Query complete", zap.Int("total", counter))
		}
		return nil
	})
	g.Go(func() error {
		return reporters.run(ctx, *clean, *output)
	})

	var workersWg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		workersWg.Add(1)
		logger := logger.Named("worker" + strconv.Itoa(i))
		g.Go(func() error {
			defer workersWg.Done()
			for {
				select {
				case <-ctx.Done():
					return nil
				case m, ok := <-links:
					if !ok {
						return nil
					}

					logger.Debug("Processing link", zap.Inline(m))
					err := worker(ctx, m, reporters)
					if err != nil {
						logger.Error("Error",
							zap.Inline(m),
							zap.Error(noVerboseError{err: err}),
						)
					} else {
						logger.Debug("Success",
							zap.Inline(m),
						)
					}
				}
			}
		})
	}
	// Wait until all writers stopped.
	workersWg.Wait()
	// Close readers.
	reporters.close()

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "wait")
	}

	if o := *stats; o != "" {
		if err := reporters.writeStats(o, total); err != nil {
			return errors.Wrap(err, "write stats")
		}
	}
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
