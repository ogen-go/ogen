package ir

import (
	"go/types"
	"path"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/tools/go/packages"
)

// ExternalType represents an external type.
type ExternalType struct {
	PackagePath string
	ImportAlias string
	TypeName    string
	Encode      ExternalEncoding
	Decode      ExternalEncoding
	IsPointer   bool
}

// String returns the string representation of the ExternalType.
func (e ExternalType) Primitive() PrimitiveType {
	var sb strings.Builder
	pkg := e.ImportAlias
	if pkg == "" && e.PackagePath != "" {
		pkg = path.Base(e.PackagePath)
	}
	if pkg != "" {
		sb.WriteString(pkg)
		sb.WriteByte('.')
	}
	sb.WriteString(e.TypeName)
	return PrimitiveType(sb.String())
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
			case ExternalNative:
				s.WriteString("Native")
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

func getExternalType(input string) (ExternalType, error) {
	pkgPath, typeName, isPointer, err := parseTypePath(input)
	if err != nil {
		return ExternalType{}, err
	}

	encode, decode, err := getExternalEncoding(pkgPath, typeName)
	if err != nil {
		return ExternalType{}, err
	}

	return ExternalType{
		PackagePath: pkgPath,
		TypeName:    typeName,
		IsPointer:   isPointer,
		Encode:      encode,
		Decode:      decode,
	}, nil
}

func parseTypePath(input string) (pkgPath, typeName string, isPointer bool, _ error) {
	i := 0
	n := len(input)
	runes := []rune(input)

	// Check for pointer prefix
	if i < n && runes[i] == '*' {
		isPointer = true
		i++
	}

	// Parse package path
	if i < n && runes[i] == '(' {
		// Look for matching ')'
		j := i + 1
		start := j
		for j < n && runes[j] != ')' {
			j++
		}
		if j == n {
			return "", "", false, errors.New("unmatched '('")
		}
		pkgPath = string(runes[start:j])
		i = j + 1 // skip ')'
		if i >= n || runes[i] != '.' {
			return "", "", false, errors.New("expected '.' after ')'")
		}
		i++ // skip '.'
		typeName = string(runes[i:])
		return pkgPath, typeName, isPointer, nil
	}

	// No parens, assume last '.' separates package and type
	lastDot := strings.LastIndex(input[i:], ".")
	if lastDot == -1 {
		return "", "", false, errors.New("no '.' found in type path")
	}
	lastDot += i // adjust for initial offset
	pkgPath = input[i:lastDot]
	typeName = input[lastDot+1:]
	return pkgPath, typeName, isPointer, nil
}

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
		return -1, -1, errors.New("type not found")
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
