# Architecture

A concise map of the codebase for anyone picking this up cold. See
[`docs/VISION.md`](VISION.md) for the why and [`docs/BACKLOG.md`](BACKLOG.md) for what's
left.

## Module layout

```
cmd/provenance-check/       thin CLI front end
internal/provenance/        the core engine — the single source of truth for classification
```

`internal/provenance` is intentionally the *only* package with classification logic. The CLI
(and, later, the web UI) are both thin callers of `provenance.Check` / `provenance.BatchCheck`
— see `docs/VISION.md`'s "one core package, two front ends" design decision.

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

## Running it

```sh
go build -o bin/provenance-check ./cmd/provenance-check
go test -race -cover ./...              # or: make test
./bin/provenance-check <url> [<url> ...]
```

## Not yet built

The web UI (`docs/BACKLOG.md` Epic 3) doesn't exist yet — no `site/` directory, no static
build. The CLI is the only front end today.
