package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
)

// tstorage is a type storage.
type tstorage struct {
	refs      map[string]*ir.Type           // Key: ref
	types     map[string]*ir.Type           // Key: type name
	responses map[string]*ir.StatusResponse // Key: ref

	// w means 'wrapped'.
	// These maps stores references to types and responses
	// that we wrapped in a [T]StatusCode struct.
	wtypes     map[string]*ir.Type           // Key: ref
	wresponses map[string]*ir.StatusResponse // Key: ref
}

func newTStorage() *tstorage {
	return &tstorage{
		refs:       map[string]*ir.Type{},
		types:      map[string]*ir.Type{},
		responses:  map[string]*ir.StatusResponse{},
		wtypes:     map[string]*ir.Type{},
		wresponses: map[string]*ir.StatusResponse{},
	}
}

func (s *tstorage) saveType(t *ir.Type) error {
	if !t.Is(ir.KindInterface, ir.KindStruct, ir.KindMap, ir.KindEnum, ir.KindAlias, ir.KindGeneric, ir.KindSum, ir.KindStream) {
		panic(unreachable(t))
	}

	if confT, ok := s.types[t.Name]; ok {
		if t.IsGeneric() {
			// HACK:
			// Currently generator can overwrite same generic type
			// multiple times during IR generation.
			//
			// We need to keep the set of features and methods consistent
			// during this overwrites...
			//
			// Maybe we should instantiate generic types only once when needed
			// and reuse them?
			for _, feature := range confT.Features {
				t.AddFeature(feature)
			}
			for iface := range confT.Implements {
				t.Implement(iface)
			}
		} else {
			return errors.Errorf("schema name conflict: %q", t.Name)
		}
	}

	s.types[t.Name] = t
	return nil
}

func (s *tstorage) saveRef(ref string, t *ir.Type) error {
	if _, ok := s.refs[ref]; ok {
		return errors.Errorf("reference conflict: %q", ref)
	}
	if _, ok := s.types[t.Name]; ok {
		return errors.Errorf("reference %q type name conflict: %q", ref, t.Name)
	}

	s.refs[ref] = t
	s.types[t.Name] = t
	return nil
}

func (s *tstorage) saveResponse(ref string, r *ir.StatusResponse) error {
	if _, ok := s.responses[ref]; ok {
		return errors.Errorf("reference conflict: %q", ref)
	}

	s.responses[ref] = r
	return nil
}

func (s *tstorage) saveWType(ref string, t *ir.Type) error {
	if _, ok := s.wtypes[ref]; ok {
		return errors.Errorf("reference conflict: %q", ref)
	}
	if _, ok := s.types[t.Name]; ok {
		return errors.Errorf("reference %q type name conflict: %q", ref, t.Name)
	}

	s.wtypes[ref] = t
	s.types[t.Name] = t
	return nil
}

func (s *tstorage) saveWResponse(ref string, r *ir.StatusResponse) error {
	if _, ok := s.wresponses[ref]; ok {
		return errors.Errorf("reference conflict: %q", ref)
	}

	s.wresponses[ref] = r
	return nil
}

func (s *tstorage) merge(other *tstorage) error {
	// Check for merge conflicts.
	for ref, t := range other.refs {
		if _, ok := s.refs[ref]; ok {
			return errors.Errorf("reference conflict: %q", ref)
		}
		if _, ok := s.types[t.Name]; ok {
			return errors.Errorf("reference type %q name conflict: %q", ref, t.Name)
		}
	}

	for name, t := range other.types {
		if confT, ok := s.types[name]; ok {
			if t.IsGeneric() {
				for _, feature := range confT.Features {
					t.AddFeature(feature)
				}
				for iface := range confT.Implements {
					t.Implement(iface)
				}
			} else {
				return errors.Errorf("anonymous type name conflict: %q", name)
			}
		}
	}

	for ref := range other.responses {
		if _, ok := s.responses[ref]; ok {
			return errors.Errorf("response reference conflict: %q", ref)
		}
	}

	for ref := range other.wresponses {
		if _, ok := s.wresponses[ref]; ok {
			return errors.Errorf("wrapped response reference conflict: %q", ref)
		}
	}

	for ref := range other.wtypes {
		if _, ok := s.wtypes[ref]; ok {
			return errors.Errorf("wrapped type reference conflict: %q", ref)
		}
	}

	// Merge types.
	for ref, t := range other.refs {
		s.refs[ref] = t
		s.types[t.Name] = t
	}

	for name, t := range other.types {
		s.types[name] = t
	}

	for name, t := range other.responses {
		s.responses[name] = t
	}

	for name, t := range other.wresponses {
		s.wresponses[name] = t
	}

	for name, t := range other.wtypes {
		s.wtypes[name] = t
	}

	return nil
}
