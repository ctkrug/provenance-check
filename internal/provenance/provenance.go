// Package provenance is the core license-clause classification engine shared
// by the CLI and the web front end. It answers one question per URL: can you
// train on this?
package provenance

import "sync"

// Verdict is the plain-English flag rendered as a badge.
type Verdict string

const (
	// VerdictClear means no training restriction was found.
	VerdictClear Verdict = "clear"
	// VerdictCaution means the license is ambiguous or permissive-with-conditions.
	VerdictCaution Verdict = "caution"
	// VerdictRestricted means an explicit AI/ML training restriction was found.
	VerdictRestricted Verdict = "restricted"
)

// Result is the outcome of checking a single URL.
type Result struct {
	URL     string
	Verdict Verdict
	License string // SPDX identifier, if one was found
	Clause  string // the exact restrictive clause text, if any
	Source  string // which file the clause was found in (e.g. "LICENSE", "README.md")
}

// Check fetches and classifies a single dataset or repo URL: resolve the
// source, fetch its LICENSE/README (or card), then run SPDX detection and
// clause-library scanning over the result.
func Check(rawURL string) (Result, error) {
	src, err := parseSource(rawURL)
	if err != nil {
		return Result{URL: rawURL}, err
	}

	var in classifyInput
	switch src.Kind {
	case sourceGitHub:
		in, err = fetchGitHub(src)
	case sourceHuggingFace:
		in, err = fetchHuggingFace(src)
	}
	if err != nil {
		return Result{URL: rawURL}, err
	}

	result := classify(in)
	result.URL = rawURL
	return result, nil
}

// BatchResult pairs a Check outcome with its URL's index-preserving slot,
// so a batch can report one broken URL without losing the others.
type BatchResult struct {
	Result Result
	Err    error
}

// BatchCheck runs Check concurrently across every URL. Each URL's fetch and
// classification is isolated from the others: one failure doesn't block or
// fail the rest, and results preserve input order. Wall-clock time tracks
// the slowest single check, not the sum of all of them.
func BatchCheck(urls []string) []BatchResult {
	results := make([]BatchResult, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			result, err := Check(url)
			results[i] = BatchResult{Result: result, Err: err}
		}(i, url)
	}
	wg.Wait()
	return results
}
