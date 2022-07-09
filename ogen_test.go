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
)

//go:embed _testdata
var testdata embed.FS

func walkTestdata(t *testing.T, root string, cb func(t *testing.T, file string, data []byte)) {
	t.Helper()

	dir, err := testdata.ReadDir(root)
	require.NoError(t, err)

	for _, e := range dir {
		entryName := e.Name()
		filePath := path.Join(root, entryName)
		if e.IsDir() {
			t.Run(entryName, func(t *testing.T) {
				walkTestdata(t, filePath, cb)
			})
			continue
		}

		testName := strings.TrimSuffix(entryName, ".json")
		testName = strings.TrimSuffix(testName, ".yml")
		testName = strings.TrimSuffix(testName, ".yaml")

		t.Run(testName, func(t *testing.T) {
			data, err := testdata.ReadFile(filePath)
			require.NoError(t, err)
			cb(t, filePath, data)
		})
	}
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
			jsonToJsonOutput, err := json.Marshal(jsonSpec)
			a.NoError(err)

			yamlToJsonOutput, err := json.Marshal(yamlSpec)
			a.NoError(err)

			a.JSONEq(string(jsonToJsonOutput), string(yamlToJsonOutput))
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
