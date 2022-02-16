package gen

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
)

func boxStructFields(ctx *genctx, s *ir.Type) error {
	for _, field := range s.Fields {
		if field.Spec == nil {
			continue
		}

		v := ir.GenericVariant{
			Nullable: field.Spec.Schema.Nullable,
			Optional: !field.Spec.Required,
		}

		boxedT, err := func(t *ir.Type) (*ir.Type, error) {
			if s.RecursiveTo(t) {
				switch {
				case v.OnlyOptional():
					return ir.Pointer(t, ir.NilOptional), nil
				case v.OnlyNullable():
					return ir.Pointer(t, ir.NilNull), nil
				case v.NullableOptional():
					t, err := boxType(ctx, ir.GenericVariant{
						Optional: true,
					}, t)
					if err != nil {
						return nil, err
					}

					return ir.Pointer(t, ir.NilNull), nil
				default:
					// Required.
					return nil, errors.Errorf("infinite recursion: %s.%s is required", s.Name, field.Name)
				}
			}
			return boxType(ctx, v, t)
		}(field.Type)
		if err != nil {
			return errors.Wrapf(err, "wrap field %q with generic type", field.Name)
		}

		field.Type = boxedT
	}

	return nil
}

func boxType(ctx *genctx, v ir.GenericVariant, t *ir.Type) (*ir.Type, error) {
	// Do not wrap if
	//  * type is Any
	//  * type is Stream
	//  * type is not nullable and not optional
	if t.IsAny() || t.IsStream() || !v.Any() {
		return t, nil
	}

	if t.IsArray() || t.Primitive == ir.ByteSlice {
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
			t = ir.Generic(postfix, t, v)
			if err := ctx.saveType(t); err != nil {
				return nil, err
			}
		}

		return t, nil
	}

	if t.CanGeneric() {
		postfix, err := genericPostfix(t)
		if err != nil {
			return nil, errors.Wrap(err, "postfix")
		}
		t = ir.Generic(postfix, t, v)
		if err := ctx.saveType(t); err != nil {
			return nil, err
		}

		return t, nil
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
		t = ir.Generic(postfix,
			t.Pointer(ir.NilNull), ir.GenericVariant{Optional: true},
		)
		if err := ctx.saveType(t); err != nil {
			return nil, err
		}

		return t, nil
	}
}

func genericPostfix(t *ir.Type) (string, error) {
	name := t.NamePostfix()
	if idx := strings.Index(name, "."); idx > 0 {
		name = name[idx+1:]
	}
	return pascal(name)
}
