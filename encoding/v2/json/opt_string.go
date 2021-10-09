package json

import json "github.com/json-iterator/go"

// OptionalString is Optional[string].
type OptionalString struct {
	Value string
	Set   bool
}

func (o *OptionalString) SetTo(v string) {
	o.Value = v
	o.Set = true
}

func (o *OptionalString) Unset() {
	o.Set = false
	o.Value = ""
}

func (o OptionalString) IsSet() bool { return o.Set }

// ReadJSON implements Unmarshaler.
func (o *OptionalString) ReadJSON(i *json.Iterator) bool {
	o.Value = i.ReadString()
	return true
}

// WriteFieldJSON implements Marshaler.
func (o OptionalString) WriteFieldJSON(k string, s *json.Stream) {
	if !o.Set {
		return
	}
	s.WriteObjectField(k)
}

func (o OptionalString) WriteJSON(s *json.Stream) { s.WriteString(o.Value) }

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

func (o NullableString) IsNil() bool { return o.Nil }

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

func (o OptionalNullableString) IsNil() bool { return o.Nil }
func (o OptionalNullableString) IsSet() bool { return o.Set }

func (o OptionalNullableString) WriteFieldJSON(k string, s *json.Stream) {
	if !o.Set {
		return
	}
	s.WriteObjectField(k)
	o.WriteJSON(s)
}

func (o OptionalNullableString) WriteJSON(s *json.Stream) {
	if o.Nil {
		s.WriteNil()
	} else {
		s.WriteString(o.Value)
	}
}

func (o *OptionalNullableString) ReadJSON(i *json.Iterator) bool {
	o.Value = ""
	o.Set = false
	o.Nil = false

	switch t := i.WhatIsNext(); t {
	case json.StringValue:
		o.Set = true
		o.Value = i.ReadString()
		return true
	case json.NilValue:
		o.Set = true
		o.Nil = true
		i.Skip()
		return true
	default:
		i.ReportError("Read", "unexpected type")
		return false
	}
}
