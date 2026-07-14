# Changelog

All notable changes to this project are documented here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.0.0] - 2026-07-14

First public release. The product ships under the name **Provenance** (the repository slug and
live URL stay `provenance-check`).

### Added

- Product landing: the web front end now carries the "Provenance" wordmark and favicon, a
  "View on GitHub" call to action, a below-the-fold explainer and FAQ answering common
  "can I train on this dataset" questions, search metadata (title, description, OpenGraph),
  and a portfolio cross-promo link.
- `docs/launch/devto.md`: launch write-up covering the Go/WASM engine and data-driven clauses.
- Project scaffold: Go module, CLI entrypoint, and `internal/provenance` core package.
- CI workflow (gofmt, vet, golangci-lint, build, test).
- `docs/VISION.md`, `docs/DESIGN.md`, and `docs/BACKLOG.md`.
- Core classification engine: SPDX license detection, a data-driven non-standard
  "no AI training" clause library, and GitHub/Hugging Face source resolution.
- Concurrent `BatchCheck` with per-URL error isolation.
- CLI now prints a verdict badge, SPDX license, and quoted clause per URL, and exits
  non-zero on any restricted or unresolvable result.
- `docs/ARCHITECTURE.md`.
- Web UI: the classification engine compiled to WebAssembly (`cmd/wasm`), exposed to the
  browser as `provenanceCheck(url)`, running entirely client-side with no backend server.
- `site/`: an exhibit-binder-styled static front end (parchment palette, rotated verdict
  stamps, live-populating exhibit grid, expandable clause quotes) matching `docs/DESIGN.md`,
  responsive at 390/768/1440px and buildable to a self-contained `site/dist/` via `make site`.
- `site/logic.js` + `site/logic.test.js`: the web UI's pure logic (URL parsing, stamp sizing)
  is now unit-tested via Node's built-in test runner, wired into CI.

### Fixed

- CLI: a single stdin line longer than 64KB no longer silently discards the rest of the
  pasted batch (switched from `bufio.Scanner` to `bufio.Reader`).
- `internal/provenance`: LICENSE/README fetches are now capped at 5MiB, guarding against a
  huge or malicious response being buffered entirely into memory.
- `internal/provenance`: Hugging Face license slugs in the CC-BY-NC family are now matched
  case-insensitively, so unrecognized-but-real lowercase slugs (e.g. `cc-by-nc-nd-4.0`) no
  longer fall through to a false "clear" verdict.
- Web UI: the Clear button is now disabled during an in-flight batch check, preventing a
  stale "N of M resolved" status from reappearing after the grid has been cleared.
