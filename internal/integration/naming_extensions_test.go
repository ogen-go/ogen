package integration

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_naming_extensions"
)

func TestNamingExtensions(t *testing.T) {
	input := `{
	"RenameField": "Field",
	"RefField": {
		"RenameField": "SubField"
	}
}`

	a := require.New(t)
	var p api.Person
	a.NoError(p.Decode(jx.DecodeStr(input)))
	a.Equal(p.Field, "Field")
	a.NotNil(p.Parent)
	a.Equal(p.Parent.Field, "SubField")
}
