package gen

import "fmt"

func (g *Generator) generateComponents() error {
	if err := func() error {
		if err := g.generateComponentSchemas(); err != nil {
			return fmt.Errorf("schemas: %w", err)
		}
		if err := g.generateComponentRequestBodies(); err != nil {
			return fmt.Errorf("requestBodies: %w", err)
		}
		if err := g.generateComponentResponses(); err != nil {
			return fmt.Errorf("responses: %w", err)
		}
		return nil
	}(); err != nil {
		return fmt.Errorf("components: %w", err)
	}

	return nil
}

func (g *Generator) generateComponentSchemas() error {
	for name, schema := range g.spec.Components.Schemas {
		s, err := g.generateSchema(name, schema)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		g.schemas[name] = s
	}

	return nil
}

func (g *Generator) generateComponentRequestBodies() error {
	for name, body := range g.spec.Components.RequestBodies {
		rbody, err := g.generateRequestBody(name, &body)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		g.requestBodies[name] = rbody
	}

	return nil
}

func (g *Generator) generateComponentResponses() error {
	for name, resp := range g.spec.Components.Responses {
		r, err := g.generateResponse(name, resp)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		g.responses[name] = r
	}

	return nil
}
