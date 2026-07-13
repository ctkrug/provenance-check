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

loadEngine()
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

const grid = document.getElementById("exhibit-grid");
const emptyState = document.getElementById("empty-state");
const cardTemplate = document.getElementById("exhibit-card-template");

// createCard clones the <template> for one URL, in its loading state, and
// appends it to the grid. Returns the new card element so the caller can
// update it in place once the check resolves.
function createCard(url, index) {
  const fragment = cardTemplate.content.cloneNode(true);
  const card = fragment.querySelector(".exhibit-card");
  card.dataset.url = url;
  card.querySelector(".exhibit-card__index").textContent =
    "Exhibit " + String(index + 1).padStart(2, "0");
  card.querySelector(".exhibit-card__stamp-slot").innerHTML = stampSVG(null, { loading: true });
  card.querySelector(".exhibit-card__url").textContent = url;
  grid.appendChild(card);
  return card;
}

// resolveCard updates a loading card in place with its settled verdict —
// no layout jump, just the shimmer and pulsing stamp replaced by the real
// result and a stamp "hit" flourish.
function resolveCard(card, payload) {
  card.classList.remove("exhibit-card--loading");
  card.classList.add("exhibit-card--" + payload.verdict);

  const stampSlot = card.querySelector(".exhibit-card__stamp-slot");
  stampSlot.innerHTML = stampSVG(payload.verdict);
  const stampEl = stampSlot.querySelector(".stamp");
  stampEl.classList.add("stamp--hit");

  const licenseEl = card.querySelector(".exhibit-card__license");
  const toggle = card.querySelector(".exhibit-card__toggle");
  const clauseSection = card.querySelector(".exhibit-card__clause");

  if (payload.verdict === "error") {
    licenseEl.textContent = "unavailable";
    toggle.hidden = false;
    toggle.textContent = "View error";
    clauseSection.querySelector(".exhibit-card__clause-source").textContent = "Fetch failed";
    clauseSection.querySelector(".exhibit-card__clause-text").textContent = payload.error;
    return;
  }

  licenseEl.textContent = payload.license || "unknown";

  if (payload.clause) {
    toggle.hidden = false;
    toggle.textContent = "View clause";
    clauseSection.querySelector(".exhibit-card__clause-source").textContent =
      "Source: " + (payload.source || "unknown");
    clauseSection.querySelector(".exhibit-card__clause-text").textContent = payload.clause;
  }
}

function setToggleHandler(card) {
  const toggle = card.querySelector(".exhibit-card__toggle");
  const clauseSection = card.querySelector(".exhibit-card__clause");
  toggle.addEventListener("click", () => {
    const expanded = toggle.getAttribute("aria-expanded") === "true";
    toggle.setAttribute("aria-expanded", String(!expanded));
    clauseSection.hidden = expanded;
  });
}

const form = document.getElementById("check-form");
const urlsField = document.getElementById("urls");
const clearButton = document.getElementById("clear-button");

function resetGrid() {
  grid.innerHTML = "";
  grid.appendChild(emptyState);
}

async function runBatch(urls) {
  grid.innerHTML = "";
  let resolved = 0;
  statusLine.textContent = `0 of ${urls.length} resolved`;

  const cards = urls.map((url, index) => {
    const card = createCard(url, index);
    setToggleHandler(card);
    return card;
  });

  await Promise.all(
    urls.map(async (url, index) => {
      const payload = await window.provenanceCheck(url);
      resolveCard(cards[index], payload);
      resolved += 1;
      statusLine.textContent = `${resolved} of ${urls.length} resolved`;
    })
  );
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const urls = parseURLs(urlsField.value);
  if (urls.length === 0) {
    statusLine.textContent = "Paste at least one URL first.";
    return;
  }

  checkButton.disabled = true;
  urlsField.readOnly = true;
  try {
    await runBatch(urls);
  } finally {
    checkButton.disabled = false;
    urlsField.readOnly = false;
  }
});

clearButton.addEventListener("click", () => {
  urlsField.value = "";
  urlsField.focus();
  resetGrid();
  statusLine.textContent = "";
});

const collapseToggle = document.getElementById("collapse-toggle");
const panelBody = document.getElementById("panel-body");

collapseToggle.addEventListener("click", () => {
  const expanded = collapseToggle.getAttribute("aria-expanded") === "true";
  collapseToggle.setAttribute("aria-expanded", String(!expanded));
  panelBody.hidden = expanded;
});

