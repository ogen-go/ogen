package ast

type PathPart struct {
	Raw   string
	Param *Parameter
}

func (m *Method) Path() string {
	var path string
	for _, part := range m.PathParts {
		if part.Raw != "" {
			path += "/" + part.Raw
			continue
		}

		path += "/{" + part.Param.SourceName + "}"
	}
	return path
}
