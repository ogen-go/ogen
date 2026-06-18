package parser

import (
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

// TestNullableAnyOfParameter checks that a parameter whose schema marks
// nullability the OpenAPI 3.1 way — anyOf with a {"type":"null"}
// branch — is accepted. Previously the parameter style validation rejected the
// null branch with `invalid schema.type:style:explode combination ("null":...)`,
// even though the very same schema is accepted in a request body.
//
// Both the $ref branch and the scalar branch are covered, since the bug is in
// the null branch handling, not in the $ref handling. Query and header
// locations are both exercised, since parameter style validation is
// location-sensitive.
func TestNullableAnyOfParameter(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.1.0",
		Paths: map[string]*ogen.PathItem{
			"/desktops": {
				Get: &ogen.Operation{
					OperationID: "listDesktops",
					Parameters: []*ogen.Parameter{
						{
							Name: "image_type",
							In:   "query",
							Schema: &ogen.Schema{
								AnyOf: []*ogen.Schema{
									{Ref: "#/components/schemas/DesktopImageType"},
									{Type: "null"},
								},
							},
						},
						{
							Name: "desktop_id",
							In:   "query",
							Schema: &ogen.Schema{
								AnyOf: []*ogen.Schema{
									{Type: "string"},
									{Type: "null"},
								},
							},
						},
						{
							Name: "X-Trace-Id",
							In:   "header",
							Schema: &ogen.Schema{
								AnyOf: []*ogen.Schema{
									{Type: "string"},
									{Type: "null"},
								},
							},
						},
					},
					Responses: map[string]*ogen.Response{
						"200": {Description: "OK"},
					},
				},
			},
		},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				"DesktopImageType": {Type: "string"},
			},
		},
	}

	a := require.New(t)

	var raw yaml.Node
	a.NoError(raw.Encode(root))
	root.Raw = &raw

	spec, err := Parse(root, Settings{
		RootURL: testRootURL,
	})
	a.NoError(err)
	a.Len(spec.Operations, 1)

	op := spec.Operations[0]
	a.Len(op.Parameters, 3)
	for _, param := range op.Parameters {
		// The null branch must survive parsing as part of the anyOf union;
		// the nullable Go type is produced later, at generation time.
		a.Len(param.Schema.AnyOf, 2)
	}
}

// TestNullOnlyParameterRejected checks that a parameter whose entire schema is
// {"type": "null"} is still rejected at parse time. A standalone null carries no
// serializable shape, so accepting it would only defer the failure to code
// generation as an opaque template error. Only null *branches* of an
// anyOf/oneOf/allOf union are tolerated (see TestNullableAnyOfParameter).
func TestNullOnlyParameterRejected(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.1.0",
		Paths: map[string]*ogen.PathItem{
			"/desktops": {
				Get: &ogen.Operation{
					OperationID: "listDesktops",
					Parameters: []*ogen.Parameter{
						{
							Name:   "image_type",
							In:     "query",
							Schema: &ogen.Schema{Type: "null"},
						},
					},
					Responses: map[string]*ogen.Response{
						"200": {Description: "OK"},
					},
				},
			},
		},
	}

	a := require.New(t)

	var raw yaml.Node
	a.NoError(raw.Encode(root))
	root.Raw = &raw

	_, err := Parse(root, Settings{
		RootURL: testRootURL,
	})
	a.Error(err)
	a.Contains(err.Error(), "combination")
}
