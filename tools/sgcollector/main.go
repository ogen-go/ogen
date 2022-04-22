package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
)

const graphQLQuery = `query ($query: String!) {
  search(query: $query, version: V2, patternType: regexp) {
    results {
      results {
        __typename
        ... on FileMatch {
          ...FileMatchFields
        }
      }
      limitHit
      matchCount
      elapsedMilliseconds
      ...SearchResultsAlertFields
    }
  }
}

fragment FileMatchFields on FileMatch {
  repository {
    name
  }
  file {
    name
    path
    byteSize
    content
  }
}


fragment SearchResultsAlertFields on SearchResults {
  alert {
    title
    description
    proposedQueries {
      description
      query
    }
  }
}
`

func run(ctx context.Context) error {
	var (
		output         = flag.String("output", "./corpus", "Path to corpus output")
		stats          = flag.String("stats", "", "Path to stats output")
		clean          = flag.Bool("clean", false, "Clean generated files before generation")
		generateYaml   = flag.Bool("yaml", false, "Query yaml files")
		q              = flag.String("query", "", "Sourcegraph query")
		cpuProfile     = flag.String("cpuprofile", "", "Write cpu profile to file")
		memProfile     = flag.String("memprofile", "", "Write memory profile to file")
		memProfileRate = flag.Int64("memprofilerate", 0, "Set runtime.MemProfileRate")
	)
	flag.Parse()

	if val := *cpuProfile; val != "" {
		f, err := os.Create(val)
		if err != nil {
			return errors.Wrap(err, "create CPU profile")
		}
		defer f.Close()

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
		defer f.Close()

		defer func() {
			runtime.GC() // get up-to-date statistics
			_ = pprof.WriteHeapProfile(f)
		}()
	}

	var queries []string
	if *q != "" {
		queries = []string{*q}
	} else {
		queries = []string{
			`(openapi|"openapi"):\s?"3 file:.*\.yml$`,
			`(openapi|"openapi"):\s+3 file:.*\.yml$`,
			`(openapi|"openapi"):\s?"3 file:.*\.yaml$`,
			`(openapi|"openapi"):\s+3 file:.*\.yaml$`,
			`"openapi":"3 file:.*\.json$`,
			`"openapi":\s+"3 file:.*\.json$`,
		}
		for i := range queries {
			queries[i] += ` count:all -repo:^github\.com/ogen-go/corpus$`
		}
		if !(*generateYaml) {
			queries = queries[2:]
		}
	}

	var (
		workers   = 1
		links     = make(chan FileMatch, workers)
		reporters = &Reporters{}
		total     int
	)
	reporters.init(workers)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(links)

		for i, q := range queries {
			results, err := query(ctx, Query{
				Query: graphQLQuery,
				Variables: QueryVariables{
					Query: q,
				},
			})
			if err != nil {
				return errors.Wrapf(err, "query: %d", i)
			}

			for _, m := range results.Matches {
				select {
				case <-ctx.Done():
					return nil
				case links <- m:
					total++
				}
			}
		}
		return nil
	})
	g.Go(func() error {
		return reporters.run(ctx, *clean, *output)
	})

	var workersWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		workersWg.Add(1)
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

					link := m.Link()
					fmt.Printf("Checking %q\n", link)
					if err := worker(ctx, m, reporters); err != nil {
						fmt.Printf("Check %q: %v\n", link, err)
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
		panic(err)
	}
}
