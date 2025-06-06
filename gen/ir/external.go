package ir

import (
	"go/types"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/tools/go/packages"
)

// ExternalType represents an external type.
type ExternalType struct {
	PackagePath string
	GoName      string
	ImportAlias string
	Encode      ExternalEncoding
	Decode      ExternalEncoding
}

// ExternalEncoding is a kind of external type.
type ExternalEncoding int

// String returns the string representation of the ExternalEncoding.
func (e ExternalEncoding) String() string {
	var s strings.Builder
	for i := ExternalEncoding(1); i <= e; i <<= 1 {
		if e.Has(i) {
			if s.Len() > 0 {
				s.WriteString("|")
			}
			switch i {
			case ExternalJSON:
				s.WriteString("JSON")
			case ExternalText:
				s.WriteString("Text")
			case ExternalBinary:
				s.WriteString("Binary")
			}
		}
	}
	return s.String()
}

// Has checks if the encoding kind is present in the ExternalEncoding.
func (e ExternalEncoding) Has(kind ExternalEncoding) bool {
	return e&kind != 0
}

const (
	// ExternalNative indicates that the type implements ogen's json.Marshaler or json.Unmarshaler.
	ExternalNative ExternalEncoding = 1 << iota
	// ExternalJSON indicates that the type implements stdlib json.Marshaler or json.Unmarshaler.
	ExternalJSON
	// ExternalText indicates that the type implements stdlib encoding.TextMarshaler or encoding.TextUnmarshaler.
	ExternalText
	// ExternalBinary indicates that the type implements stdlib encoding.BinaryMarshaler or encoding.BinaryUnmarshaler.
	ExternalBinary
)

var encoders = map[[2]string]ExternalEncoding{
	{"github.com/ogen-go/ogen/json", "Marshaler"}: ExternalNative,
	{"encoding/json", "Marshaler"}:                ExternalJSON,
	{"encoding", "TextMarshaler"}:                 ExternalText,
	{"encoding", "BinaryMarshaler"}:               ExternalBinary,
}

var decoders = map[[2]string]ExternalEncoding{
	{"github.com/ogen-go/ogen/json", "Unmarshaler"}: ExternalNative,
	{"encoding/json", "Unmarshaler"}:                ExternalJSON,
	{"encoding", "TextUnmarshaler"}:                 ExternalText,
	{"encoding", "BinaryUnmarshaler"}:               ExternalBinary,
}

func getExternalEncoding(pkgPath, typeName string) (encode, decode ExternalEncoding, _ error) {
	pkgPaths := []string{pkgPath}
	for _, m := range [2]map[[2]string]ExternalEncoding{encoders, decoders} {
		for k := range m {
			pkgPaths = append(pkgPaths, k[0])
		}
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes}, pkgPaths...)
	if err != nil || len(pkgs) == 0 {
		return -1, -1, errors.Wrap(err, "failed to load packages")
	}

	getType := func(pkgs []*packages.Package, pkgPath, typeName string) *types.Named {
		for _, pkg := range pkgs {
			if pkg.Types == nil || pkg.Types.Scope() == nil {
				continue
			}
			obj := pkg.Types.Scope().Lookup(typeName)
			if obj != nil && obj.Pkg().Path() == pkgPath {
				if named, ok := obj.Type().(*types.Named); ok {
					return named
				}
			}
		}
		return nil
	}

	getIface := func(pkgs []*packages.Package, pkgPath, typeName string) *types.Interface {
		return getType(pkgs, pkgPath, typeName).Underlying().(*types.Interface).Complete()
	}

	typ := getType(pkgs, pkgPath, typeName)
	if typ == nil {
		return -1, -1, errors.Errorf("type not found: %s.%s", pkgPath, typeName)
	}
	ptr := types.NewPointer(typ)

	for name, kind := range encoders {
		iface := getIface(pkgs, name[0], name[1])
		if types.Implements(typ, iface) || types.Implements(ptr, iface) {
			encode |= kind
		}
	}
	for name, kind := range decoders {
		iface := getIface(pkgs, name[0], name[1])
		if types.Implements(typ, iface) || types.Implements(ptr, iface) {
			decode |= kind
		}
	}

	return encode, decode, nil
}
