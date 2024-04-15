<p align="center">
<img width="256" height="256" src="_logo/logo.svg" alt="ogen svg logo">
</p>

# ogen [![Go Reference](https://img.shields.io/badge/go-pkg-00ADD8)](https://pkg.go.dev/github.com/ogen-go/ogen#section-documentation) [![codecov](https://img.shields.io/codecov/c/github/ogen-go/ogen?label=cover)](https://codecov.io/gh/ogen-go/ogen) [![stable](https://img.shields.io/badge/-stable-brightgreen)](https://go-faster.org/docs/projects/status#stable)

OpenAPI v3 Code Generator for Go.

- [Getting started](https://ogen.dev/docs/intro)
- [Sample project](https://github.com/ogen-go/example)
- [Security policy](https://github.com/ogen-go/ogen/blob/-/SECURITY.md)
- [Telegram group `@ogen_dev`](https://t.me/ogen_dev)

# Install

```console
go get -d github.com/ogen-go/ogen
```

# Usage

```go
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target target/dir -package api --clean schema.json
```

or using container:
```shell
docker run --rm \
  --volume ".:/workspace" \
  ghcr.io/ogen-go/ogen:latest --target workspace/petstore --clean workspace/petstore.yml
```

# Features

- No reflection or `interface{}`
  - The json encoding is code-generated, optimized and uses [go-faster/jx](https://github.com/go-faster/jx) for speed and overcoming `encoding/json` limitations
  - Validation is code-generated according to spec
- Code-generated static radix router
- No more boilerplate
  - Structures are generated from OpenAPI v3 specification
  - Arguments, headers, url queries are parsed according to specification into structures
  - String formats like `uuid`, `date`, `date-time`, `uri` are represented by go types directly
- Statically typed client and server
- Convenient support for optional, nullable and optional nullable fields
  - No more pointers
  - Generated Optional[T], Nullable[T] or OptionalNullable[T] wrappers with helpers
  - Special case for array handling with `nil` semantics relevant to specification
    - When array is optional, `nil` denotes absence of value
    - When nullable, `nil` denotes that value is `nil`
    - When required, `nil` currently the same as `[]`, but is actually invalid
    - If both nullable and required, wrapper will be generated (TODO)
- Generated sum types for oneOf
  - Primitive types (`string`, `number`) are detected by type
  - Discriminator field is used if defined in schema
  - Type is inferred by unique fields if possible
- Extra Go struct field tags in the generated types
- OpenTelemetry tracing and metrics

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

## Extension properties

OpenAPI enables [Specification Extensions](https://spec.openapis.org/oas/v3.1.0#specification-extensions),
which are implemented as patterned fields that are always prefixed by `x-`.

### Server name

Optionally, server name can be specified by `x-ogen-server-name`, for example:

```json
{
  "openapi": "3.0.3",
  "servers": [
    {
      "x-ogen-server-name": "production",
      "url": "https://{region}.example.com/{val}/v1",
    },
    {
      "x-ogen-server-name": "prefix",
      "url": "/{val}/v1",
    },
    {
      "x-ogen-server-name": "const",
      "url": "https://cdn.example.com/v1"
    }
  ],
(...)
```

### Custom type name

Optionally, type name can be specified by `x-ogen-name`, for example:

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "x-ogen-name": "Name",
  "properties": {
    "foobar": {
      "$ref": "#/$defs/FooBar"
    }
  },
  "$defs": {
    "FooBar": {
      "x-ogen-name": "FooBar",
      "type": "object",
      "properties": {
        "foo": {
          "type": "string"
        }
      }
    }
  }
}
```

### Custom field name

Optionally, type name can be specified by `x-ogen-properties`, for example:

```yaml
components:
  schemas:
    Node:
      type: object
      properties:
        parent:
          $ref: "#/components/schemas/Node"
        child:
          $ref: "#/components/schemas/Node"
      x-ogen-properties:
        parent:
          name: "Prev"
        child:
          name: "Next"
```

The generated source code looks like:

```go
// Ref: #/components/schemas/Node
type Node struct {
    Prev *Node `json:"parent"`
    Next *Node `json:"child"`
}
```

### Extra struct field tags

Optionally, additional Go struct field tags can be specified by `x-oapi-codegen-extra-tags`, for example:

```yaml
components:
  schemas:
    Pet:
      type: object
      required:
        - id
      properties:
        id:
          type: integer
          format: int64
          x-oapi-codegen-extra-tags:
            gorm: primaryKey
            valid: customIdValidator
```

The generated source code looks like:

```go
// Ref: #/components/schemas/Pet
type Pet struct {
    ID   int64     `gorm:"primaryKey" valid:"customNameValidator" json:"id"`
}
```

### Streaming JSON encoding

By default, ogen loads the entire JSON body into memory before decoding it.
Optionally, streaming JSON encoding can be enabled by `x-ogen-json-streaming`, for example:

```yaml
requestBody:
  required: true
  content:
    application/json:
      x-ogen-json-streaming: true
      schema:
        type: array
        items:
          type: number
```

### Operation groups

Optionally, operations can be grouped so a handler interface will be generated for each group of operations. 
This is useful for organizing operations for large APIs.

The group for operations on a path or individual operations can be specified by `x-ogen-operation-group`, for example:

```yaml
paths:
  /images:
    x-ogen-operation-group: Images
    get:
      operationId: listImages
      ...
  /images/{imageID}:
    x-ogen-operation-group: Images
    get:
      operationId: getImageByID
      ...
  /users:
    x-ogen-operation-group: Users
    get:
      operationId: listUsers
      ...
```

The generated handler interfaces look like this:

```go
// x-ogen-operation-group: Images
type ImagesHandler interface {
    ListImages(ctx context.Context, req *ListImagesRequest) (*ListImagesResponse, error)
    GetImageByID(ctx context.Context, req *GetImagesByIDRequest) (*GetImagesByIDResponse, error)
}

// x-ogen-operation-group: Users
type UsersHandler interface {
    ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
}

type Handler interface {
    ImagesHandler
    UsersHandler
    // All un-grouped operations will be on this interface
}
```

## JSON

Code generation provides very efficient and flexible encoding and decoding of json:

```go
// Decode decodes Error from json.
func (s *Error) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Error to nil")
	}
	return d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "code":
			if err := func() error {
				v, err := d.Int64()
				s.Code = int64(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"code\"")
			}
		case "message":
			if err := func() error {
				v, err := d.Str()
				s.Message = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"message\"")
			}
		default:
			return d.Skip()
		}
		return nil
	})
}
```

# Links

- [Getting started](https://ogen.dev/docs/intro)
- [Sample project](https://github.com/ogen-go/example)
- [Security policy](https://github.com/ogen-go/ogen/blob/-/SECURITY.md)
- [Telegram chat `@ogen_dev`](https://t.me/ogen_dev)
- [Roadmap](https://github.com/ogen-go/ogen/blob/-/ROADMAP.md)
