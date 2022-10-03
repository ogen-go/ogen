package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateServer(s openapi.Server) (ir.Server, error) {
	if s.Name == "" {
		return ir.Server{}, errors.New("server name must be non-empty")
	}
	// The server name is passed using ogen extension, so it guaranteed to be
	// valid Go identifier, but we need to make pascal case anyway.
	name, err := pascal(s.Name, "Server")
	if err != nil {
		return ir.Server{}, errors.Wrapf(err, "server name: %q", s.Name)
	}

	var params []ir.ServerParam
	for _, part := range s.Template {
		if !part.IsParam() {
			continue
		}
		if part.Param.Default == "" {
			return ir.Server{}, &ErrNotImplemented{"empty server variable default"}
		}

		v := part.Param
		paramName, err := pascalNonEmpty(v.Name)
		if err != nil {
			return ir.Server{}, errors.Wrapf(err, "server param name: %q", v.Name)
		}
		params = append(params, ir.ServerParam{
			Name: paramName,
			Spec: v,
		})
	}

	return ir.Server{
		Name:   name,
		Params: params,
		Spec:   s,
	}, nil
}
