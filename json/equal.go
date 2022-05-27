package json

import (
	"bytes"
	"math/big"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

type compare struct {
	left, right *jx.Decoder
}

func (c compare) equalBool() (bool, error) {
	lval, err := c.left.Bool()
	if err != nil {
		return false, errors.Wrap(err, "left")
	}
	rval, err := c.right.Bool()
	if err != nil {
		return false, errors.Wrap(err, "right")
	}
	return lval == rval, nil
}

func (c compare) equalString() (bool, error) {
	lval, err := c.left.StrBytes()
	if err != nil {
		return false, errors.Wrap(err, "left")
	}
	rval, err := c.right.StrBytes()
	if err != nil {
		return false, errors.Wrap(err, "right")
	}
	return bytes.Equal(lval, rval), nil
}

func (c compare) equalNumber() (bool, error) {
	lval, err := c.left.Num()
	if err != nil {
		return false, errors.Wrap(err, "left")
	}
	rval, err := c.right.Num()
	if err != nil {
		return false, errors.Wrap(err, "right")
	}

	// Fast path comparing.
	switch {
	case lval.Zero() != rval.Zero():
		return false, nil
	case lval.Equal(rval): // Compare like byte slices.
		return true, nil
	}
	// TODO(tdakkota): try compare as int64/float64 if value fits into it.

	lnum, rnum := new(big.Rat), new(big.Rat)
	if err := lnum.UnmarshalText(lval); err != nil {
		return false, errors.Wrap(err, "left")
	}
	if err := rnum.UnmarshalText(rval); err != nil {
		return false, errors.Wrap(err, "right")
	}
	return lnum.Cmp(rnum) == 0, nil
}

func (c compare) equalArray() (bool, error) {
	liter, err := c.left.ArrIter()
	if err != nil {
		return false, errors.Wrap(err, "left")
	}
	riter, err := c.right.ArrIter()
	if err != nil {
		return false, errors.Wrap(err, "right")
	}

	i := 0
	for liter.Next() {
		// Left array is bigger than right.
		if !riter.Next() {
			return false, nil
		}

		ok, err := c.equal()
		if err != nil {
			return false, errors.Wrapf(err, "[%d]", i)
		}
		if !ok {
			return false, nil
		}
		i++
	}

	if err := liter.Err(); err != nil {
		return false, errors.Wrap(err, "left")
	}
	if err := riter.Err(); err != nil {
		return false, errors.Wrap(err, "right")
	}

	// Right array is bigger than left.
	return !riter.Next(), nil
}

func (c compare) equalObject() (bool, error) {
	// TODO(tdakkota): is there a more efficient way?
	collectObject := func(d *jx.Decoder) (m map[string]jx.Raw, err error) {
		m = map[string]jx.Raw{}
		err = d.ObjBytes(func(d *jx.Decoder, key []byte) error {
			raw, err := d.Raw()
			if err != nil {
				return errors.Wrapf(err, "%q", key)
			}
			m[string(key)] = raw
			return nil
		})
		return m, err
	}

	lmap, err := collectObject(c.left)
	if err != nil {
		return false, errors.Wrap(err, "left")
	}
	rmap, err := collectObject(c.right)
	if err != nil {
		return false, errors.Wrap(err, "right")
	}

	if len(lmap) != len(rmap) {
		return false, nil
	}

	for name, lvalue := range lmap {
		rvalue, ok := rmap[name]
		if !ok {
			return false, nil
		}
		// Raw already validated.
		if ok, _ := Equal(rvalue, lvalue); !ok {
			return false, nil
		}
	}
	return true, nil
}

func (c compare) equal() (bool, error) {
	lt, rt := c.left.Next(), c.right.Next()
	switch {
	case lt == jx.Invalid:
		return false, errors.Wrap(c.left.Validate(), "left")
	case rt == jx.Invalid:
		return false, errors.Wrap(c.right.Validate(), "right")
	case lt != rt:
		return false, nil
	}

	switch lt {
	case jx.Null:
		// lt is equal to rt, so values are equal.
		return true, nil
	case jx.Bool:
		return c.equalBool()
	case jx.String:
		return c.equalString()
	case jx.Number:
		return c.equalNumber()
	case jx.Array:
		return c.equalArray()
	case jx.Object:
		return c.equalObject()
	default:
		panic("unreachable")
	}
}

// Equal compares two JSON values.
func Equal(a, b []byte) (bool, error) {
	c := compare{
		left: jx.DecodeBytes(a), right: jx.DecodeBytes(b),
	}
	return c.equal()
}
