package ir

const (
	lineLimit = 100
)

func prettyDoc(s, deprecation string) (r []string) {
	r = renderMarkdown(s, lineLimit)
	if deprecation != "" {
		if len(r) > 0 {
			// Insert empty line between description and deprecated notice.
			r = append(r, "")
		}
		r = append(r, deprecation)
	}

	return r
}
