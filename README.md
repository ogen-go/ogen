<p align="center">
<img width="256" height="256" src="_logo/logo.svg" alt="ogen svg logo">
</p>

# ogen [![Go Reference](https://img.shields.io/badge/go-pkg-00ADD8)](https://pkg.go.dev/github.com/ogen-go/ogen#section-documentation) [![codecov](https://img.shields.io/codecov/c/github/ogen-go/ogen?label=cover)](https://codecov.io/gh/ogen-go/ogen) [![openapi v3](https://img.shields.io/badge/OAS%203-brightgreen)](https://swagger.io/specification/) [![experimental](https://img.shields.io/badge/-unstable-blueviolet)](https://go-faster.org/docs/projects/status#experimental)

Opinionated OpenAPI v3 Code Generator for Go.

[Getting started](https://ogen.dev/docs/intro).

Work is still in progress, so currently no backward compatibility is provided. However, we are close to alpha.

Telegram group for development: [@ogen_dev](https://t.me/ogen_dev)

# Install
```console
go get -d github.com/ogen-go/ogen
```

# Usage
```go
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target target/dir -package api --clean schema.json
```

# Features

* No reflection or `interface{}`
  * The json encoding is code-generated, optimized and uses [go-faster/jx](https://github.com/go-faster/jx) for speed and overcoming `encoding/json` limitations
  * Validation is code-generated according to spec
* Code-generated static radix router
* No more boilerplate
  * Structures are generated from OpenAPI v3 specification
  * Arguments, headers, url queries are parsed according to specification into structures
  * String formats like `uuid`, `date`, `date-time`, `uri` are represented by go types directly
* Statically typed client and server
* Convenient support for optional, nullable and optional nullable fields
  * No more pointers
  * Generated Optional[T], Nullable[T] or OptionalNullable[T] wrappers with helpers
  * Special case for array handling with `nil` semantics relevant to specification
    * When array is optional, `nil` denotes absence of value
    * When nullable, `nil` denotes that value is `nil`
    * When required, `nil` currently the same as `[]`, but is actually invalid
    * If both nullable and required, wrapper will be generated (TODO)
* Generated sum types for oneOf
  * Primitive types (`string`, `number`) are detected by type
  * Discriminator field is used if defined in schema
  * Type is inferred by unique fields if possible
* OpenTelemetry tracing and metrics

Example generated structure from schema:
```go
// Pet describes #/components/schemas/Pet.
type Pet struct {
	Birthday     time.Time     `json:"birthday"`
	Friends      []Pet         `json:"friends"`
	ID           int64         `json:"id"`
	IP           net.IP        `json:"ip"`
	IPV4         net.IP        `json:"ip_v4"`
	IPV6         net.IP        `json:"ip_v6"`
	Kind         PetKind       `json:"kind"`
	Name         string        `json:"name"`
	Next         OptData       `json:"next"`
	Nickname     NilString     `json:"nickname"`
	NullStr      OptNilString  `json:"nullStr"`
	Rate         time.Duration `json:"rate"`
	Tag          OptUUID       `json:"tag"`
	TestArray1   [][]string    `json:"testArray1"`
	TestDate     OptTime       `json:"testDate"`
	TestDateTime OptTime       `json:"testDateTime"`
	TestDuration OptDuration   `json:"testDuration"`
	TestFloat1   OptFloat64    `json:"testFloat1"`
	TestInteger1 OptInt        `json:"testInteger1"`
	TestTime     OptTime       `json:"testTime"`
	Type         OptPetType    `json:"type"`
	URI          url.URL       `json:"uri"`
	UniqueID     uuid.UUID     `json:"unique_id"`
}
```

Example generated server interface:
```go
// Server handles operations described by OpenAPI v3 specification.
type Server interface {
	PetGetByName(ctx context.Context, params PetGetByNameParams) (Pet, error)
	// ...
}
```

Example generated client method signature:
```go
type PetGetByNameParams struct {
    Name string
}

// GET /pet/{name}
func (c *Client) PetGetByName(ctx context.Context, params PetGetByNameParams) (res Pet, err error)
```

## Generics
Instead of using pointers, `ogen` generates generic wrappers.

For example, `OptNilString` is `string` that is optional (no value) and can be `null`.
```go
// OptNilString is optional nullable string.
type OptNilString struct {
	Value string
	Set   bool
	Null  bool
}
```

Multiple convenience helper methods and functions are generated, some of them:
```go
func (OptNilString) Get() (v string, ok bool)
func (OptNilString) IsNull() bool
func (OptNilString) IsSet() bool

func NewOptNilString(v string) OptNilString
```

## Recursive types
If `ogen` encounters recursive types that can't be expressed in go, pointers are used as fallback.

## Sum types
For `oneOf` sum-types are generated. `ID` that is one of `[string, integer]` will be represented like that:
```go
type ID struct {
	Type   IDType
	String string
	Int    int
}

// Also, some helpers:
func NewStringID(v string) ID
func NewIntID(v int) ID
```

## JSON

Code generation provides very efficient and flexible encoding and decoding of json:
```go
// ReadJSON reads Error from json stream.
func (s *Error) ReadJSON(r *json.Reader) error {
	if s == nil {
		return fmt.Errorf(`invalid: unable to decode Error to nil`)
	}
	return r.ObjBytes(func(r *json.Reader, k []byte) error {
		switch string(k) {
		case "code":
			v, err := r.Int64()
			s.Code = int64(v)
			if err != nil {
				return err
			}
		case "message":
			v, err := r.Str()
			s.Message = string(v)
			if err != nil {
				return err
			}
		default:
			return r.Skip()
		}
		return nil
	})
}
```

# Roadmap

* Security (e.g. Bearer token)
* Cookie params
* Default value
* Tests for `ip` package
* Convenient global errors schema (e.g. 500, 404)
  * Add convenience for `Error`, not only `ErrorWithCode`
  * Handle case when `ref` is not used, but responses are equal
* Webhook support
* AnyOf
* Full validation support
* Client retries
  * Retry strategy (e.g. exponential backoff)
  * Configuring via `x-ogen-*` annotations
  * Configuring via generation config
* Separate JSON Schema generator
* Tool for OAS validation for ogen compatibility
  * Multiple error reporting with references
    * JSON path
    * Line and column (optional)
* Tool for OAS backward compatibility check
* DSL-based ent-like code-first approach for writing schemas
* Reduce generated code via generics
* Extreme optimizations
  * Code generation for [regex](https://github.com/CAFxX/regexp2go)
  * Streaming/iterator API support
    * Enable via x-ogen-streaming extension
    * Iteration over array or map elements of object
    * Also can fit njson
  * Advanced Code Generation
    * HTTP
      * URI
      * Header
      * Cookie
    * Templating
    * Encoding/Decoding
      * MessagePack
      * ProtoBuff
  * String interning
* Websocket support via extension?
* Async support (Websocket, other protocols)
  * [asyncapi](https://github.com/asyncapi/spec/blob/v2.2.0/spec/asyncapi.md)
* More marshaling protocols support
  * msgpack
  * protobuf
  * [ndjson](https://github.com/ndjson/ndjson-spec), newline-delimited json
  * text/html
* Automatic end-to-end tests support via routing header
  * Header selects specific response variant
  * Code-generated tests with full coverage
* TechEmpower benchmark implementation
* Integrations
  * Tests with [autorest](https://github.com/Azure/autorest)
  * Use testdata from [autorest](https://github.com/Azure/autorest.typescript/tree/main/test/integration/swaggers)
