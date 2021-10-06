// Package json implements next-generation json.
//
// In pre-generic era and relies mostly on codegen.
// MUST be rewritten when go1.18rc1 is out.
package json

import (
	json "github.com/json-iterator/go"
	"golang.org/x/xerrors"
)

// Unmarshaler implements json reading.
type Unmarshaler interface {
	ReadJSON(i *json.Iterator) error
}

// Marshaler implements json writing.
type Marshaler interface {
	WriteFieldJSON(k string, s *json.Stream) error
	WriteJSON(s *json.Stream) error
}

// Value represents a json value.
type Value interface {
	Marshaler
	Unmarshaler
}

// OptionalString is Optional[string].
type OptionalString struct {
	Value string
	Set   bool
}

// ReadJSON implements Unmarshaler.
func (o *OptionalString) ReadJSON(i *json.Iterator) error {
	o.Value = i.ReadString()
	return i.Error
}

// WriteFieldJSON implements Marshaler.
func (o OptionalString) WriteFieldJSON(k string, s *json.Stream) error {
	if !o.Set {
		return nil
	}
	s.WriteObjectField(k)
	return o.WriteJSON(s)
}

func (o OptionalString) WriteJSON(s *json.Stream) error {
	s.WriteString(o.Value)
	return nil
}

// Get optional string.
func (o OptionalString) Get() (v string, ok bool) {
	return o.Value, o.Set
}

// GetDefault is Get with default value
func (o OptionalString) GetDefault(defaultValue string) string {
	if o.Set {
		return o.Value
	}
	return defaultValue
}

// NullableString is string that is either null or defined.
type NullableString struct {
	Value string
	Nil   bool
}

// OptionalNullableString is combined Optional[Nullable[string]].
//
// Value can be one of those:
//	* undefined
//	* ""
//	* nil
//	* "some value"
type OptionalNullableString struct {
	NullableString
	Set bool
}

func (o OptionalNullableString) WriteFieldJSON(k string, s *json.Stream) error {
	if !o.Set {
		return nil
	}
	s.WriteObjectField(k)
	return o.WriteJSON(s)
}

func (o OptionalNullableString) WriteJSON(s *json.Stream) error {
	if o.Nil {
		s.WriteNil()
	} else {
		s.WriteString(o.Value)
	}
	return s.Error
}

func (o *OptionalNullableString) ReadJSON(i *json.Iterator) error {
	o.Value = ""
	o.Set = false
	o.Nil = false

	switch t := i.WhatIsNext(); t {
	case json.StringValue:
		o.Set = true
		o.Value = i.ReadString()
		return i.Error
	case json.NilValue:
		o.Set = true
		o.Nil = true
		i.Skip()
		return i.Error
	default:
		return xerrors.Errorf("unexpected type %v", t)
	}
}
