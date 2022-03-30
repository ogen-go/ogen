// Package jsonpointer contains RFC 6901 JSON Pointer implementation.
package jsonpointer

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// Resolve takes given pointer and returns byte slice of requested value if any.
// If value not found, returns NotFoundError.
func Resolve(ptr string, buf []byte) ([]byte, error) {
	switch {
	case len(ptr) == 0 || ptr == "#":
		return validate(buf)
	case ptr[0] == '/':
		return find(ptr, buf)
	case ptr[0] == '#': // Note that length is bigger than 1.
		unescaped, err := url.PathUnescape(ptr[1:])
		if err != nil {
			return nil, errors.Wrap(err, "unescape")
		}
		// Fast-path to not parse URL.
		return find(unescaped, buf)
	}

	u, err := url.Parse(ptr)
	if err != nil {
		return nil, err
	}
	return find(u.Fragment, buf)
}

func validate(buf []byte) ([]byte, error) {
	if err := jx.DecodeBytes(buf).Validate(); err != nil {
		return nil, errors.Wrap(err, "validate")
	}
	return buf, nil
}

func find(ptr string, buf []byte) ([]byte, error) {
	d := jx.GetDecoder()
	defer jx.PutDecoder(d)

	if len(ptr) == 0 {
		return validate(buf)
	}

	if ptr[0] != '/' {
		return nil, errors.Errorf("invalid pointer %q: pointer must start with '/'", ptr)
	}
	ptr = ptr[1:]

	err := splitFunc(ptr, '/', func(part string) (err error) {
		part = unescape(part)
		var (
			result []byte
			ok     bool
		)
		d.ResetBytes(buf)
		switch tt := d.Next(); tt {
		case jx.Object:
			result, ok, err = findKey(d, part)
			if err != nil {
				return errors.Wrapf(err, "find key %q", part)
			}
		case jx.Array:
			result, ok, err = findIdx(d, part)
			if err != nil {
				return errors.Wrapf(err, "find index %q", part)
			}
		default:
			return errors.Errorf("unexpected type %q", tt)
		}
		if !ok {
			return &NotFoundError{Pointer: ptr}
		}

		buf = result
		return err
	})
	return buf, err
}

func findIdx(d *jx.Decoder, part string) (result []byte, ok bool, _ error) {
	index, err := strconv.ParseUint(part, 10, 64)
	if err != nil {
		return nil, false, errors.Wrap(err, "index")
	}

	counter := uint64(0)
	if err := d.Arr(func(d *jx.Decoder) error {
		defer func() {
			counter++
		}()

		if index == counter {
			raw, err := d.Raw()
			if err != nil {
				return errors.Wrapf(err, "parse %d", counter)
			}
			result = raw
			ok = true
			return nil
		}
		return d.Skip()
	}); err != nil {
		return nil, false, err
	}
	return result, ok, nil
}

func findKey(d *jx.Decoder, part string) (result []byte, ok bool, _ error) {
	if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case part:
			raw, err := d.Raw()
			if err != nil {
				return errors.Wrapf(err, "parse %q", key)
			}
			result = raw
			ok = true
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return nil, false, err
	}
	return result, true, nil
}

var (
	unescapeReplacer = strings.NewReplacer(
		"~1", "/",
		"~0", "~",
	)
)

func unescape(part string) string {
	return unescapeReplacer.Replace(part)
}
