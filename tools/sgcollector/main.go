package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
)

const graphQLQuery = `query ($query: String!) {
  search(query: $query, version: V2) {
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

const sgQuery = `"openapi": "3. file:.*\.json -file:(^|/)vendor/ count:2000` +
	` -file:.*petstore.*` +
	` -repo:^github\.com/Redocly/redoc$` +
	` -repo:^github\.com/kubernetes/kubernetes$` +
	` -repo:^github\.com/aws/serverless-application-model$`

func run(ctx context.Context) error {
	resp, err := query(ctx, Query{
		Query: graphQLQuery,
		Variables: QueryVariables{
			Query: sgQuery,
		},
	})
	if err != nil {
		return errors.Wrap(err, "query")
	}

	var (
		results = resp.Data.Search.Results

		workers       = runtime.GOMAXPROCS(-1)
		links         = make(chan FileMatch, workers)
		invalidSchema = make(chan Report)
		crashSchema   = make(chan Report)
	)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(links)
		defer close(invalidSchema)

		for _, m := range results.Matches {
			if !strings.HasPrefix(m.Repository.Name, "github.com") {
				continue
			}
			select {
			case <-ctx.Done():
				return nil
			case links <- m:
			}
		}
		return nil
	})
	g.Go(func() error {
		if err := schemasWriter(ctx, "invalid_schemas", invalidSchema); err != nil {
			return errors.Wrap(err, "invalid schemas")
		}
		return nil
	})
	g.Go(func() error {
		if err := schemasWriter(ctx, "crash_schemas", crashSchema); err != nil {
			return errors.Wrap(err, "crash schemas")
		}
		return nil
	})

	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case link, ok := <-links:
					if !ok {
						return nil
					}
					if err := worker(ctx, link, invalidSchema, crashSchema); err != nil {
						fmt.Printf("Check %q: %v\n", link.Link(), err)
					}
				}
			}
		})
	}

	return g.Wait()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
