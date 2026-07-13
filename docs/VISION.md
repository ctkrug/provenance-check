# Vision

## The problem

Anyone assembling a training set — a hobbyist fine-tuning a model, a researcher building a
corpus, a startup scraping public repos — eventually has to answer a legal question fast:
"can I actually train on this?" SPDX-tagged open-source licenses mostly say nothing about ML
training use one way or the other, but a growing minority of datasets and repos carry
non-standard restrictions bolted onto a README or a custom LICENSE file: "not for use in
training machine learning models," "no AI training," OpenRAIL-style behavioral-use clauses,
or a CC-BY-NC license applied to a dataset in a way that's ambiguous for commercial ML use.

Generic license scanners (built for software supply-chain compliance) don't catch these,
because "restricts AI training" isn't a concept their taxonomies model. The result: people
either ignore the question and hope, or read fifty licenses by hand before they can start a
training run.

## Who it's for

- ML engineers and researchers assembling a training corpus from public GitHub repos and
  Hugging Face datasets, who want a fast pre-flight check before committing compute.
- Open-source maintainers who want to verify their own dataset's license reads the way they
  intend it to.
- Anyone who has heard "check the license" as advice and wants a tool that actually tells them
  what the license says about their specific use case, not just its SPDX name.

This is explicitly not a general software-license-compliance tool (there are plenty of those).
It is single-purpose: paste URLs, get an AI-training-specific answer.

## The core idea

Paste a list of URLs. For each one, fetch the LICENSE and README, run two passes over the
text — SPDX identifier extraction, then a maintained pattern library of known non-standard
"no AI training" phrasings — and render a badge (green / yellow / red) with the exact clause
quoted. No verdict is ever a black box: every non-green result points at the sentence that
caused it and which file it came from.

## Key design decisions

- **Go, stdlib-first.** No heavy web framework, no ORM, minimal third-party dependencies. A
  license-clause classifier doesn't need much beyond an HTTP client and a text scanner, and
  a small dependency surface keeps the trust story simple for a tool making legal-adjacent
  claims about someone else's project.
- **One core package, two front ends.** `internal/provenance` (and its sibling packages added
  during build) is the single source of truth for classification logic. The CLI and the web UI
  are both thin callers of the same package — never two parsers that can drift out of sync.
- **Clause library as data, not code sprawl.** Non-standard clause patterns live in a
  reviewable, versioned list (not scattered regex calls), so adding a newly-discovered "no AI
  training" phrasing is a data change, not a refactor.
- **Every verdict is explainable.** A verdict without the source quote is worse than no verdict
  — it invites blind trust. The exact clause and which file it came from is always shown.
- **Static, deployable front end.** The web UI builds to a single static directory with
  relative asset paths, so it can be hosted at a subpath (`apps.charliekrug.com/provenance-check`)
  with no server of its own; it calls the same classification logic compiled to run client-side
  or via a small stateless endpoint (decided in BUILD once the fetch/CORS approach is settled).
- **Speed is a feature.** The wow moment is five URLs resolving to badges within a couple of
  seconds — that means concurrent fetches and no unnecessary round trips, not just correct
  parsing.

## What "v1 done" looks like

- CLI and web UI both take a list of URLs and return a verdict badge, SPDX identifier (if any),
  and — for non-green results — the exact clause and its source file, for both GitHub repos and
  Hugging Face datasets/models.
- The non-standard clause library catches the known common phrasings (OpenRAIL-style behavioral
  restrictions, explicit "no AI/ML training" addenda, ambiguous NC-license-on-dataset cases) with
  test coverage proving each pattern matches its real-world source text.
- Batch checks of five URLs complete and render within a couple of seconds under normal network
  conditions.
- The web UI is deployed as a static site with relative asset paths, matches `docs/DESIGN.md`'s
  direction, and works at desktop and phone widths.
- CI is green: build, vet, and tests pass with no known false-negative on a documented clause
  in the test fixtures.
