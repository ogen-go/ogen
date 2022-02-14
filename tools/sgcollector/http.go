package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-faster/errors"
)

func query(ctx context.Context, q Query) (Response, error) {
	body, err := json.Marshal(q)
	if err != nil {
		return Response{}, errors.Wrap(err, "encode")
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, "https://sourcegraph.com/.api/graphql",
		bytes.NewReader(body),
	)
	if err != nil {
		return Response{}, errors.Wrap(err, "create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, errors.Wrap(err, "do request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Response{}, errors.Wrap(err, "parse")
	}

	return r, nil
}

var bomPrefix = []byte{0xEF, 0xBB, 0xBF}

func get(ctx context.Context, link string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes.TrimPrefix(data, bomPrefix), nil
}
