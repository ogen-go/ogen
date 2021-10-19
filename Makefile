test:
	@./go.test.sh
.PHONY: test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate ./...
.PHONY: generate

examples:
	@cd examples && go generate
.PHONY: examples

logo:
	inkscape -z -w 512 -h 512 _logo/logo.svg -e _logo/logo.x512.png
	inkscape -z -w 256 -h 256 _logo/logo.svg -e _logo/logo.x256.png
