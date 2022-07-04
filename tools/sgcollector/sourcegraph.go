package main

import (
	"bytes"
	"context"
	"net/http"

	"github.com/go-json-experiment/json"

	"github.com/go-faster/errors"
	"go.uber.org/zap/zapcore"
)

type (
	Repository struct {
		Name string `json:"name"`
	}

	ExternalURL struct {
		URL         string `json:"url"`
		ServiceKind string `json:"serviceKind"`
	}

	File struct {
		Name         string        `json:"name"`
		Size         int           `json:"size"`
		Path         string        `json:"path"`
		ByteSize     uint64        `json:"byteSize"`
		Content      string        `json:"content"`
		CanonicalURL string        `json:"canonicalURL"`
		ExternalURLs []ExternalURL `json:"externalURLs"`
	}

	FileMatch struct {
		Typename   string     `json:"__typename"`
		Repository Repository `json:"repository"`
		File       File       `json:"file"`
	}

	Alert struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	SearchResult struct {
		Matches             []FileMatch `json:"results"`
		LimitHit            bool        `json:"limitHit"`
		MatchCount          int         `json:"matchCount"`
		ElapsedMilliseconds int         `json:"elapsedMilliseconds"`
		Alert               Alert       `json:"alert"`
	}

	SearchResults struct {
		Results SearchResult `json:"results"`
	}

	Data struct {
		Search SearchResults `json:"search"`
	}

	Response struct {
		Data Data `json:"data"`
	}

	QueryVariables struct {
		Query string `json:"query"`
	}

	GraphQLQuery struct {
		Query     string `json:"query"`
		Variables QueryVariables
	}
)

func (m FileMatch) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("link", m.Link())
	return nil
}

func (m FileMatch) Link() string {
	for _, external := range m.File.ExternalURLs {
		return external.URL
	}
	return "https://sourcegraph.com" + m.File.CanonicalURL
}

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
    canonicalURL
    externalURLs {
      serviceKind
      url
    }
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

func search(ctx context.Context, query string) (SearchResult, error) {
	r, err := querySourcegraph(ctx, GraphQLQuery{
		Query: graphQLQuery,
		Variables: QueryVariables{
			Query: query,
		},
	})
	if err != nil {
		return r, errors.Wrapf(err, "query: %q", query)
	}
	return r, nil
}

func querySourcegraph(ctx context.Context, q GraphQLQuery) (SearchResult, error) {
	body, err := json.Marshal(q)
	if err != nil {
		return SearchResult{}, errors.Wrap(err, "encode")
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, "https://sourcegraph.com/.api/graphql",
		bytes.NewReader(body),
	)
	if err != nil {
		return SearchResult{}, errors.Wrap(err, "create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SearchResult{}, errors.Wrap(err, "do request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return SearchResult{}, errors.Errorf("http error: %s", resp.Status)
	}

	var r Response
	if err := json.UnmarshalFull(resp.Body, &r); err != nil {
		return SearchResult{}, errors.Wrap(err, "parse")
	}

	result := r.Data.Search.Results
	if a := result.Alert; a.Title != "" {
		alert := a.Title
		if a.Description != "" {
			alert = a.Description
		}
		return SearchResult{}, errors.Errorf("alert: %s", alert)
	}

	return result, nil
}
