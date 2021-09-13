package validator

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
)

type validator struct {
	spec *ogen.Spec
}

func Validate(spec *ogen.Spec) error {
	return (&validator{spec}).validate()
}

func (v *validator) validate() error {
	for path, item := range v.spec.Paths {
		if err := v.validatePathItem(item); err != nil {
			return fmt.Errorf("paths: '%s': %w", path, err)
		}
	}

	if err := v.validateComponents(v.spec.Components); err != nil {
		return fmt.Errorf("components: %w", err)
	}

	return nil
}

func (v *validator) validateComponents(c *ogen.Components) error {
	if c == nil {
		return nil
	}

	for name, schema := range c.Schemas {
		if err := v.validateSchema(schema); err != nil {
			return fmt.Errorf("schema '%s': %w", name, err)
		}
	}
	for name, param := range c.Parameters {
		if err := v.validateParameter(param); err != nil {
			return fmt.Errorf("parameter '%s': %w", name, err)
		}
	}
	for name, body := range c.RequestBodies {
		if err := v.validateRequestBody(body); err != nil {
			return fmt.Errorf("requestBody '%s': %w", name, err)
		}
	}
	for name, resp := range c.Responses {
		if err := v.validateResponse(resp); err != nil {
			return fmt.Errorf("response '%s': %w", name, err)
		}
	}
	return nil
}

func (v *validator) validatePathItem(item ogen.PathItem) error {
	if item.Ref != "" {
		return fmt.Errorf("referenced path items not supported yet")
	}

	if err := v.validateOperation(item.Get); err != nil {
		return fmt.Errorf("get: %w", err)
	}
	if err := v.validateOperation(item.Put); err != nil {
		return fmt.Errorf("put: %w", err)
	}
	if err := v.validateOperation(item.Post); err != nil {
		return fmt.Errorf("post: %w", err)
	}
	if err := v.validateOperation(item.Delete); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if err := v.validateOperation(item.Options); err != nil {
		return fmt.Errorf("options: %w", err)
	}
	if err := v.validateOperation(item.Head); err != nil {
		return fmt.Errorf("head: %w", err)
	}
	if err := v.validateOperation(item.Patch); err != nil {
		return fmt.Errorf("patch: %w", err)
	}
	if err := v.validateOperation(item.Trace); err != nil {
		return fmt.Errorf("trace: %w", err)
	}

	for _, param := range item.Parameters {
		if err := v.validateParameter(param); err != nil {
			return fmt.Errorf("parameter '%s': %w", param.Name, err)
		}
	}

	return nil
}

func (v *validator) validateOperation(op *ogen.Operation) error {
	if op == nil {
		return nil
	}

	for _, param := range op.Parameters {
		if err := v.validateParameter(param); err != nil {
			return fmt.Errorf("parameter '%s': %w", param.Name, err)
		}
	}

	if body := op.RequestBody; body != nil {
		if err := v.validateRequestBody(*body); err != nil {
			return fmt.Errorf("requestBody: %w", err)
		}
	}

	for status, resp := range op.Responses {
		if err := v.validateResponse(resp); err != nil {
			return fmt.Errorf("response '%s': %w", status, err)
		}
	}

	return nil
}

func (v *validator) validateParameter(p ogen.Parameter) error {
	if p.Ref != "" {
		return v.validateParameterRef(p.Ref)
	}

	switch p.In {
	case "query", "header", "cookie":
	case "path":
		if !p.Required {
			return fmt.Errorf("for parameters located in 'path' field 'required' should be true")
		}
	default:
		return fmt.Errorf("unexpected 'in' field value: '%s'", p.In)
	}

	// TODO: Make p.Schema optional.
	//
	// if p.Schema != nil && len(p.Content) > 1 {
	// 	rerr = multierr.Append(rerr, fmt.Errorf("parameter MUST contain either a schema property, or a content property, but not both"))
	// 	return
	// }

	if len(p.Content) > 1 {
		return fmt.Errorf("field 'content'  MUST only contain one entry")
	}

	if err := v.validateSchema(p.Schema); err != nil {
		return fmt.Errorf("field 'schema': %s", err)
	}

	return nil
}

func (v *validator) validateRequestBody(r ogen.RequestBody) error {
	if r.Ref != "" {
		return v.validateRequestBodyRef(r.Ref)
	}

	for contentType, media := range r.Content {
		if err := v.validateMedia(media); err != nil {
			return fmt.Errorf("%s: %w", contentType, err)
		}
	}

	return nil
}

func (v *validator) validateResponse(r ogen.Response) error {
	if r.Ref != "" {
		return v.validateResponseRef(r.Ref)
	}

	if r.Description == "" {
		return fmt.Errorf("field 'description' is required")
	}

	for contentType, media := range r.Content {
		if err := v.validateMedia(media); err != nil {
			return fmt.Errorf("%s: %w", contentType, err)
		}
	}

	return nil
}

func (v *validator) validateMedia(m ogen.Media) error {
	return v.validateSchema(m.Schema)
}

func (v *validator) validateSchema(s ogen.Schema) error {
	if s.Ref != "" {
		return nil
	}

	switch s.Type {
	case "string", "number", "integer", "boolean":
		if s.Items != nil {
			return fmt.Errorf("type '%s' cannot contain 'items' field", s.Type)
		}
		if len(s.Properties) > 0 {
			return fmt.Errorf("type '%s' cannot contain properties", s.Type)
		}

	case "object":
		if s.Items != nil {
			return fmt.Errorf("type '%s' cannot contain 'items' field", s.Type)
		}
		for propName, propSchema := range s.Properties {
			if err := v.validateSchema(propSchema); err != nil {
				return fmt.Errorf("property '%s': %s", propName, err)
			}
		}

	case "array":
		if len(s.Properties) > 0 {
			return fmt.Errorf("type '%s' cannot contain properties", s.Type)
		}
		if s.Items == nil {
			return fmt.Errorf("type '%s' must contain 'items' field", s.Type)
		}
		if err := v.validateSchema(*s.Items); err != nil {
			return fmt.Errorf("field 'items': %s", err)
		}
	default:
		return fmt.Errorf("unexpected type: '%s'", s.Type)
	}

	return nil
}

func (v *validator) validateResponseRef(ref string) error {
	if !strings.HasPrefix(ref, "#/components/responses/") {
		return fmt.Errorf("invalid response reference: '%s'", ref)
	}

	targetName := strings.TrimPrefix(ref, "#/components/responses/")
	for name, resp := range v.spec.Components.Responses {
		if name == targetName && resp.Ref == "" {
			return nil
		}
	}

	return fmt.Errorf("referenced response with name '%s' not found in components section", targetName)
}

func (v *validator) validateRequestBodyRef(ref string) error {
	if !strings.HasPrefix(ref, "#/components/requestBodies/") {
		return fmt.Errorf("invalid requestBody reference: '%s'", ref)
	}

	targetName := strings.TrimPrefix(ref, "#/components/requestBodies/")
	for name, body := range v.spec.Components.RequestBodies {
		if name == targetName && body.Ref == "" {
			return nil
		}
	}

	return fmt.Errorf("referenced requestBody with name '%s' not found in components section", targetName)
}

func (v *validator) validateParameterRef(ref string) error {
	if !strings.HasPrefix(ref, "#/components/parameters/") {
		return fmt.Errorf("invalid parameters reference: '%s'", ref)
	}

	targetName := strings.TrimPrefix(ref, "#/components/parameters/")
	for name, param := range v.spec.Components.Parameters {
		if name == targetName && param.Ref == "" {
			return nil
		}
	}

	return fmt.Errorf("referenced parameter with name '%s' not found in components section", targetName)
}
