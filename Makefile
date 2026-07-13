.PHONY: build test lint run fmt wasm site

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

# wasm compiles the classification engine for the browser. cmd/wasm is
# js/wasm-only (see cmd/wasm/main.go's build tag), so it's built here
# rather than folded into `build`, which targets the native CLI.
wasm:
	GOOS=js GOARCH=wasm go build -o site/dist/main.wasm ./cmd/wasm

# site assembles the self-contained static build: the wasm engine, the Go
# runtime glue it needs, and the static HTML/CSS/JS shell. Every asset path
# inside site/ is relative so the output works when served from a subpath.
site: wasm
	mkdir -p site/dist
	cp site/index.html site/style.css site/app.js site/dist/
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" site/dist/wasm_exec.js
