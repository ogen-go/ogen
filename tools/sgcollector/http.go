package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-faster/errors"
)

func query(ctx context.Context, q Query) (SearchResult, error) {
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
