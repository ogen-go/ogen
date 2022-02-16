package main

type (
	Repository struct {
		Name        string `json:"name"`
		Language    string `json:"language"`
		Description string `json:"description"`
	}

	Commit struct {
		Oid string `json:"oid"`
	}

	File struct {
		Name     string `json:"name"`
		Size     int    `json:"size"`
		Path     string `json:"path"`
		ByteSize uint64 `json:"byteSize"`
		Commit   Commit `json:"commit"`
		Content  string `json:"content"`
	}

	FileMatch struct {
		Typename   string     `json:"__typename"`
		Repository Repository `json:"repository"`
		File       File       `json:"file"`
	}

	SearchResult struct {
		Matches             []FileMatch `json:"results"`
		LimitHit            bool        `json:"limitHit"`
		MatchCount          int         `json:"matchCount"`
		ElapsedMilliseconds int         `json:"elapsedMilliseconds"`
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
	return "https://" + m.Repository.Name + "/blob/-/" + m.File.Path
}
