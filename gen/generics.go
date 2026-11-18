package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/naming"
)

// checkStructRecursions breaks recursive cycles by boxing recursive
// optional or nullable fields of s into pointers.
//
// It must be called in two passes over all struct types: first with
// final=false, which only boxes optional and nullable fields, then with
// final=true, which reports required fields that are still recursive.
// Otherwise a required field may be reported as infinitely recursive before
// the nullable field that breaks the same cycle has been boxed.
func checkStructRecursions(s *ir.Type, final bool) error {
	for _, field := range s.Fields {
		if field.Spec == nil {
			continue
		}

		// Apply the same transformations as type generation, so that
		// nullable unions like anyOf: [X, {type: 'null'}] are seen as nullable.
		schema := transformSchema(field.Spec.Schema)
		v := ir.GenericVariant{
			Nullable: schema != nil && schema.Nullable,
			Optional: !field.Spec.Required,
		}
		if !final && !v.Any() {
			continue
		}

		boxedT, err := func(t *ir.Type) (*ir.Type, error) {
			if s.RecursiveTo(t) {
				if t.IsGeneric() {
					t = t.GenericOf
				}

				switch {
				case v.OnlyOptional():
					return ir.Pointer(t, ir.NilOptional), nil
				case v.OnlyNullable():
					return ir.Pointer(t, ir.NilNull), nil
				case v.NullableOptional():
					t, err := boxType(t, ir.GenericVariant{
						Optional: true,
					})
					if err != nil {
						return nil, err
					}
					return ir.Pointer(t, ir.NilNull), nil
				default:
					// Required.
					return nil, errors.Errorf("infinite recursion: %s.%s is required", s.Name, field.Name)
				}
			}
			return t, nil
		}(field.Type)
		if err != nil {
			return errors.Wrapf(err, "wrap field %q with generic type", field.Name)
		}

		field.Type = boxedT
	}

	return nil
}

func boxType(t *ir.Type, v ir.GenericVariant) (*ir.Type, error) {
	dealiased := t
	if dealiased.IsAlias() {
		dealiased = dealiased.AliasTo
	}

	// Do not wrap if
	//  * type is Any
	//  * type is Stream
	//  * type is not nullable and not optional
	if dealiased.IsAny() || dealiased.IsStream() || !v.Any() {
		return t, nil
	}
	// Do not wrap if type is Null primitive and generic is nullable only.
	if dealiased.IsNull() {
		if v.OnlyNullable() {
			return t, nil
		}
		v.Nullable = false
	}

	if dealiased.IsArray() || dealiased.Primitive == ir.ByteSlice {
		// Using special case for array nil value if possible.
		switch {
		case v.OnlyOptional():
			t.NilSemantic = ir.NilOptional
		case v.OnlyNullable():
			t.NilSemantic = ir.NilNull
		default:
			postfix, err := genericPostfix(t)
			if err != nil {
				return nil, errors.Wrap(err, "postfix")
			}
			return ir.Generic(postfix, t, v), nil
		}
		dealiased.NilSemantic = t.NilSemantic

		return t, nil
	}

	if t.CanGeneric() {
		postfix, err := genericPostfix(t)
		if err != nil {
			return nil, errors.Wrap(err, "postfix")
		}
		return ir.Generic(postfix, t, v), nil
	}

	switch {
	case v.OnlyOptional():
		return t.Pointer(ir.NilOptional), nil
	case v.OnlyNullable():
		return t.Pointer(ir.NilNull), nil
	default:
		postfix, err := genericPostfix(t)
		if err != nil {
			return nil, errors.Wrap(err, "postfix")
		}
		return ir.Generic(postfix,
			t.Pointer(ir.NilNull), ir.GenericVariant{Optional: true},
		), nil
	}
}

func genericPostfix(t *ir.Type) (string, error) {
	name := naming.AfterDot(t.NamePostfix())
	return pascal(name)
}
