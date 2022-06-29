package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-faster/errors"
)

type (
	Repository struct {
		Name string `json:"name"`
	}

	File struct {
		Name     string `json:"name"`
		Size     int    `json:"size"`
		Path     string `json:"path"`
		ByteSize uint64 `json:"byteSize"`
		Content  string `json:"content"`
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

func (m FileMatch) Link() string {
	return "https://" + m.Repository.Name + "/blob/-/" + m.File.Path
}

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
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
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
