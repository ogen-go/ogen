# ogen

WIP Opinionated OpenAPI v3 Code Generator for Go

On early stages of development.

Telegram group for development: [@ogen_dev](https://t.me/ogen_dev)

# Install
```console
go get github.com/ogen-go/ogen
```

# Draft Roadmap

* [Generated client](https://github.com/ogen-go/ogen/issues/8)
* Tests
* Enums
* Convenient global errors schema (e.g. 500, 404)
* End-to-end tests
* Security (e.g. Bearer token)
* Framework/Router support
  * stdlib
  * gin
  * echo
  * fasthttp
* Middlewares, logging (e.g. how to pass request id)
* RED metrics for client and server
* Tracing for client and server
* Basic validation
* OneOf/AnyOf
* Client retries
  * Retry strategy (e.g. exponential backoff)
  * Configuring via `x-ogen-*` annotations
  * Configuring via generation config
* Benchmarks
* Full validation support
* Extreme optimizations
  * fasthttp
  * total zero alloc
    * memory pools for entities with automatic management in generated code
    * [gnet](https://github.com/panjf2000/gnet) support
