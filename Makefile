test:
	@./go.test.sh
.PHONY: test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate ./... && go mod tidy
.PHONY: generate

schemas:
	cd schemas && go generate
.PHONY: schemas

# Update the vendored Swagger UI, e.g. make update-swaggerui VERSION=5.33.2
update-swaggerui:
	go run ./tools/swaggerui-update -version $(VERSION)
.PHONY: update-swaggerui

examples:
	cd examples && go generate && go mod tidy
.PHONY: examples

test_examples:
	cd examples && go test ./...

test_fast:
	go test ./...

logo:
	inkscape -z -w 512 -h 512 _logo/logo.svg -e _logo/logo.x512.png
	inkscape -z -w 256 -h 256 _logo/logo.svg -e _logo/logo.x256.png

tidy:
	go mod tidy

tidy_examples:
	cd examples && go mod tidy

tidy_all: tidy tidy_examples

lint:
	@go version
	@golangci-lint version
	golangci-lint run -v -c .golangci.yml ./...

clean: tidy_all generate examples

commit_gen:
	git add ./examples ./internal/integration/*/*_gen*.go
	git commit -m "chore: commit generated files"
