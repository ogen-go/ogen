<p align="center">
<img width="256" height="256" src="_logo/logo.svg" alt="ogen svg logo">
</p>

# ogen [![Go Reference](https://img.shields.io/badge/go-pkg-00ADD8)](https://pkg.go.dev/github.com/ogen-go/ogen#section-documentation) [![codecov](https://img.shields.io/codecov/c/github/ogen-go/ogen?label=cover)](https://codecov.io/gh/ogen-go/ogen) [![alpha](https://img.shields.io/badge/-alpha-orange)](https://go-faster.org/docs/projects/status#alpha)

Opinionated OpenAPI v3 Code Generator for Go.

- [Getting started](https://ogen.dev/docs/intro)
- [Sample project](https://github.com/ogen-go/example)
- [Security policy](https://github.com/ogen-go/ogen/blob/-/SECURITY.md)
- [Telegram group `@ogen_dev`](https://t.me/ogen_dev)
- [Roadmap](https://github.com/ogen-go/ogen/blob/-/ROADMAP.md)

Work is still in progress, so currently no backward compatibility is provided.

# Features

* RPC-like experience
* No reflection or `interface{}`
* Statically typed client and server
* Strict api schema compliance (minimal possibility of behavior deviation from developer side)
* Rich OpenAPI and JSON Schema feature support (including validation and reference resolving)
* OpenTelemetry tracing and metrics
* Human-friendly error messages
* Large test codebase (including [k8s](examples/ex_k8s/oas_server_gen.go) and [github](examples/ex_github/oas_server_gen.go) api specs)
* [...and more](FEATURES.md)

# Installation

```console
go install github.com/ogen-go/ogen/cmd/ogen@latest
```

Now you will be able to generate the code:
```console
ogen --target target/dir --package api --clean schema.yml
```

For installation using Go Modules visit [ogen.dev](https://ogen.dev/docs/intro) site.

# License

Source code is available under the [Apache License](LICENSE).