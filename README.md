# ogen ![logo](_logo/logo.svg)

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

# Draft Roadmap

* Enums
* Optionals
* Convenient global errors schema (e.g. 500, 404)
* End-to-end tests
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
    * Length
    * Regex
* OneOf/AnyOf
* Format
  * [String format](https://json-schema.org/understanding-json-schema/reference/string.html)
    * Date
    * DateTime
    * Time
    * Duration
    * IPv4/IPv6
    * Email
    * UUID
    * URI
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
  * [simd](https://github.com/minio/simdjson-go) for json
    * Better for streaming multi-megabyte jsons
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
      * JSON
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
* Logo
  * OpenAPI-like logo with Go colours
* Integrations
  * Tests with [autorest](https://github.com/Azure/autorest)
  * Use testdata from [autorest](https://github.com/Azure/autorest.typescript/tree/main/test/integration/swaggers)
