<p align="center">
<img width="256" height="256" src="_logo/logo.svg" alt="ogen logo">
</p>

# ogen

WIP Opinionated OpenAPI v3 Code Generator for Go

On early stages of development.

Telegram group for development: [@ogen_dev](https://t.me/ogen_dev)

# Install
```console
go get github.com/ogen-go/ogen
```

# Usage
```go
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema schema.json --target target/dir -package api --clean
```

# Features

* No reflection or `interface{}`
  * The json encoding is code-generated, optimized and uses `jsoniter` for speed and overcoming `encoding/json` limitations
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

# Draft Roadmap

* Handle unexpected json keys
* Convenient global errors schema (e.g. 500, 404)
* Security (e.g. Bearer token)
* Separate JSON Schema generator
* Framework/Router support
  * stdlib
  * gin
  * echo
  * fasthttp
* Middlewares, logging (e.g. how to pass request id)
* RED metrics for client and server
* Tracing for client and server
* Basic validation
  * String
    * Regex
* OneOf/AnyOf
* Format
  * [String format](https://json-schema.org/understanding-json-schema/reference/string.html)
    * Email
    * Hostname
    * Regular expression
* Webhook support
* Files support (with streaming, like io.Reader/Writer)
* Client retries
  * Retry strategy (e.g. exponential backoff)
  * Configuring via `x-ogen-*` annotations
  * Configuring via generation config
* Tool for OAS validation for ogen compatibility
  * Multiple error reporting with references
    * JSON path
    * Line and column (optional)
* Tool for OAS backward compatibility check
* DSL-based ent-like code-first approach for writing schemas
* Benchmarks
* Generics
  * Target go1.18
  * Use Optional[T]
  * Reduce generated code via generics
* Full validation support
* Extreme optimizations
  * Code generation for [regex](https://github.com/CAFxX/regexp2go)
  * Streaming/iterator API support
    * Enable via x-ogen-streaming extension
    * Iteration over array or map elements of object
    * Also can fit njson
  * fasthttp
  * total zero alloc
    * memory pools for entities with automatic management in generated code
    * [cloudwego/netpoll](https://github.com/cloudwego/netpoll) support
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
  * [simd](https://github.com/minio/simdjson-go) for json
    * Better for streaming multi-megabyte jsons
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
