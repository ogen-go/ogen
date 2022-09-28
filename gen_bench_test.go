package ogen_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

func BenchmarkGenerator(b *testing.B) {
	a := require.New(b)
	data, err := testdata.ReadFile("_testdata/examples/firecracker.json")
	a.NoError(err)
	spec, err := ogen.Parse(data)
	a.NoError(err)

	opts := gen.Options{
		IgnoreNotImplemented: []string{
			"all",
		},
	}
	dir := b.TempDir()
	fs := genfs.FormattedSource{
		Root: dir,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g, err := gen.NewGenerator(spec, opts)
		if err != nil {
			b.Fatal(err)
		}
		if err := g.WriteSource(fs, "api"); err != nil {
			b.Fatal(err)
		}
	}
}
