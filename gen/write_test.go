package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/ogenregex"
	"github.com/ogen-go/ogen/validate"
)

// TestRegexStrings_CollectsAllPatternTypes is a regression test for ensuring
// that RegexStrings() collects regex patterns from all validator types:
// - String.Regex
// - Int.Pattern
// - Float.Pattern
// - MapPattern
//
// Previously, Int.Pattern and Float.Pattern were not collected, causing
// generated code to reference undefined regexMap entries when numeric types
// had pattern constraints (added in lenient-validation-mode feature).
func TestRegexStrings_CollectsAllPatternTypes(t *testing.T) {
	stringRegex := ogenregex.MustCompile("^string-pattern$")
	intPattern := ogenregex.MustCompile("^\\d+$")
	floatPattern := ogenregex.MustCompile("^\\d+\\.\\d+$")
	mapPattern := ogenregex.MustCompile("^key-.*$")

	config := TemplateConfig{
		Types: map[string]*ir.Type{
			"StringType": {
				Name: "StringType",
				Validators: ir.Validators{
					String: validate.String{
						Regex: stringRegex,
					},
				},
			},
			"IntType": {
				Name: "IntType",
				Validators: ir.Validators{
					Int: validate.Int{
						Pattern: intPattern,
					},
				},
			},
			"FloatType": {
				Name: "FloatType",
				Validators: ir.Validators{
					Float: validate.Float{
						Pattern: floatPattern,
					},
				},
			},
			"MapType": {
				Name:       "MapType",
				MapPattern: mapPattern,
			},
		},
	}

	regexStrings := config.RegexStrings()

	// Verify all four pattern types are collected
	expectedPatterns := []string{
		"^string-pattern$",
		"^\\d+$",
		"^\\d+\\.\\d+$",
		"^key-.*$",
	}

	require.ElementsMatch(t, expectedPatterns, regexStrings,
		"RegexStrings() must collect patterns from String.Regex, Int.Pattern, Float.Pattern, and MapPattern")
}

// TestRegexStrings_HandlesNilPatterns verifies that nil patterns are handled correctly
func TestRegexStrings_HandlesNilPatterns(t *testing.T) {
	stringRegex := ogenregex.MustCompile("^test$")

	config := TemplateConfig{
		Types: map[string]*ir.Type{
			"StringType": {
				Name: "StringType",
				Validators: ir.Validators{
					String: validate.String{
						Regex: stringRegex,
					},
				},
			},
			"IntType": {
				Name: "IntType",
				Validators: ir.Validators{
					Int: validate.Int{
						Pattern: nil, // nil pattern should be skipped
					},
				},
			},
		},
	}

	regexStrings := config.RegexStrings()

	require.Equal(t, []string{"^test$"}, regexStrings,
		"RegexStrings() should skip nil patterns and only return non-nil ones")
}

// TestRegexStrings_DeduplicatesPatterns verifies that duplicate patterns are deduplicated
func TestRegexStrings_DeduplicatesPatterns(t *testing.T) {
	samePattern := ogenregex.MustCompile("^shared-pattern$")

	config := TemplateConfig{
		Types: map[string]*ir.Type{
			"StringType1": {
				Name: "StringType1",
				Validators: ir.Validators{
					String: validate.String{
						Regex: samePattern,
					},
				},
			},
			"StringType2": {
				Name: "StringType2",
				Validators: ir.Validators{
					String: validate.String{
						Regex: samePattern,
					},
				},
			},
		},
	}

	regexStrings := config.RegexStrings()

	require.Equal(t, []string{"^shared-pattern$"}, regexStrings,
		"RegexStrings() should deduplicate identical patterns")
}

func TestGenerator_WriteSource_ProblemJSONErrorType(t *testing.T) {
	g := &Generator{
		opt: GenerateOptions{
			Features: &FeatureOptions{
				DisableAll: true,
			},
		},
		tstorage: newTStorage(),
		errType: &ir.Response{
			Contents: map[ir.ContentType]ir.Media{
				ir.ContentType("application/problem+json"): {
					Encoding: ir.EncodingProblemJSON,
					Type: &ir.Type{
						Kind: ir.KindStruct,
						Name: "ErrorResponse",
					},
				},
			},
		},
		imports: defaultImports(),
	}

	require.NoError(t, g.WriteSource(genfs.CheckFS{}, "api"))
}
