# Changelog

All notable changes to this project are documented here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

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
