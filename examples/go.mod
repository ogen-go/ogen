module github.com/ogen-go/ogen/examples

go 1.17

require (
	github.com/go-chi/chi/v5 v5.0.4
	github.com/ogen-go/ogen v0.0.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

require (
	github.com/bytedance/sonic v1.0.0-beta.0.20210924085059-00716d86349c // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20210823082418-56861234f7ea // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
)

replace github.com/ogen-go/ogen v0.0.0 => ./..
