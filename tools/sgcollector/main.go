package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
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
    language
    description
  }
  file {
    name
    path
    byteSize
    commit {
      oid
    }
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
		output       = flag.String("output", "./corpus", "Path to corpus output")
		stats        = flag.String("stats", "", "Path to stats output")
		clean        = flag.Bool("clean", false, "Clean generated files before generation")
		generateYaml = flag.Bool("yaml", false, "Query yaml files")
		q            = flag.String("query", "", "Sourcegraph query")
	)
	flag.Parse()

	var queries []string
	if *q != "" {
		queries = []string{*q}
	} else {
		queries = []string{
			`(openapi|"openapi"):\s(3|"3) file:.*\.yml$ count:20000`,
			`(openapi|"openapi"):\s(3|"3) file:.*\.yaml$ count:20000`,
			`"openapi":\s(3|"3) file:.*\.json$ count:20000`,
		}
		if !(*generateYaml) {
			queries = queries[2:]
		}
	}

	var (
		workers   = runtime.GOMAXPROCS(-1)
		links     = make(chan FileMatch, workers)
		reporters = &Reporters{}
		total     int
	)
	reporters.init(workers)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(links)

		for i, q := range queries {
			resp, err := query(ctx, Query{
				Query: graphQLQuery,
				Variables: QueryVariables{
					Query: q,
				},
			})
			if err != nil {
				return errors.Wrapf(err, "query: %d", i)
			}
			results := resp.Data.Search.Results

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
