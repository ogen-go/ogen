package gen

import (
	"net/http"

	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) fix() {
	for _, m := range g.methods {
		g.fixEqualResponses(m)
	}
}

func (g *Generator) fixEqualResponses(m *ast.Method) {
	if len(m.Responses.StatusCode) < 2 {
		return
	}

	iface, ok := m.ResponseType.(*ast.Interface)
	if !ok {
		return
	}

	for lstat, lresp := range m.Responses.StatusCode {
		for rstat, rresp := range m.Responses.StatusCode {
			if lstat == rstat {
				continue
			}

			if lresp.NoContent != nil {
				if rresp.NoContent == nil {
					continue
				}

				if lresp.NoContent.Equal(rresp.NoContent) {
					lname := pascal(m.Name, http.StatusText(lstat))
					rname := pascal(m.Name, http.StatusText(rstat))
					la := ast.Alias(lname, lresp.NoContent)
					ra := ast.Alias(rname, rresp.NoContent)
					lresp.NoContent.Unimplement(iface)
					rresp.NoContent.Unimplement(iface)
					la.Implement(iface)
					ra.Implement(iface)
					g.schemas[la.Name] = la
					g.schemas[ra.Name] = ra
					lresp.NoContent = la
					rresp.NoContent = ra
				}

				continue
			}

			for lct, lschema := range lresp.Contents {
				for rct, rschema := range rresp.Contents {
					if lschema.Equal(rschema) {
						lname := pascal(m.Name, lct, http.StatusText(lstat))
						rname := pascal(m.Name, rct, http.StatusText(rstat))
						la := ast.Alias(lname, lschema)
						ra := ast.Alias(rname, rschema)
						lschema.Unimplement(iface)
						rschema.Unimplement(iface)
						la.Implement(iface)
						ra.Implement(iface)
						g.schemas[la.Name] = la
						g.schemas[ra.Name] = ra
						lresp.Contents[lct] = la
						rresp.Contents[rct] = ra
					}
				}
			}
		}
	}

	for _, r := range m.Responses.StatusCode {
		r.Implement(iface)
	}
	if m.Responses.Default != nil {
		m.Responses.Default.Implement(iface)
	}
}
