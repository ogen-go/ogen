package gen

import (
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

// TestInitialismsYAMLDecode locks the nil-vs-empty contract that the inherit
// semantics rely on: an omitted list decodes to nil (use built-in defaults),
// while an explicit empty list decodes to a non-nil empty slice (disable all).
func TestInitialismsYAMLDecode(t *testing.T) {
	decode := func(t *testing.T, src string) Initialisms {
		var o GenerateOptions
		require.NoError(t, yaml.Unmarshal([]byte(src), &o))
		return o.Initialisms
	}

	t.Run("Omitted", func(t *testing.T) {
		require.Nil(t, decode(t, "convenient_errors: auto"))
	})
	t.Run("ExplicitEmpty", func(t *testing.T) {
		got := decode(t, "initialisms: []")
		require.NotNil(t, got)
		require.Empty(t, got)
	})
	t.Run("List", func(t *testing.T) {
		require.Equal(t, Initialisms{InitialismsInherit, "FQDN"}, decode(t, "initialisms: [inherit, FQDN]"))
	})
}

func TestInitialismsFeatureE2E(t *testing.T) {
	objSchema := &ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{Name: "userId", Schema: &ogen.Schema{Type: "string"}},
			{Name: "httpUrl", Schema: &ogen.Schema{Type: "string"}},
		},
	}
	spec := &ogen.Spec{
		OpenAPI: "3.0.3",
		Info:    ogen.Info{Title: "test", Version: "1.0.0"},
		Paths: ogen.Paths{
			"/obj": &ogen.PathItem{
				Get: &ogen.Operation{
					OperationID: "getObj",
					Responses: ogen.Responses{
						"200": &ogen.Response{
							Description: "ok",
							Content: map[string]ogen.Media{
								"application/json": {Schema: objSchema},
							},
						},
					},
				},
			},
		},
	}

	gen := func(t *testing.T, enable bool) map[string]struct{} {
		opts := Options{}
		if enable {
			opts.Generator.Features = &FeatureOptions{Enable: FeatureSet{NamingInitialisms.Name: {}}}
		}
		g, err := NewGenerator(spec, opts)
		require.NoError(t, err)
		fields := map[string]struct{}{}
		for _, typ := range g.tstorage.types {
			for _, f := range typ.Fields {
				fields[f.Name] = struct{}{}
			}
		}
		return fields
	}

	off := gen(t, false)
	require.Contains(t, off, "UserId")
	require.Contains(t, off, "HttpUrl")

	on := gen(t, true)
	require.Contains(t, on, "UserID")
	require.Contains(t, on, "HTTPURL")
}

// TestInitialismsVariantNames ensures the feature also reaches enum and
// discriminator variant naming, not only struct fields.
func TestInitialismsVariantNames(t *testing.T) {
	enumName := func(t *testing.T, initialisms bool) string {
		gen, err := namer{initialisms: initialisms}.enumVariantNameGen("Status", []any{"userId"})
		require.NoError(t, err)
		name, err := gen("userId", 0)
		require.NoError(t, err)
		return name
	}
	discriminatorName := func(t *testing.T, initialisms bool) string {
		gen, err := namer{initialisms: initialisms}.discriminatorMappingNameGen("Pet", []string{"userId"})
		require.NoError(t, err)
		name, err := gen("userId", 0)
		require.NoError(t, err)
		return name
	}

	t.Run("Enum", func(t *testing.T) {
		require.Equal(t, "StatusUserId", enumName(t, false))
		require.Equal(t, "StatusUserID", enumName(t, true))
	})
	t.Run("Discriminator", func(t *testing.T) {
		require.Equal(t, "PetUserId", discriminatorName(t, false))
		require.Equal(t, "PetUserID", discriminatorName(t, true))
	})
}

func TestInitialismsBuild(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		rs, err := Initialisms(nil).build()
		require.NoError(t, err)
		require.Nil(t, rs, "omitted list should fall back to the package default")
	})
	t.Run("ExplicitEmptyIsZero", func(t *testing.T) {
		rs, err := Initialisms{}.build()
		require.NoError(t, err)
		require.NotNil(t, rs, "explicit empty list should build an explicit ruleset")
		_, ok := rs.Rule("id")
		require.False(t, ok, "no initialisms should be applied")
	})
	t.Run("Inherit", func(t *testing.T) {
		rs, err := Initialisms{InitialismsInherit, "FQDN"}.build()
		require.NoError(t, err)
		require.NotNil(t, rs)
		got, ok := rs.Rule("fqdn")
		require.True(t, ok)
		require.Equal(t, "FQDN", got)
		// Built-in defaults are spliced in by "inherit".
		got, ok = rs.Rule("id")
		require.True(t, ok)
		require.Equal(t, "ID", got)
	})
	t.Run("Replace", func(t *testing.T) {
		rs, err := Initialisms{"FOO", "BAR"}.build()
		require.NoError(t, err)
		require.NotNil(t, rs)
		for _, want := range []string{"FOO", "BAR"} {
			got, ok := rs.Rule(want)
			require.True(t, ok)
			require.Equal(t, want, got)
		}
		// Without "inherit", the default "id" rule is gone.
		_, ok := rs.Rule("id")
		require.False(t, ok)
	})
	t.Run("LaterEntriesOverride", func(t *testing.T) {
		// "inherit" brings id->ID; a later "Id" overrides it.
		rs, err := Initialisms{InitialismsInherit, "Id"}.build()
		require.NoError(t, err)
		got, ok := rs.Rule("id")
		require.True(t, ok)
		require.Equal(t, "Id", got)
	})
	t.Run("Invalid", func(t *testing.T) {
		// Empty, separators, and non-ASCII runes are rejected: the latter could
		// never match ASCII-only word parts, so we fail loudly instead of
		// silently ignoring them.
		for _, bad := range []string{"", "FOO BAR", "foo-bar", "Café", "Ä"} {
			_, err := Initialisms{bad}.build()
			require.Error(t, err, "%q should be rejected", bad)
		}
	})
}

// TestInitialismsCustomE2E ensures custom initialisms reach generated struct
// field names end to end.
func TestInitialismsCustomE2E(t *testing.T) {
	objSchema := &ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{Name: "fqdn", Schema: &ogen.Schema{Type: "string"}},
			{Name: "serverFqdn", Schema: &ogen.Schema{Type: "string"}},
		},
	}
	spec := &ogen.Spec{
		OpenAPI: "3.0.3",
		Info:    ogen.Info{Title: "test", Version: "1.0.0"},
		Paths: ogen.Paths{
			"/obj": &ogen.PathItem{
				Get: &ogen.Operation{
					OperationID: "getObj",
					Responses: ogen.Responses{
						"200": &ogen.Response{
							Description: "ok",
							Content: map[string]ogen.Media{
								"application/json": {Schema: objSchema},
							},
						},
					},
				},
			},
		},
	}

	fieldNames := func(t *testing.T, opts Options) map[string]struct{} {
		g, err := NewGenerator(spec, opts)
		require.NoError(t, err)
		fields := map[string]struct{}{}
		for _, typ := range g.tstorage.types {
			for _, f := range typ.Fields {
				fields[f.Name] = struct{}{}
			}
		}
		return fields
	}

	custom := Initialisms{InitialismsInherit, "FQDN"}

	// Without customization, "fqdn" is not an initialism.
	builtin := fieldNames(t, Options{})
	require.Contains(t, builtin, "Fqdn")
	require.Contains(t, builtin, "ServerFqdn")

	// A custom initialism applies to whole word parts (the standalone "fqdn"
	// property), but the camelCase "serverFqdn" is not split without the feature,
	// so its "Fqdn" sub-word stays untouched.
	customNoFeature := fieldNames(t, Options{
		Generator: GenerateOptions{Initialisms: custom},
	})
	require.Contains(t, customNoFeature, "FQDN")
	require.Contains(t, customNoFeature, "ServerFqdn")

	// With the NamingInitialisms feature, the camelCase token is split and the
	// custom initialism matches the sub-word too.
	customWithFeature := fieldNames(t, Options{
		Generator: GenerateOptions{
			Features:    &FeatureOptions{Enable: FeatureSet{NamingInitialisms.Name: {}}},
			Initialisms: custom,
		},
	})
	require.Contains(t, customWithFeature, "FQDN")
	require.Contains(t, customWithFeature, "ServerFQDN")
}
