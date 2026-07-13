.PHONY: build test lint run fmt

build:
	go build -o bin/provenance-check ./cmd/provenance-check

test:
	go test -race -cover ./...

lint:
	gofmt -l .
	go vet ./...

fmt:
	gofmt -w .

run: build
	./bin/provenance-check $(ARGS)
