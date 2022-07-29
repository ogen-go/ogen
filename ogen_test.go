package ogen_test

import (
	"embed"
	"encoding/json"
	"path"
	"strings"
	"testing"

	helperyaml "github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/testutil"
)

//go:embed _testdata
var testdata embed.FS

func walkTestdata(t *testing.T, root string, cb func(t *testing.T, file string, data []byte)) {
	t.Helper()
	testutil.WalkTestdata(t, testdata, root, cb)
}

func TestParse(t *testing.T) {
	testcb := func(t *testing.T, file string, data []byte) {
		t.Helper()
		a := require.New(t)

		var yamlInput, jsonInput []byte
		if strings.HasSuffix(file, ".json") {
			jsonInput = data
			val, err := helperyaml.JSONToYAML(data)
			a.NoError(err)
			yamlInput = val
		} else {
			yamlInput = data
			val, err := helperyaml.YAMLToJSON(data)
			a.NoError(err)
			jsonInput = val
		}

		jsonSpec, err := ogen.Parse(jsonInput)
		a.NoError(err)

		yamlSpec, err := ogen.Parse(yamlInput)
		a.NoError(err)

		{
			jsonOutput, err := json.Marshal(jsonSpec)
			a.NoError(err)

			yamlOutput, err := json.Marshal(yamlSpec)
			a.NoError(err)

			a.JSONEq(string(jsonOutput), string(yamlOutput))
		}
	}

	for _, dir := range []string{
		"positive",
		"negative",
		"examples",
	} {
		dir := dir
		t.Run(strings.ToTitle(dir[:1])+dir[1:], func(t *testing.T) {
			walkTestdata(t, path.Join("_testdata", dir), testcb)
		})
	}
}
