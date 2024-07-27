package jsonschema

import (
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// Infer returns a JSON Schema that is inferred from the given JSON.
type Infer struct {
	target RawSchema
}

// Target returns the target schema.
func (i Infer) Target() RawSchema {
	return i.target
}

// Apply applies given data to the schema state.
func (i *Infer) Apply(data []byte) error {
	return apply(&i.target, jx.DecodeBytes(data))
}

func applyType(s *RawSchema, tt string) {
	if hasType(s, tt) {
		return
	}
	if len(s.OneOf) > 0 {
		s.OneOf = append(s.OneOf, &RawSchema{Type: tt})
		return
	}

	if s.Type == "" {
		s.Type = tt
		return
	}

	old := *s
	*s = RawSchema{
		OneOf: []*RawSchema{
			&old,
			{Type: tt},
		},
	}
}

func hasType(s *RawSchema, tt string) bool {
	if s.Type == tt {
		return true
	}
	for _, v := range s.OneOf {
		if v.Type == tt {
			return true
		}
	}
	return false
}

func replaceType(s *RawSchema, from, to string) bool {
	if s.Type == from {
		s.Type = to
		return true
	}
	for _, v := range s.OneOf {
		if v.Type == from {
			v.Type = to
			return true
		}
	}
	return false
}

func apply(s *RawSchema, d *jx.Decoder) error {
	switch tt := d.Next(); tt {
	case jx.String:
		applyType(s, "string")
		return d.Skip()
	case jx.Number:
		n, err := d.Num()
		if err != nil {
			return err
		}
		if n.IsInt() && !hasType(s, "number") {
			applyType(s, "integer")
			return nil
		}
		if replaceType(s, "integer", "number") {
			return nil
		}
		applyType(s, "number")
		return nil
	case jx.Null:
		s.Nullable = true
		return d.Skip()
	case jx.Bool:
		applyType(s, "boolean")
		return d.Skip()
	case jx.Array:
		applyType(s, "array")

		i := 0
		return d.Arr(func(d *jx.Decoder) error {
			if s.Items == nil {
				s.Items = new(RawItems)
			}
			if s.Items.Item == nil {
				s.Items.Item = new(RawSchema)
			}
			if err := apply(s.Items.Item, d); err != nil {
				return errors.Wrapf(err, "apply item %d", i)
			}
			i++
			return nil
		})
	case jx.Object:
		applyType(s, "object")

		// Set s.Properties to non-nil slice to mark that it is not first apply.
		firstApply := s.Properties == nil
		if firstApply {
			s.Properties = RawProperties{}
		}

		// Collect existing properties.
		props := map[string]*RawSchema{}
		for _, prop := range s.Properties {
			props[prop.Name] = prop.Schema
		}

		// Collect required properties.
		required := map[string]struct{}{}
		for _, key := range s.Required {
			required[key] = struct{}{}
		}

		this := map[string]struct{}{}
		if err := d.Obj(func(d *jx.Decoder, key string) error {
			this[key] = struct{}{}

			if err := func() error {
				if prop, ok := props[key]; ok {
					return apply(prop, d)
				}

				// If it is the first apply, mark property as required.
				if firstApply {
					required[key] = struct{}{}
				}

				prop := new(RawSchema)
				if err := apply(prop, d); err != nil {
					return err
				}
				s.Properties = append(s.Properties, RawProperty{
					Name:   key,
					Schema: prop,
				})
				return nil
			}(); err != nil {
				return errors.Wrapf(err, "apply property %q", key)
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "collect properties")
		}

		// Delete required properties that are not in this object.
		for key := range required {
			if _, ok := this[key]; !ok {
				delete(required, key)
			}
		}

		// Write required properties.
		s.Required = s.Required[:0]
		for key := range required {
			s.Required = append(s.Required, key)
		}

		// Sort fields to make output deterministic.
		slices.Sort(s.Required)
		slices.SortStableFunc(s.Properties, func(a, b RawProperty) int {
			return strings.Compare(a.Name, b.Name)
		})

		return nil
	default:
		return errors.Errorf("invalid type: %s", tt)
	}
}
