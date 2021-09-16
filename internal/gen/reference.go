package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
)

// componentsParameter searches parameter defined in components section.
func (g *Generator) componentsParameter(ref string) (ogen.Parameter, bool) {
	if !strings.HasPrefix(ref, "#/components/parameters/") {
		return ogen.Parameter{}, false
	}

	targetName := strings.TrimPrefix(ref, "#/components/parameters/")
	for name, param := range g.spec.Components.Parameters {
		if name == targetName && param.Ref == "" {
			return param, true
		}
	}

	return ogen.Parameter{}, false
}

func componentName(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/components/") {
		return "", fmt.Errorf("bad reference: '%s'", ref)
	}

	s := strings.TrimPrefix(ref, "#/components/")
	switch {
	case strings.HasPrefix(s, "schemas/"):
		return strings.TrimPrefix(s, "schemas/"), nil
	case strings.HasPrefix(s, "requestBodies/"):
		return strings.TrimPrefix(s, "requestBodies/"), nil
	case strings.HasPrefix(s, "responses/"):
		return strings.TrimPrefix(s, "responses/"), nil
	default:
		return "", fmt.Errorf("bad reference: '%s'", ref)
	}
}
