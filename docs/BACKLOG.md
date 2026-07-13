# Backlog

Epics and stories for the v1 build. Every story has 1–3 verifiable acceptance criteria —
concrete checks, not vibes. The first story of Epic 1 is the wow moment: it must land before
anything optional gets built.

## Epic 1 — Core engine & wow-moment demo

- [ ] **Wow moment: batch-check five URLs end-to-end via CLI.** Fetch → SPDX detect → clause
  scan → verdict, for a batch of five real URLs, printing one line per URL with badge, license,
  and (for non-green results) the quoted clause.
  - [ ] Running `provenance-check` against 5 real URLs (a mix of MIT/Apache-2.0 repos and a
    fixture carrying a "no AI training" clause) resolves all 5 within a couple of seconds on a
    typical connection.
  - [ ] A plain-MIT-licensed repo is classified CLEAR — no false positive on the flagship case.

- [ ] **SPDX identifier detection from LICENSE files.**
  - [ ] Standard MIT/Apache-2.0/BSD-3-Clause LICENSE boilerplate is identified with the correct
    SPDX ID string.
  - [ ] A LICENSE with no recognizable SPDX text returns an explicit "unknown" state rather than
    a wrong guess.

- [ ] **Non-standard "no AI training" clause pattern library, data-driven and versioned.**
  - [ ] Fixture text for at least 5 known real-world phrasings (OpenRAIL-style behavioral
    restrictions, explicit "not for ML/AI training" addenda, ambiguous CC-BY-NC-on-dataset
    cases, etc.) is flagged RESTRICTED or CAUTION as appropriate, with the matched sentence
    captured verbatim.
  - [ ] Adding a new clause pattern requires only a data-file change plus a test fixture — no
    changes to the matching code.

- [ ] **Concurrent multi-URL fetch with per-URL error isolation.**
  - [ ] One unreachable/404 URL in a batch of five does not block or fail the other four; the
    failed one reports its own error row.
  - [ ] Five reachable URLs are fetched concurrently — batch wall-clock time tracks the slowest
    single fetch, not the sum of all five.

## Epic 2 — GitHub & Hugging Face source support

- [ ] **GitHub repo URL resolution.**
  - [ ] A `github.com/<owner>/<repo>` URL resolves to that repo's LICENSE and README regardless
    of default branch name (`main`, `master`, or other).

- [ ] **Hugging Face dataset/model URL resolution.**
  - [ ] A `huggingface.co/datasets/<name>` or `huggingface.co/<org>/<model>` URL is fetched and
    its card's YAML front-matter `license:` field is parsed into an SPDX identifier when present.
  - [ ] Non-standard usage-restriction text in a Hugging Face card body is scanned by the same
    clause library used for GitHub sources (no duplicated or divergent matching logic).

- [ ] **Malformed/unsupported URL handling.**
  - [ ] Pasting a URL that isn't a GitHub or Hugging Face URL produces a clear inline
    "unsupported source" row — not a crash and not a silent skip.

## Epic 3 — Web UI (exhibit grid)

- [ ] **Paste-box input with batch submission.**
  - [ ] The textarea accepts one URL per line; submitting with 1–50 URLs triggers a check per
    line, with blank lines ignored.

- [ ] **Live-populating exhibit grid matching `docs/DESIGN.md`.**
  - [ ] Each URL renders as a card showing a shimmer loading state immediately on submit, then
    updates in place (no layout jump) to its resolved verdict stamp within a couple of seconds.
  - [ ] Grid, cards, and stamps use the tokens and layout defined in `docs/DESIGN.md` (parchment
    palette, serif+mono type pairing, rotated stamp motif).

- [ ] **Clause detail on non-green results.**
  - [ ] Clicking/tapping a CAUTION or RESTRICTED card expands to show the exact quoted clause
    text and which file it came from (LICENSE vs README/card).

- [ ] **Responsive layout at 390 / 768 / 1440px.**
  - [ ] No horizontal scroll and no overlapping elements at any of the three widths; the
    exhibit grid remains the majority of the viewport at desktop width.

- [ ] **Design polish pass: interaction states, favicon, empty/error states.**
  - [ ] Every interactive control (button, textarea, card) has themed hover/focus/active
    states — no unstyled native widgets.
  - [ ] The favicon and empty-state illustration described in `docs/DESIGN.md` are implemented
    (no default browser globe icon).

## Epic 4 — Ship & deploy

- [ ] **Static build pipeline for the web UI.**
  - [ ] A documented build command produces one self-contained directory (relative asset paths
    only, no leading `/`) that renders and functions correctly when served from an arbitrary
    subpath.

- [ ] **CI coverage for the clause library and URL parsing.**
  - [ ] CI runs `go test ./...` covering the SPDX detector, clause matcher, and URL source
    resolver, and fails the build on any regression.

- [ ] **README quickstart reflects the shipped CLI and web UI.**
  - [ ] README's usage section matches the actual CLI flags/output and links to the deployed
    web UI.
