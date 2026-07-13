# Backlog

Epics and stories for the v1 build. Every story has 1–3 verifiable acceptance criteria —
concrete checks, not vibes. The first story of Epic 1 is the wow moment: it must land before
anything optional gets built.

## Epic 1 — Core engine & wow-moment demo

- [x] **Wow moment: batch-check five URLs end-to-end via CLI.** Fetch → SPDX detect → clause
  scan → verdict, for a batch of five real URLs, printing one line per URL with badge, license,
  and (for non-green results) the quoted clause.
  - [x] Running `provenance-check` against 5 real URLs (a mix of MIT/Apache-2.0 repos and a
    fixture carrying a "no AI training" clause) resolves all 5 within a couple of seconds on a
    typical connection.
  - [x] A plain-MIT-licensed repo is classified CLEAR — no false positive on the flagship case.

- [x] **SPDX identifier detection from LICENSE files.**
  - [x] Standard MIT/Apache-2.0/BSD-3-Clause LICENSE boilerplate is identified with the correct
    SPDX ID string.
  - [x] A LICENSE with no recognizable SPDX text returns an explicit "unknown" state rather than
    a wrong guess.

- [x] **Non-standard "no AI training" clause pattern library, data-driven and versioned.**
  - [x] Fixture text for at least 5 known real-world phrasings (OpenRAIL-style behavioral
    restrictions, explicit "not for ML/AI training" addenda, ambiguous CC-BY-NC-on-dataset
    cases, etc.) is flagged RESTRICTED or CAUTION as appropriate, with the matched sentence
    captured verbatim.
  - [x] Adding a new clause pattern requires only a data-file change plus a test fixture — no
    changes to the matching code.

- [x] **Concurrent multi-URL fetch with per-URL error isolation.**
  - [x] One unreachable/404 URL in a batch of five does not block or fail the other four; the
    failed one reports its own error row.
  - [x] Five reachable URLs are fetched concurrently — batch wall-clock time tracks the slowest
    single fetch, not the sum of all five.

## Epic 2 — GitHub & Hugging Face source support

- [x] **GitHub repo URL resolution.**
  - [x] A `github.com/<owner>/<repo>` URL resolves to that repo's LICENSE and README regardless
    of default branch name (`main`, `master`, or other).

- [x] **Hugging Face dataset/model URL resolution.**
  - [x] A `huggingface.co/datasets/<name>` or `huggingface.co/<org>/<model>` URL is fetched and
    its card's YAML front-matter `license:` field is parsed into an SPDX identifier when present.
  - [x] Non-standard usage-restriction text in a Hugging Face card body is scanned by the same
    clause library used for GitHub sources (no duplicated or divergent matching logic).

- [x] **Malformed/unsupported URL handling.**
  - [x] Pasting a URL that isn't a GitHub or Hugging Face URL produces a clear inline
    "unsupported source" row — not a crash and not a silent skip.

## Epic 3 — Web UI (exhibit grid)

- [x] **Paste-box input with batch submission.**
  - [x] The textarea accepts one URL per line; submitting with 1–50 URLs triggers a check per
    line, with blank lines ignored.

- [x] **Live-populating exhibit grid matching `docs/DESIGN.md`.**
  - [x] Each URL renders as a card showing a shimmer loading state immediately on submit, then
    updates in place (no layout jump) to its resolved verdict stamp within a couple of seconds.
  - [x] Grid, cards, and stamps use the tokens and layout defined in `docs/DESIGN.md` (parchment
    palette, serif+mono type pairing, rotated stamp motif).

- [x] **Clause detail on non-green results.**
  - [x] Clicking/tapping a CAUTION or RESTRICTED card expands to show the exact quoted clause
    text and which file it came from (LICENSE vs README/card).

- [x] **Responsive layout at 390 / 768 / 1440px.**
  - [x] No horizontal scroll and no overlapping elements at any of the three widths; the
    exhibit grid remains the majority of the viewport at desktop width.

- [x] **Design polish pass: interaction states, favicon, empty/error states.**
  - [x] Every interactive control (button, textarea, card) has themed hover/focus/active
    states — no unstyled native widgets.
  - [x] The favicon and empty-state illustration described in `docs/DESIGN.md` are implemented
    (no default browser globe icon).

## Epic 4 — Ship & deploy

- [x] **Static build pipeline for the web UI.**
  - [x] A documented build command produces one self-contained directory (relative asset paths
    only, no leading `/`) that renders and functions correctly when served from an arbitrary
    subpath.

- [x] **CI coverage for the clause library and URL parsing.**
  - [x] CI runs `go test ./...` covering the SPDX detector, clause matcher, and URL source
    resolver, and fails the build on any regression.

- [ ] **README quickstart reflects the shipped CLI and web UI.**
  - [x] README's usage section matches the actual CLI flags/output.
  - [ ] Links to the deployed web UI — pending CLOSEOUT's deploy to
    `apps.charliekrug.com/provenance-check`.
