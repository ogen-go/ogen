package main

import "net/url"

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

	Query struct {
		Query     string `json:"query"`
		Variables QueryVariables
	}
)

func (m FileMatch) Link() string {
	return "https://" + m.Repository.Name + "/blob/-/" + url.PathEscape(m.File.Path)
}
