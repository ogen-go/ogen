// Code generated by ogen, DO NOT EDIT.

package api

import (
	"github.com/google/uuid"

	ht "github.com/ogen-go/ogen/http"
)

// NewOptInt returns new OptInt with value set to v.
func NewOptInt(v int) OptInt {
	return OptInt{
		Value: v,
		Set:   true,
	}
}

// OptInt is optional int.
type OptInt struct {
	Value int
	Set   bool
}

// IsSet returns true if OptInt was set.
func (o OptInt) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptInt) Reset() {
	var v int
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptInt) SetTo(v int) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptInt) Get() (v int, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptInt) Or(d int) int {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptMultipartFile returns new OptMultipartFile with value set to v.
func NewOptMultipartFile(v ht.MultipartFile) OptMultipartFile {
	return OptMultipartFile{
		Value: v,
		Set:   true,
	}
}

// OptMultipartFile is optional ht.MultipartFile.
type OptMultipartFile struct {
	Value ht.MultipartFile
	Set   bool
}

// IsSet returns true if OptMultipartFile was set.
func (o OptMultipartFile) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptMultipartFile) Reset() {
	var v ht.MultipartFile
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptMultipartFile) SetTo(v ht.MultipartFile) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptMultipartFile) Get() (v ht.MultipartFile, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptMultipartFile) Or(d ht.MultipartFile) ht.MultipartFile {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptString returns new OptString with value set to v.
func NewOptString(v string) OptString {
	return OptString{
		Value: v,
		Set:   true,
	}
}

// OptString is optional string.
type OptString struct {
	Value string
	Set   bool
}

// IsSet returns true if OptString was set.
func (o OptString) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptString) Reset() {
	var v string
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptString) SetTo(v string) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptString) Get() (v string, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptString) Or(d string) string {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptTestFormDeepObject returns new OptTestFormDeepObject with value set to v.
func NewOptTestFormDeepObject(v TestFormDeepObject) OptTestFormDeepObject {
	return OptTestFormDeepObject{
		Value: v,
		Set:   true,
	}
}

// OptTestFormDeepObject is optional TestFormDeepObject.
type OptTestFormDeepObject struct {
	Value TestFormDeepObject
	Set   bool
}

// IsSet returns true if OptTestFormDeepObject was set.
func (o OptTestFormDeepObject) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptTestFormDeepObject) Reset() {
	var v TestFormDeepObject
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptTestFormDeepObject) SetTo(v TestFormDeepObject) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptTestFormDeepObject) Get() (v TestFormDeepObject, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptTestFormDeepObject) Or(d TestFormDeepObject) TestFormDeepObject {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptTestFormObject returns new OptTestFormObject with value set to v.
func NewOptTestFormObject(v TestFormObject) OptTestFormObject {
	return OptTestFormObject{
		Value: v,
		Set:   true,
	}
}

// OptTestFormObject is optional TestFormObject.
type OptTestFormObject struct {
	Value TestFormObject
	Set   bool
}

// IsSet returns true if OptTestFormObject was set.
func (o OptTestFormObject) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptTestFormObject) Reset() {
	var v TestFormObject
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptTestFormObject) SetTo(v TestFormObject) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptTestFormObject) Get() (v TestFormObject, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptTestFormObject) Or(d TestFormObject) TestFormObject {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptUUID returns new OptUUID with value set to v.
func NewOptUUID(v uuid.UUID) OptUUID {
	return OptUUID{
		Value: v,
		Set:   true,
	}
}

// OptUUID is optional uuid.UUID.
type OptUUID struct {
	Value uuid.UUID
	Set   bool
}

// IsSet returns true if OptUUID was set.
func (o OptUUID) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptUUID) Reset() {
	var v uuid.UUID
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptUUID) SetTo(v uuid.UUID) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptUUID) Get() (v uuid.UUID, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptUUID) Or(d uuid.UUID) uuid.UUID {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// Ref: #/components/schemas/SharedRequest
type SharedRequest struct {
	Filename OptString "json:\"filename\""
	File     OptString "json:\"file\""
}

func (*SharedRequest) testShareFormSchemaReq() {}

// Ref: #/components/schemas/SharedRequest
type SharedRequestForm struct {
	Filename OptString        "json:\"filename\""
	File     OptMultipartFile "json:\"file\""
}

func (*SharedRequestForm) testShareFormSchemaReq() {}

// Ref: #/components/schemas/TestForm
type TestForm struct {
	ID          OptInt                "json:\"id\""
	UUID        OptUUID               "json:\"uuid\""
	Description string                "json:\"description\""
	Array       []string              "json:\"array\""
	Object      OptTestFormObject     "json:\"object\""
	DeepObject  OptTestFormDeepObject "json:\"deepObject\""
}

type TestFormDeepObject struct {
	Min OptInt "json:\"min\""
	Max int    "json:\"max\""
}

type TestFormObject struct {
	Min OptInt "json:\"min\""
	Max int    "json:\"max\""
}

// TestFormURLEncodedOK is response for TestFormURLEncoded operation.
type TestFormURLEncodedOK struct{}

// TestMultipartOK is response for TestMultipart operation.
type TestMultipartOK struct{}

type TestMultipartUploadOK struct {
	File         string    "json:\"file\""
	OptionalFile OptString "json:\"optional_file\""
	Files        []string  "json:\"files\""
}

type TestMultipartUploadReq struct {
	OrderId      OptInt    "json:\"orderId\""
	UserId       OptInt    "json:\"userId\""
	File         string    "json:\"file\""
	OptionalFile OptString "json:\"optional_file\""
	Files        []string  "json:\"files\""
}

type TestMultipartUploadReqForm struct {
	OrderId      OptInt             "json:\"orderId\""
	UserId       OptInt             "json:\"userId\""
	File         ht.MultipartFile   "json:\"file\""
	OptionalFile OptMultipartFile   "json:\"optional_file\""
	Files        []ht.MultipartFile "json:\"files\""
}

// TestShareFormSchemaOK is response for TestShareFormSchema operation.
type TestShareFormSchemaOK struct{}
