# Provenance Check

[![CI](https://github.com/ctkrug/provenance-check/actions/workflows/ci.yml/badge.svg)](https://github.com/ctkrug/provenance-check/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Paste a list of dataset or repo URLs you're about to train on. Get a plain-English flag on
any license that restricts AI/ML training use — in seconds, before you commit compute to it.

## Why

License text for datasets and model-adjacent repos is a minefield: SPDX identifiers cover the
common cases, but a growing number of projects bolt on non-standard clauses like "no use in
training machine learning models" or "not for AI training purposes" directly into a README or
a custom LICENSE file. Generic license scanners built for software compliance don't look for
these clauses because they're not part of any SPDX taxonomy. Provenance Check is a
single-purpose tool built for exactly one question: **can I train on this?**

## What it does

Paste GitHub and Hugging Face URLs (one per line). For each one, Provenance Check:

1. Fetches the repo's LICENSE and README.
2. Identifies the SPDX license identifier where present.
3. Scans the full text for known non-standard "no AI training" clauses (a maintained pattern
   library, not a single regex) — including phrasing common to datasets published under
   OpenRAIL-style restrictions, CC-BY-NC variants misapplied to ML use, and ad-hoc "do not use
   for training" addenda.
4. Renders a badge — green (clear to train), yellow (ambiguous / restricted-but-permissive with
   conditions), red (explicit training restriction) — with the exact clause and its source
   quoted alongside.

## Planned features

- [x] CLI: `provenance-check <url> [<url> ...]` — prints one line per URL, concurrently checked.
- [ ] Web UI: paste a list, get live rows with badges as each URL resolves.
- [x] SPDX identifier detection from `LICENSE` file text (MIT, Apache-2.0, BSD-2/3-Clause, ISC,
      MPL-2.0, GPL-3.0, Unlicense, CC-BY-NC variants).
- [x] Non-standard clause pattern library, versioned and reviewable independently of code.
- [x] Hugging Face dataset/model card parsing (YAML front matter + license field).
- [x] Exact clause + source file quoted for every non-green result (no black-box verdicts).
- [x] Batch mode: every URL in a run is checked concurrently, isolated from its neighbors' errors.

## Usage

```sh
go build -o bin/provenance-check ./cmd/provenance-check

./bin/provenance-check \
  https://github.com/expressjs/express \
  https://huggingface.co/gpt2 \
  https://huggingface.co/datasets/example/no-training-dataset

# or pipe a list, one URL per line:
cat urls.txt | ./bin/provenance-check
```

```
CLEAR      MIT              https://github.com/expressjs/express
CLEAR      MIT              https://huggingface.co/gpt2
RESTRICTED unknown          https://huggingface.co/datasets/example/no-training-dataset
           clause: "not permitted to use this dataset for AI training purposes." (README.md)
```

Each result line is `<verdict> <license> <url>`; a `RESTRICTED` or `CAUTION` result is
followed by an indented line quoting the exact clause and which file it came from. The
process exits non-zero if any URL is restricted or fails to resolve (unsupported host,
unreachable repo, and similar). Only `github.com/<owner>/<repo>` and
`huggingface.co/<org>/<model>` / `huggingface.co/datasets/<name>` URLs are supported today;
anything else reports a clear "unsupported source" error instead of a crash or a silent skip.

## Stack

Go (stdlib-first: `net/http`, no heavy framework). A static web front end talks to the same
core parsing package the CLI uses, so the license-clause logic is never duplicated.

## Status

The core engine is functional end-to-end: paste GitHub or Hugging Face URLs, get a verdict,
license, and quoted clause via the CLI. The web UI (`docs/BACKLOG.md` Epic 3) is not built
yet. See [`docs/VISION.md`](docs/VISION.md) for the full plan and
[`docs/BACKLOG.md`](docs/BACKLOG.md) for the build breakdown.

## License

MIT — see [`LICENSE`](LICENSE).
