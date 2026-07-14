---
title: I built a tool to check if you can train on a dataset
published: false
tags: go, webassembly, machinelearning, opensource
---

Every time I sit down to fine-tune a model, I hit the same boring question before I write a
line of code: can I actually train on this data? The honest answer is that the SPDX tag at the
top of a repo does not tell you. A dataset can be labelled MIT and still carry a line in its
README that says "not for use in training machine learning models." Generic license scanners,
the ones built for software supply-chain compliance, skip that line completely, because "no AI
training" is not a concept their taxonomies model.

So I built **Provenance**: paste a list of GitHub or Hugging Face URLs, and it returns a
plain-English verdict per source. CLEAR, CAUTION, or RESTRICTED, with the exact clause quoted
so you can check the call yourself. It runs as a CLI and as a web app, and I want to write up
two decisions that made it more fun to build than I expected.

## Decision one: one engine, compiled twice

The classification logic lives in a single Go package, `internal/provenance`. The CLI is a thin
caller of it. The interesting part is the web UI: instead of standing up a backend to proxy the
fetches, I compiled the exact same package to WebAssembly and shipped it as a static site.

The thing that makes this work with no server at all is CORS. When the browser needs a repo's
LICENSE and README, it fetches them directly: `raw.githubusercontent.com` and `api.github.com`
send `Access-Control-Allow-Origin: *`, and `huggingface.co` reflects the request origin. Go's
`net/http` on `GOOS=js` transparently backs `http.Get` with the browser Fetch API, so
`github.go` and `huggingface.go` needed zero browser-specific code. The CLI and the tab run
byte-for-byte the same matcher. There is no second parser to drift out of sync.

One trap cost me an afternoon: Go's standard library deliberately disables that Fetch transport
when it detects Node.js and falls back to real `net.Dial`, which fails outside a browser. If you
try to smoke-test wasm network behavior with `node wasm_exec_node.js`, you get a `dial tcp`
error that has nothing to do with your code. Test the wasm build in an actual browser. I used
headless Chromium.

## Decision two: the clause library is data, not code

The non-standard clauses are the whole point of the tool, and new phrasings show up all the
time. If each one were a hand-written regex call buried in a function, adding a pattern would
mean a refactor and a review of matching logic. So the clauses live in a `clauses.json` file:
an id, a verdict, a regex pattern, and a description per entry. It is embedded with `go:embed`
and compiled once at package load.

Adding a newly discovered "no AI training" phrasing is now a data change plus one fixture test
proving it matches real-world text. No Go changes. That separation also made the tool easier to
trust: the rules a verdict is based on are a reviewable list, not scattered code.

## What I would do differently

The clause library is small and regex-based, which means the real risk is false negatives:
phrasings I have not seen yet slip through as CLEAR. A stronger version would be built against a
much larger corpus of real restrictive licenses, and might lean on fuzzier matching for the long
tail. The verdicts are also a heuristic, not legal advice, and the tool says so. It is a
pre-flight check to catch the obvious cases fast, not a substitute for reading the license when
the stakes are high.

If you pull training data from public repos, I would love to know which restriction phrasings
you have run into that Provenance misses.

- Live demo: https://apps.charliekrug.com/provenance-check/
- Source: https://github.com/ctkrug/provenance-check
