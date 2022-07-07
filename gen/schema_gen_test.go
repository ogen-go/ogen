package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func TestSchemaGenAnyWarn(t *testing.T) {
	a := require.New(t)

	core, ob := observer.New(zap.InfoLevel)
	s := newSchemaGen("", func(ref string) (*ir.Type, bool) {
		return nil, false
	})
	s.log = zap.New(core)

	_, err := s.generate("foo", &jsonschema.Schema{
		Type: "",
	}, false)
	a.NoError(err)

	entries := ob.FilterMessage("Type is not defined, using any").All()
	a.Len(entries, 1)
	args := entries[0].ContextMap()
	a.Equal("foo", args["name"])
}
