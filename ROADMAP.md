### OpenAPI

- [ ] Encoding
    - [ ] XML (#169)
    - [ ] MessagePack (?)
    - [ ] Custom (?)
    - [ ] Other (?)
- [ ] Parameters
    - [ ] Complex types
        - [ ] Any type (?)
        - [ ] Sum types
        - [ ] Complex object schemas (e.g `additionalProperties`)
    - [ ] Cookie parameters
- [ ] Request body
    - [ ] `Content-Type` masks
    - [ ] Forms
        - [ ] Complex form schemas (e.g `additionalProperties`)
        - [ ] Multipart
            - [ ] Part headers
            - [ ] `Content-Type` part header (special encoding field)
- [ ] Security
    - [ ] OAuth2
    - [ ] OpenID Connect
    - [ ] HTTP Digest
    - [ ] Other (?)
- [ ] Webhooks
- [ ] Links (?)
- [ ] Documentation
    - [ ] Handle Common Mark in description (#142)

### JSON Schema

- [ ] Complex `anyOf`
- [ ] Default values for `object`
- [ ] ECMA-262 Regex (#419)
- [ ] Enum
    - [ ] `enum` with `format`
    - [ ] Object `enum`
- [ ] Tuples
- [ ] Validation
    - [ ] Validate Any type
    - [ ] `uniqueItems`
    - [ ] `not` (?)
- [ ] `format`
    - [ ] Make validators compliant with spec
    - [ ] Support more formats (?)
- [ ] [Code generation for regex (?)](https://github.com/CAFxX/regexp2go)
- [ ] Support more drafts (?)

### General code generator features

- [ ] User-friendly error reporting
    - [x] Print location
    - [ ] Report multiple errors
- [ ] Optimization
    - [ ] Streaming (?)
- [ ] Use generics
- [ ] Websocket extension (?)
- [ ] [`asyncapi` support (?)](https://github.com/asyncapi/spec/blob/v2.2.0/spec/asyncapi.md)
- [ ] Client retries (?)

### OAS Tooling

- [ ] Lint specification (?)
    - [ ] Lint HTTP compliance of definitions
    - [ ] Handle possible bugs in specification
- [ ] Backward compatibility check
