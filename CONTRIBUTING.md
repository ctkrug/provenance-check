# Contributing

This is a personal side project, but the workflow is standard Go:

```sh
make build   # compile the CLI to bin/
make test    # go test -race -cover ./...
make lint    # gofmt -l . && go vet ./...
make run     # build then run, e.g. make run ARGS="https://github.com/example/repo"
```

CI runs gofmt, `go vet`, golangci-lint, build, and test on every push and pull request — all
four must pass before a change merges. See [`docs/BACKLOG.md`](docs/BACKLOG.md) for the story
list and acceptance criteria driving the current build.
