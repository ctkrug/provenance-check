// Provenance Check — browser front end. Loads the Go/WASM engine built
// from cmd/wasm and drives the exhibit grid; no backend server involved.
"use strict";

const statusLine = document.getElementById("status-line");
const checkButton = document.getElementById("check-button");

async function loadEngine() {
  const go = new Go();
  let result;
  try {
    result = await WebAssembly.instantiateStreaming(fetch("./main.wasm"), go.importObject);
  } catch (streamingError) {
    // instantiateStreaming requires the server to report application/wasm;
    // fall back to a plain fetch + instantiate for hosts that don't.
    const response = await fetch("./main.wasm");
    const bytes = await response.arrayBuffer();
    result = await WebAssembly.instantiate(bytes, go.importObject);
  }

  go.run(result.instance); // never resolves — the wasm program parks in select{}

  while (!window.provenanceCheckReady) {
    await new Promise((resolve) => setTimeout(resolve, 0));
  }
}

const engineReady = loadEngine()
  .then(() => {
    statusLine.textContent = "Engine ready.";
    checkButton.disabled = false;
  })
  .catch((err) => {
    console.error("provenance-check: failed to load the wasm engine", err);
    statusLine.textContent = "Could not load the checking engine. Reload the page to try again.";
  });

checkButton.disabled = true;

const MAX_URLS = 50;

// parseURLs splits the textarea into one URL per non-blank line, capped at
// MAX_URLS (the backlog's documented batch ceiling).
function parseURLs(raw) {
  return raw
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line.length > 0)
    .slice(0, MAX_URLS);
}

const VERDICT_LABELS = {
  clear: "CLEAR",
  caution: "CAUTION",
  restricted: "RESTRICTED",
  error: "ERROR",
};

// stampFontSize shrinks the label to fit the ring as word length grows —
// "CLEAR" reads comfortably at 9px, "RESTRICTED" needs to go condensed.
function stampFontSize(label) {
  if (label.length <= 5) return 9;
  if (label.length <= 7) return 7.5;
  return 6;
}

// stampSVG builds the rotated double-ring verdict badge described in
// docs/DESIGN.md. `loading` renders a dashed placeholder ring instead of a
// resolved verdict.
function stampSVG(verdict, { loading = false } = {}) {
  const modifier = loading ? "loading" : verdict;
  const label = loading ? "" : VERDICT_LABELS[verdict] || VERDICT_LABELS.error;
  const labelMarkup = label
    ? `<text x="32" y="35" class="stamp__label" style="font-size:${stampFontSize(label)}px">${label}</text>`
    : "";
  return `
    <svg class="stamp stamp--${modifier}" viewBox="0 0 64 64" role="img" aria-label="${loading ? "Checking" : "Verdict: " + label}">
      <circle cx="32" cy="32" r="28" class="stamp__ring-outer"${loading ? ' stroke-dasharray="5 4"' : ""}></circle>
      <circle cx="32" cy="32" r="21" class="stamp__ring-inner"></circle>
      ${labelMarkup}
    </svg>
  `;
}

