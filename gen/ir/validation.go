package ir

import (
	"fmt"
	"maps"
	"math/big"
	"slices"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/shopspring/decimal"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/ogenregex"
	"github.com/ogen-go/ogen/validate"
)

type Validators struct {
	String  validate.String
	Int     validate.Int
	Float   validate.Float
	Decimal validate.Decimal
	Array   validate.Array
	Object  validate.Object
	// Ogen contains parameters for custom validation.
	Ogen map[string]any
}

func (v *Validators) SetString(schema *jsonschema.Schema) (err error) {
	if schema.Pattern != "" {
		v.String.Regex, err = ogenregex.Compile(schema.Pattern)
		if err != nil {
			return errors.Wrap(err, "pattern")
		}
	}
	if schema.MaxLength != nil {
		v.String.SetMaxLength(int(*schema.MaxLength))
	}
	if schema.MinLength != nil {
		v.String.SetMinLength(int(*schema.MinLength))
	}
	if schema.Format == "email" {
		v.String.Email = true
	}
	if schema.Format == "hostname" {
		v.String.Hostname = true
	}
	if schema.Format == "byte" {
		v.String.Byte = true
	}
	return nil
}

func (v *Validators) SetInt(schema *jsonschema.Schema) error {
	if num := jx.Num(schema.MultipleOf); len(num) > 0 {
		val, err := num.Uint64()
		if err != nil {
			return errors.Wrap(err, "set multipleOf")
		}
		v.Int.SetMultipleOf(val)
	}
	set := func(num jx.Num, f func(int64)) error {
		if len(num) < 1 {
			return nil
		}
		val, err := num.Int64()
		if err != nil {
			return err
		}
		f(val)
		return nil
	}
	if err := set(jx.Num(schema.Maximum), v.Int.SetMaximum); err != nil {
		return errors.Wrap(err, "set maximum")
	}
	if err := set(jx.Num(schema.Minimum), v.Int.SetMinimum); err != nil {
		return errors.Wrap(err, "set minimum")
	}
	v.Int.MaxExclusive = schema.ExclusiveMaximum
	v.Int.MinExclusive = schema.ExclusiveMinimum
	return nil
}

func (v *Validators) SetFloat(schema *jsonschema.Schema) error {
	if num := jx.Num(schema.MultipleOf); len(num) > 0 {
		n := new(big.Rat)
		if err := n.UnmarshalText(num); err != nil {
			return errors.Wrap(err, "parse multipleOf")
		}
		v.Float.SetMultipleOf(n)
	}
	set := func(num jx.Num, f func(float64)) error {
		if len(num) < 1 {
			return nil
		}
		val, err := num.Float64()
		if err != nil {
			return err
		}
		f(val)
		return nil
	}
	if err := set(jx.Num(schema.Maximum), v.Float.SetMaximum); err != nil {
		return errors.Wrap(err, "set maximum")
	}
	if err := set(jx.Num(schema.Minimum), v.Float.SetMinimum); err != nil {
		return errors.Wrap(err, "set minimum")
	}
	v.Float.MaxExclusive = schema.ExclusiveMaximum
	v.Float.MinExclusive = schema.ExclusiveMinimum
	return nil
}

func (v *Validators) SetDecimal(schema *jsonschema.Schema) error {
	if num := jx.Num(schema.MultipleOf); len(num) > 0 {
		n, err := decimal.NewFromString(string(num))
		if err != nil {
			return errors.Wrap(err, "parse multipleOf")
		}
		v.Decimal.SetMultipleOf(n)
	}
	set := func(num jx.Num, f func(decimal.Decimal)) error {
		if len(num) == 0 {
			return nil
		}
		val, err := decimal.NewFromString(string(num))
		if err != nil {
			return err
		}
		f(val)
		return nil
	}
	if err := set(jx.Num(schema.Maximum), v.Decimal.SetMaximum); err != nil {
		return errors.Wrap(err, "set maximum")
	}
	if err := set(jx.Num(schema.Minimum), v.Decimal.SetMinimum); err != nil {
		return errors.Wrap(err, "set minimum")
	}
	v.Decimal.MaxExclusive = schema.ExclusiveMaximum
	v.Decimal.MinExclusive = schema.ExclusiveMinimum
	return nil
}

func (v *Validators) SetArray(schema *jsonschema.Schema) {
	if schema.MaxItems != nil {
		v.Array.SetMaxLength(int(*schema.MaxItems))
	}
	if schema.MinItems != nil {
		v.Array.SetMinLength(int(*schema.MinItems))
	}
	if schema.UniqueItems {
		v.Array.SetUniqueItems(true)
	}
}

func (v *Validators) SetObject(schema *jsonschema.Schema) {
	if schema.MaxProperties != nil {
		v.Object.SetMaxProperties(int(*schema.MaxProperties))
	}
	if schema.MinProperties != nil {
		v.Object.SetMinProperties(int(*schema.MinProperties))
	}
	if schema.MinLength != nil {
		v.Object.SetMinLength(int(*schema.MinLength))
	}
	if schema.MaxLength != nil {
		v.Object.SetMaxLength(int(*schema.MaxLength))
	}
}

func (t *Type) NeedValidation() bool {
	return t.needValidation(&walkpath{})
}

func (v *Validators) SetOgenValidate(schema *jsonschema.Schema) {
	if len(schema.OgenValidate) == 0 {
		return
	}
	if v.Ogen == nil {
		v.Ogen = make(map[string]any, len(schema.OgenValidate))
	}
	maps.Copy(v.Ogen, schema.OgenValidate)
}

func (t *Type) needValidation(path *walkpath) (result bool) {
	if t == nil {
		return false
	}

	if path.has(t) {
		return false
	}
	path.add(t)
	defer path.delete(t)

	switch t.Kind {
	case KindPrimitive:
		if t.IsFloat() {
			// NaN, Inf, float validators.
			return true
		}
		if t.IsNumeric() {
			return t.Validators.Int.Set() || t.Validators.Decimal.Set()
		}
		if t.Validators.String.Set() {
			switch t.Primitive {
			case String, ByteSlice:
				return true
			}
		}
		if len(t.Validators.Ogen) > 0 {
			return true
		}
		return false
	case KindEnum:
		return true
	case KindSum:
		return slices.ContainsFunc(t.SumOf, func(s *Type) bool {
			return s.needValidation(path)
		})
	case KindAlias:
		return t.AliasTo.needValidation(path)
	case KindPointer:
		if t.NilSemantic == NilInvalid {
			return true
		}
		return t.PointerTo.needValidation(path)
	case KindGeneric:
		return t.GenericOf.needValidation(path)
	case KindArray:
		if t.NilSemantic == NilInvalid {
			return true
		}
		if t.Validators.Array.Set() {
			return true
		}
		if len(t.Validators.Ogen) > 0 {
			return true
		}
		return t.Item.needValidation(path)
	case KindStruct:
		if len(t.Validators.Ogen) > 0 {
			return true
		}
		return slices.ContainsFunc(t.Fields, func(f *Field) bool {
			return f.Type.needValidation(path)
		})
	case KindMap:
		if t.Validators.Object.Set() {
			return true
		}
		if len(t.Validators.Ogen) > 0 {
			return true
		}
		return t.Item.needValidation(path)
	case KindStream, KindInterface, KindAny:
		// FIXME(tdakkota): try to validate Any.
		return false
	default:
		panic(fmt.Sprintf("unreachable: %s", t.Kind))
	}
}
