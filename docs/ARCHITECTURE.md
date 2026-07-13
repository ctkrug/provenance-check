# Architecture

A concise map of the codebase for anyone picking this up cold. See
[`docs/VISION.md`](VISION.md) for the why and [`docs/BACKLOG.md`](BACKLOG.md) for what's
left.

## Module layout

```
cmd/provenance-check/       thin CLI front end
cmd/wasm/                   thin browser front end — compiles to WebAssembly
internal/provenance/        the core engine — the single source of truth for classification
site/                       static HTML/CSS/JS shell for the web UI (source, not build output)
```

`internal/provenance` is intentionally the *only* package with classification logic. The CLI
and the web UI are both thin callers of `provenance.Check` / `provenance.BatchCheck` — see
`docs/VISION.md`'s "one core package, two front ends" design decision.

## Data flow

```
URL string
  -> parseSource()        (source.go)     classify the URL: github / huggingface / unsupported
  -> fetchGitHub() /       (github.go /    fetch LICENSE + README (or card) over HTTP
     fetchHuggingFace()     huggingface.go)
  -> classify()            (classify.go)   SPDX detect + clause scan -> Result
```

`Check(url string) (Result, error)` runs that pipeline for one URL. `BatchCheck(urls []string)
[]BatchResult` runs `Check` concurrently across a batch (one goroutine per URL, results written
to a pre-sized slice by index — no shared mutable state, no mutex needed) so one broken URL
never blocks or fails its neighbors, and wall-clock time tracks the slowest single check.

## Files in `internal/provenance`

| File | Responsibility |
|---|---|
| `provenance.go` | Public API: `Result`, `Verdict`, `Check`, `BatchCheck`. |
| `source.go` | Pure URL parsing (`parseSource`) — no network. Classifies a raw URL into a `parsedSource` (GitHub repo, HF dataset, or HF model) or a descriptive "unsupported source" error. |
| `github.go` | Resolves a repo's actual default branch via the GitHub API, then fetches LICENSE/README from `raw.githubusercontent.com`. A missing file is not an error — that's a legitimate "unknown" outcome. |
| `huggingface.go` | Fetches a card's README (trying `main` then `master`) and a sibling LICENSE if present. Parses the card's YAML front-matter `license:` field into an SPDX-ish override. |
| `spdx.go` | `DetectSPDX(text)` — sniffs standard license boilerplate (MIT, Apache-2.0, BSD-2/3-Clause, ISC, MPL-2.0, GPL-3.0, Unlicense, CC-BY-NC variants) via an ordered list of distinguishing phrases. Returns `ok=false` rather than guessing on unrecognized text. |
| `clauses.json` | The non-standard "no AI training" clause library: reviewable data, not code. Each entry has an id, a verdict (`caution`/`restricted`), a regexp pattern, and a description. |
| `clauses.go` | Loads and compiles `clauses.json` (via `go:embed`) once at package init. |
| `scan.go` | `scanText`/`scanDocuments` — runs the clause library against one or more documents and returns the single strongest match (`restricted` always outranks `caution`). |
| `classify.go` | `classify(classifyInput) Result` — the one place SPDX detection and clause scanning combine into a verdict. Both fetchers funnel through here so GitHub and Hugging Face never risk divergent matching logic. A detected CC-BY-NC license with no explicit clause match still yields a `caution` verdict (NonCommercial terms are ambiguous for ML training) via a synthesized clause. |

## Extending the clause library

Add an entry to `clauses.json` (id, verdict, regexp pattern, description) and a fixture test in
`scan_test.go` proving it matches real-world phrasing. No Go code changes required — that's the
whole point of keeping the library as data (see `docs/VISION.md`).

## Testing approach

- `spdx_test.go`, `clauses_test.go`, `scan_test.go`, `classify_test.go`, `source_test.go` test
  pure logic with no network I/O.
- `github_test.go`, `huggingface_test.go`, `provenance_test.go`, `batch_test.go` test the
  fetchers and the end-to-end `Check`/`BatchCheck` pipeline against a local `httptest` server —
  `githubAPIBase`, `githubRawBase`, and `huggingFaceBase` are package-level vars overridden for
  the duration of a test, so no live network access is required for `go test` to pass.
- `cmd/provenance-check/main_test.go` builds and exercises the real binary, using
  unsupported-host URLs (which fail during `parseSource`, before any HTTP call) to keep CLI
  plumbing tests deterministic without depending on live GitHub/Hugging Face access.

## The web front end (`cmd/wasm/`, `site/`)

`cmd/wasm` compiles `internal/provenance` to WebAssembly and exposes one function to the
browser: `window.provenanceCheck(url)`, returning a Promise that resolves to a JSON-decoded
result. It runs entirely client-side — no backend, no proxy:

```
cmd/wasm/result.go     JSON conversion (provenance.Result -> checkResult); no syscall/js,
                        so it's a normal package on any GOOS and unit-testable with `go test`.
cmd/wasm/main.go        //go:build js && wasm — the syscall/js wiring. Registers
                        provenanceCheck(url) and parks in select{} to keep callbacks alive.
```

Each call spawns its own goroutine, so pasting N URLs and calling `provenanceCheck` once per
URL from JS resolves them concurrently — the same one-broken-URL-doesn't-block-the-rest
guarantee `BatchCheck` gives the CLI, just driven from the JS side instead of `sync.WaitGroup`.

This works from a browser (not from Node — see below) because both GitHub and Hugging Face
send CORS headers that allow a browser-origin `fetch()`: `raw.githubusercontent.com` and
`api.github.com` send `Access-Control-Allow-Origin: *`; `huggingface.co` reflects the
request's `Origin`. Go's `net/http` on `GOOS=js` transparently backs `http.Get` with the
Fetch API (`net/http/roundtrip_js.go` in the Go stdlib), so `github.go`/`huggingface.go` need
no browser-specific code at all.

`site/` holds the static shell as source: `index.html`, `style.css`, `app.js`,
`wasm_exec.js` (copied from `$(go env GOROOT)/misc/wasm/wasm_exec.js` — the Go runtime's JS
glue for instantiating a wasm module). `app.js` loads the wasm module, then drives the exhibit
grid: parses the pasted URLs, creates a card per URL in a loading state, and calls
`provenanceCheck` once per URL, updating each card in place as its own promise resolves.

`make site` builds `site/dist/` — the wasm binary, `wasm_exec.js`, and the static files
together in one self-contained, relative-path directory (gitignored; not committed). See
`docs/DESIGN.md` for the exhibit-binder visual direction the CSS implements.

### A Go/wasm gotcha worth knowing

Go's stdlib deliberately disables its Fetch-based transport when it detects it's running
under Node.js (`jsFetchDisabled` in `net/http/roundtrip_js.go`, for Go's own test suite —
see go.dev/issue/57613) and falls back to real `net.Dial`, which doesn't work outside a
browser. **Don't use `node wasm_exec_node.js` to test network behavior** — it will fail with
a `dial tcp` error that has nothing to do with the actual browser code path. Verify the wasm
build with a real browser (this project used headless Chromium via Playwright during BUILD).

## Running it

```sh
go build -o bin/provenance-check ./cmd/provenance-check
go test -race -cover ./...              # or: make test
./bin/provenance-check <url> [<url> ...]

make site                               # builds site/dist/ — open site/dist/index.html
                                         # via a local static server (not file://, since
                                         # WebAssembly.instantiateStreaming needs http(s))
```
