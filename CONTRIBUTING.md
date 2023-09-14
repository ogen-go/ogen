# Contributing

This project follows [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

Before creating pull requests, please read the [coding guidelines](https://github.com/uber-go/guide/blob/master/style.md) and
follow some existing [pull requests](https://github.com/ogen-go/ogen/pulls).

## Optimizations

Please provide [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) output if your PR
tries to optimize something.

## Committing generated code

If you are contributing to the project and make changes to the code generator, please commit the
generated code as well. This is to make sure that the generated code is always up-to-date.

To update generated code run:

```console
$ make generate examples
```

Generated code should be committed in a one separate commit `chore: commit generated files`.

```console
$ git add ./examples ./internal/integration/*/*_gen*.go
$ git commit -m "chore: commit generated files"
```

## Coding guidance

Please read [Uber code style](https://github.com/uber-go/guide/blob/master/style.md).
