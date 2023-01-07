package ogen_test

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

func BenchmarkGenerator(b *testing.B) {
	files := []string{
		"api.github.com.json", // Giant (>90kLOC)
		"gotd_bot_api.json",   // Medium (>10kLOC)
		"tinkoff.json",        // Small (>2kLOC)
		"manga.json",          // Tiny (<1kLOC)
	}
	for _, file := range files {
		file := file
		name := strings.TrimSuffix(file, ".json")
		b.Run(name, func(b *testing.B) {
			a := require.New(b)
			data, err := testdata.ReadFile(path.Join("_testdata/examples", file))
			a.NoError(err)
			spec, err := ogen.Parse(data)
			a.NoError(err)

			opts := gen.Options{
				IgnoreNotImplemented: []string{
					"all",
				},
				InferSchemaType: true,
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
		})
	}
}
