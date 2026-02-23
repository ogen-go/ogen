package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
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

func search(ctx context.Context, query string, cb func(FileMatch) error) (rerr error) {
	unmarshal := func(d *jx.Decoder, v any) error {
		raw, err := d.Raw()
		if err != nil {
			return errors.Wrap(err, "scan")
		}

		if err := json.Unmarshal(raw, v); err != nil {
			return errors.Wrap(err, "unmarshal")
		}

		return nil
	}

	err := querySourcegraph(ctx, GraphQLQuery{
		Query: graphQLQuery,
		Variables: QueryVariables{
			Query: query,
		},
	}, func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "results":
			return d.Arr(func(d *jx.Decoder) error {
				var fm FileMatch
				if err := unmarshal(d, &fm); err != nil {
					return errors.Wrap(err, "parse FileMatch")
				}

				if err := cb(fm); err != nil {
					return errors.Wrap(err, "callback")
				}
				return nil
			})
		case "alert":
			var a Alert
			if err := unmarshal(d, &a); err != nil {
				return errors.Wrap(err, "parse Alert")
			}

			if a.Title != "" {
				alert := a.Title
				if a.Description != "" {
					alert = a.Description
				}
				return errors.Errorf("alert: %s", alert)
			}

			return nil
		default:
			return d.Skip()
		}
	})
	if err != nil {
		return errors.Wrapf(err, "query: %q", query)
	}
	return nil
}

func querySourcegraph(ctx context.Context, q GraphQLQuery, cb func(d *jx.Decoder, key []byte) error) error {
	// Handling of the response may take a while, so we save the response to avoid timeouts.
	f, err := os.CreateTemp("", "sgcollector*")
	if err != nil {
		return errors.Wrap(err, "create temp file")
	}
	defer func() {
		_ = f.Close()
	}()

	if err := sendSourcegraph(ctx, q, f); err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, "seek")
	}

	d := jx.Decode(f, 4096)
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "data":
			return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
				switch string(key) {
				case "search":
					return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
						switch string(key) {
						case "results":
							return d.ObjBytes(cb)
						default:
							return d.Skip()
						}
					})
				default:
					return d.Skip()
				}
			})
		default:
			return d.Skip()
		}
	})
}

func sendSourcegraph(ctx context.Context, q GraphQLQuery, out io.Writer) error {
	body, err := json.Marshal(q)
	if err != nil {
		return errors.Wrap(err, "encode query")
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, "https://sourcegraph.com/.api/graphql",
		bytes.NewReader(body),
	)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	//#nosec G704
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return errors.Errorf("http error: %s", resp.Status)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return errors.Wrap(err, "copy response")
	}
	return nil
}
