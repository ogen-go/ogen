package gen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func TestFeatureSetEnableForGoVersion(t *testing.T) {
	for _, tt := range []struct {
		name      string
		version   string
		cfg       *FeatureOptions
		want      bool
		wantError bool
	}{
		{name: "Empty", version: "", want: false},
		{name: "Old", version: "1.26", want: false},
		{name: "OldPatch", version: "1.26.3", want: false},
		{name: "Min", version: "1.27", want: true},
		{name: "MinPatch", version: "1.27.1", want: true},
		{name: "Prefixed", version: "go1.27", want: true},
		{name: "Newer", version: "1.28", want: true},
		{name: "Disabled", version: "1.27", cfg: &FeatureOptions{
			Disable: FeatureSet{FormatStdlibUUID.Name: {}},
		}, want: false},
		{name: "DisableAll", version: "1.27", cfg: &FeatureOptions{DisableAll: true}, want: false},
		{name: "Invalid", version: "latest", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			set, err := tt.cfg.Build()
			require.NoError(t, err)

			err = set.enableForGoVersion(tt.cfg, tt.version)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, set.Has(FormatStdlibUUID))
		})
	}
}

func TestStdlibUUID(t *testing.T) {
	uuidSchema := &ogen.Schema{Type: "string", Format: "uuid"}
	spec := &ogen.Spec{
		OpenAPI: "3.0.3",
		Info:    ogen.Info{Title: "test", Version: "1.0.0"},
		Paths: ogen.Paths{
			"/items/{id}": &ogen.PathItem{
				Post: &ogen.Operation{
					OperationID: "postItem",
					Parameters: []*ogen.Parameter{
						{Name: "id", In: "path", Required: true, Schema: uuidSchema},
						{Name: "ids", In: "query", Schema: &ogen.Schema{Type: "array", Items: &ogen.Items{Item: uuidSchema}}},
					},
					Responses: ogen.Responses{
						"200": &ogen.Response{
							Description: "ok",
							Content: map[string]ogen.Media{
								"application/json": {Schema: &ogen.Schema{
									Type:     "object",
									Required: []string{"id"},
									Properties: []ogen.Property{
										{Name: "id", Schema: uuidSchema},
										{Name: "parent", Schema: &ogen.Schema{Type: "string", Format: "uuid", Nullable: true}},
									},
								}},
							},
						},
					},
				},
			},
		},
	}

	generate := func(t *testing.T, opts GenerateOptions) string {
		g, err := NewGenerator(spec, Options{Generator: opts})
		require.NoError(t, err)

		fs := &recordFS{files: map[string][]byte{}}
		require.NoError(t, g.WriteSource(fs, "api"))

		var sb strings.Builder
		for _, content := range fs.files {
			sb.Write(content)
		}
		return sb.String()
	}

	t.Run("Google", func(t *testing.T) {
		for _, opts := range []GenerateOptions{
			{},
			{GoVersion: "1.26"},
			{GoVersion: "1.27", Features: &FeatureOptions{Disable: FeatureSet{FormatStdlibUUID.Name: {}}}},
		} {
			src := generate(t, opts)
			require.Contains(t, src, `"github.com/google/uuid"`)
			require.NotContains(t, src, "UUIDOf")
		}
	})
	t.Run("Stdlib", func(t *testing.T) {
		for _, opts := range []GenerateOptions{
			{GoVersion: "1.27"},
			{Features: &FeatureOptions{Enable: FeatureSet{FormatStdlibUUID.Name: {}}}},
		} {
			src := generate(t, opts)
			require.NotContains(t, src, `"github.com/google/uuid"`)
			for _, s := range []string{
				`"uuid"`,
				"json.EncodeUUIDOf(",
				"json.DecodeUUIDOf[uuid.UUID](",
				"conv.UUIDOfToString(",
				"conv.ToUUIDOf[uuid.UUID](",
			} {
				require.Contains(t, src, s)
			}
		}
	})
}
