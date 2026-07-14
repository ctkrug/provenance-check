// Pure, DOM-free logic shared by the browser UI (app.js, loaded as a
// classic script — these become plain globals) and its Node test suite
// (logic.test.js, loaded via require). No framework, no build step.
"use strict";

// MAX_URLS is the backlog's documented batch ceiling.
const MAX_URLS = 50;

// parseURLs splits the textarea into one URL per non-blank line, capped at
// MAX_URLS.
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

if (typeof module !== "undefined" && module.exports) {
  module.exports = { MAX_URLS, parseURLs, VERDICT_LABELS, stampFontSize };
}
