package ir

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/openapi"
)

// Tag of Field or Parameter.
type Tag struct {
	JSON      string             // json tag, empty for none
	Form      *openapi.Parameter // query form parameter
	ExtraTags map[string]string  // a map of extra struct field tags
}

// EscapedJSON returns quoted and escaped JSON tag.
func (t Tag) EscapedJSON() string {
	return strconv.Quote(t.JSON)
}

// GetTags returns a formatted list of struct tags, which must be quoted by '`'
func (t Tag) GetTags() string {
	tags := make([]string, 0, 1+len(t.ExtraTags))
	if t.JSON != "" {
		tags = append(tags, fmt.Sprintf(`%s:%q`, "json", t.JSON))
	}
	if t.ExtraTags != nil {
		for _, k := range xmaps.SortedKeys(t.ExtraTags) {
			tags = append(tags, fmt.Sprintf(`%s:%q`, k, t.ExtraTags[k]))
		}
	}

	return strings.Join(tags, " ")
}
