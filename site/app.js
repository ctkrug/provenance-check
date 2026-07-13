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
