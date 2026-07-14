# Provenance

**▶ Live demo: [apps.charliekrug.com/provenance-check](https://apps.charliekrug.com/provenance-check/)**

Check if you can train on that dataset. Paste a list of dataset or repo URLs and get a
plain-English verdict on any license that restricts AI/ML training use, before you commit
compute to it.

[![CI](https://github.com/ctkrug/provenance-check/actions/workflows/ci.yml/badge.svg)](https://github.com/ctkrug/provenance-check/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why

License text for datasets and model-adjacent repos is a minefield. SPDX identifiers cover the
common cases, but a growing number of projects bolt on non-standard clauses like "no use in
training machine learning models" or "not for AI training purposes" directly into a README or
a custom LICENSE file. Generic license scanners built for software compliance do not look for
these clauses, because they are not part of any SPDX taxonomy. Provenance is a single-purpose
tool built for exactly one question: **can I train on this?**

## What it does

Paste GitHub and Hugging Face URLs, one per line. For each one, Provenance:

1. Fetches the repo's LICENSE and README.
2. Identifies the SPDX license identifier where present.
3. Scans the full text for known non-standard "no AI training" clauses (a maintained pattern
   library, not a single regex), including phrasing common to OpenRAIL-style restrictions,
   CC-BY-NC variants misapplied to ML use, and ad-hoc "do not use for training" addenda.
4. Returns a verdict with the exact clause and its source file quoted: **CLEAR** (no restriction
   found), **CAUTION** (ambiguous, such as a NonCommercial or research-only grant), or
   **RESTRICTED** (an explicit training restriction).

## Features

- **CLI**: `provenance-check <url> [<url> ...]` prints one line per URL, checked concurrently.
- **Web UI**: paste a list and watch live exhibit cards fill in with a verdict stamp as each URL
  resolves. Runs entirely in the browser (the engine compiles to WebAssembly; no backend server).
- **SPDX detection** from LICENSE text (MIT, Apache-2.0, BSD-2/3-Clause, ISC, MPL-2.0, GPL-3.0,
  Unlicense, CC-BY-NC variants), returning an explicit "unknown" rather than a wrong guess.
- **Data-driven clause library**, versioned and reviewable independently of the code.
- **Hugging Face** dataset and model card parsing (YAML front matter plus license field).
- **Exact clause quoted** for every non-green result, so no verdict is a black box.
- **Batch mode**: every URL in a run is checked concurrently and isolated from its neighbors'
  errors, so one broken URL never fails the rest.

## Usage

```sh
go build -o bin/provenance-check ./cmd/provenance-check

./bin/provenance-check \
  https://github.com/expressjs/express \
  https://huggingface.co/gpt2
```

```
CLEAR      MIT              https://github.com/expressjs/express
CLEAR      MIT              https://huggingface.co/gpt2
```

Pipe a list instead, one URL per line:

```sh
cat urls.txt | ./bin/provenance-check
```

A `RESTRICTED` or `CAUTION` result adds an indented line quoting the exact clause and which
file it came from:

```
RESTRICTED unknown          https://huggingface.co/datasets/example/no-training-dataset
           clause: "not permitted to use this dataset for AI training purposes." (README.md)
```

(That `huggingface.co/datasets/example/no-training-dataset` is illustrative, not a live dataset.
Swap in any real URL from a project with a non-standard training-restriction clause to see this
for yourself.)

Each result line is `<verdict> <license> <url>`. The process exits non-zero if any URL is
restricted or fails to resolve (unsupported host, unreachable repo, and similar). Only
`github.com/<owner>/<repo>`, `huggingface.co/<org>/<model>`, and
`huggingface.co/datasets/<name>` URLs are supported today; anything else reports a clear
"unsupported source" error instead of a crash or a silent skip.

### Web UI

```sh
make site                                            # builds site/dist/: wasm engine + static HTML/CSS/JS
python3 -m http.server 8080 --directory site/dist    # any static file server works
```

Open `http://localhost:8080/` and paste up to 50 GitHub or Hugging Face URLs, one per line. The
classification engine (the same `internal/provenance` package the CLI uses) runs entirely in
your browser as WebAssembly, so no URL you paste is sent anywhere but the source itself. Each
exhibit card fills in as its own check resolves; clicking a `CAUTION` or `RESTRICTED` card
expands the quoted clause. `WebAssembly.instantiateStreaming` needs the page served over
http(s), not opened as a `file://` URL.

## Stack

Go, stdlib-first (`net/http`, no heavy framework). The CLI and the static web front end both
call the same core parsing package, so the license-clause logic is never duplicated between
them.

## Documentation

- [`docs/VISION.md`](docs/VISION.md): the problem, who it is for, and what "v1 done" means.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md): module map, data flow, and the Go/WASM setup.
- [`docs/DESIGN.md`](docs/DESIGN.md): the exhibit-binder visual direction and tokens.
- [`docs/BACKLOG.md`](docs/BACKLOG.md): the build breakdown with acceptance criteria.

## License

MIT, see [`LICENSE`](LICENSE).

---

More of Charlie's projects &rarr; [apps.charliekrug.com](https://apps.charliekrug.com)
