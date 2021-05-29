package ogen

type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Spec struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	TermsOfService string   `json:"termsOfService"`
	Contact        *Contact `json:"contact"`
	License        *License `json:"license"`
	Version        string   `json:"version"`
}
