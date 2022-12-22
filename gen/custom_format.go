package gen

import (
	"fmt"
	"go/token"
	"reflect"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
)

func checkImportableType(typ reflect.Type) error {
	path := typ.PkgPath()
	if path == "main" {
		return errors.New("type must be in importable package")
	}

	name := typ.Name()
	if name == "" {
		return errors.New("type must be named or primitive")
	}

	if path != "" && !token.IsExported(name) {
		return errors.New("type must be exported")
	}

	return nil
}

func (g *Generator) makeCustomFormats() error {
	importPaths := map[string]string{}

	makeExternal := func(typ reflect.Type) (ir.ExternalType, error) {
		if err := checkImportableType(typ); err != nil {
			return ir.ExternalType{}, err
		}

		path := typ.PkgPath()
		if path == "" {
			// Primitive type.
			return ir.ExternalType{Type: typ}, nil
		}

		importName, ok := importPaths[path]
		if !ok {
			importName = fmt.Sprintf("custom%d", len(importPaths))
			importPaths[path] = importName
			g.imports = append(g.imports, fmt.Sprintf("%s %q", importName, path))
		}

		return ir.ExternalType{
			Pkg:  importName,
			Type: typ,
		}, nil
	}

	for _, jsonTyp := range xmaps.SortedKeys(g.opt.CustomFormats) {
		formats := g.opt.CustomFormats[jsonTyp]
		for _, format := range xmaps.SortedKeys(formats) {
			def := formats[format]

			if _, ok := g.customFormats[jsonTyp]; !ok {
				g.customFormats[jsonTyp] = map[string]ir.CustomFormat{}
			}

			f, err := func() (f ir.CustomFormat, _ error) {
				goName, err := pascalNonEmpty(format)
				if err != nil {
					return f, errors.Wrap(err, "generate go name")
				}

				typ, err := makeExternal(def.typ)
				if err != nil {
					return f, errors.Wrap(err, "format type")
				}

				json, err := makeExternal(def.json)
				if err != nil {
					return f, errors.Wrap(err, "json encoding")
				}

				text, err := makeExternal(def.text)
				if err != nil {
					return f, errors.Wrap(err, "text encoding")
				}

				return ir.CustomFormat{
					Name:   format,
					GoName: goName,
					Type:   typ,
					JSON:   json,
					Text:   text,
				}, nil
			}()
			if err != nil {
				return errors.Wrapf(err, "custom format %q:%q", jsonTyp, format)
			}

			g.customFormats[jsonTyp][format] = f
		}
	}

	return nil
}
