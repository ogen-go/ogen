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

func explodePtr(v bool) *bool { return &v }

func stringArraySchema() *ogen.Schema {
	return &ogen.Schema{
		Type:  "array",
		Items: &ogen.Items{Item: &ogen.Schema{Type: "string"}},
	}
}

func TestParameterExplodeDefault(t *testing.T) {
	object := &ogen.Schema{
		Type:       "object",
		Properties: []ogen.Property{{Name: "id", Schema: &ogen.Schema{Type: "string"}}},
	}

	tests := []struct {
		name    string
		style   string
		explode *bool
		schema  *ogen.Schema
		want    bool
	}{
		{"Unset", "", nil, stringArraySchema(), true},
		{"Form", "form", nil, stringArraySchema(), true},
		{"SpaceDelimited", "spaceDelimited", nil, stringArraySchema(), false},
		{"PipeDelimited", "pipeDelimited", nil, stringArraySchema(), false},
		{"DeepObject", "deepObject", nil, object, true},
		{"SpaceDelimitedExplodeTrue", "spaceDelimited", explodePtr(true), stringArraySchema(), true},
		{"PipeDelimitedExplodeTrue", "pipeDelimited", explodePtr(true), stringArraySchema(), true},
		{"FormExplodeFalse", "form", explodePtr(false), stringArraySchema(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &ogen.Spec{
				OpenAPI: "3.0.3",
				Paths: map[string]*ogen.PathItem{
					"/items": {
						Get: &ogen.Operation{
							OperationID: "listItems",
							Parameters: []*ogen.Parameter{
								{
									Name:    "ids",
									In:      "query",
									Style:   tt.style,
									Explode: tt.explode,
									Schema:  tt.schema,
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

			spec, err := Parse(root, Settings{
				RootURL: testRootURL,
			})
			a.NoError(err)
			a.Len(spec.Operations, 1)

			params := spec.Operations[0].Parameters
			a.Len(params, 1)
			a.Equal(tt.want, params[0].Explode)
		})
	}
}

func TestEncodingExplodeDefault(t *testing.T) {
	tests := []struct {
		name    string
		style   string
		explode *bool
		want    bool
	}{
		{"Unset", "", nil, true},
		{"Form", "form", nil, true},
		{"SpaceDelimited", "spaceDelimited", nil, false},
		{"PipeDelimited", "pipeDelimited", nil, false},
		{"PipeDelimitedExplodeTrue", "pipeDelimited", explodePtr(true), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const ct = "application/x-www-form-urlencoded"

			root := &ogen.Spec{
				OpenAPI: "3.0.3",
				Paths: map[string]*ogen.PathItem{
					"/submit": {
						Post: &ogen.Operation{
							OperationID: "submit",
							RequestBody: &ogen.RequestBody{
								Required: true,
								Content: map[string]ogen.Media{
									ct: {
										Schema: &ogen.Schema{
											Type:       "object",
											Properties: []ogen.Property{{Name: "tags", Schema: stringArraySchema()}},
										},
										Encoding: map[string]ogen.Encoding{
											"tags": {Style: tt.style, Explode: tt.explode},
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

			encoding := spec.Operations[0].RequestBody.Content[ct].Encoding
			a.Contains(encoding, "tags")
			a.Equal(tt.want, encoding["tags"].Explode)
		})
	}
}

func TestHeaderExplodeDefault(t *testing.T) {
	tests := []struct {
		name    string
		explode *bool
		want    bool
	}{
		{"Unset", nil, false},
		{"ExplodeTrue", explodePtr(true), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &ogen.Spec{
				OpenAPI: "3.0.3",
				Paths: map[string]*ogen.PathItem{
					"/items": {
						Get: &ogen.Operation{
							OperationID: "listItems",
							Responses: map[string]*ogen.Response{
								"200": {
									Description: "OK",
									Headers: map[string]*ogen.Header{
										"X-Tags": {
											Explode: tt.explode,
											Schema:  stringArraySchema(),
										},
									},
								},
							},
						},
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

			headers := spec.Operations[0].Responses.StatusCode[200].Headers
			a.Contains(headers, "X-Tags")
			a.Equal(tt.want, headers["X-Tags"].Explode)
		})
	}
}
