package ogen_test

import (
	"embed"
	"testing"

	"github.com/ogen-go/ogen/internal/testutil"
)

//go:embed _testdata
var testdata embed.FS

func walkTestdata(t *testing.T, root string, cb func(t *testing.T, file string, data []byte)) {
	t.Helper()
	testutil.WalkTestdata(t, testdata, root, cb)
}
