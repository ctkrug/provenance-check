// Command wasm compiles the provenance-check engine to WebAssembly so the
// static web UI can run real checks with no backend server of its own.
package main

import (
	"encoding/json"

	"github.com/ctkrug/provenance-check/internal/provenance"
)

// checkResult is the JSON-serializable shape handed back to JavaScript for
// one checked URL. A non-empty Error means the check failed before
// producing a verdict (network error, unsupported source, etc.), in which
// case Verdict is always "error".
type checkResult struct {
	URL     string `json:"url"`
	Verdict string `json:"verdict"`
	License string `json:"license"`
	Clause  string `json:"clause"`
	Source  string `json:"source"`
	Error   string `json:"error,omitempty"`
}

// toCheckResult converts one provenance.Check outcome into its
// JSON-serializable form. Kept free of syscall/js so it's plain data
// transformation, unit-testable without a wasm build.
func toCheckResult(url string, result provenance.Result, err error) checkResult {
	if err != nil {
		return checkResult{URL: url, Verdict: "error", Error: err.Error()}
	}
	return checkResult{
		URL:     url,
		Verdict: string(result.Verdict),
		License: result.License,
		Clause:  result.Clause,
		Source:  result.Source,
	}
}

// checkJSON runs provenance.Check for one URL and marshals the outcome to
// a JSON string ready to hand to JSON.parse on the JS side.
func checkJSON(url string) (string, error) {
	result, err := provenance.Check(url)
	data, marshalErr := json.Marshal(toCheckResult(url, result, err))
	if marshalErr != nil {
		return "", marshalErr
	}
	return string(data), nil
}
