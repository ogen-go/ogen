package main

type (
	Repository struct {
		Name string `json:"name"`
	}

	File struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Content string `json:"content"`
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
