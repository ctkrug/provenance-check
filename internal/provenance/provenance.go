// Package provenance is the core license-clause classification engine shared
// by the CLI and the web front end. It answers one question per URL: can you
// train on this?
package provenance

import "fmt"

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

// Check fetches and classifies a single dataset or repo URL. The
// classification engine (SPDX detection + non-standard clause matching) is
// not implemented yet; this stub exists so the CLI and its tests have a
// stable entrypoint to build against.
func Check(url string) (Result, error) {
	return Result{}, fmt.Errorf("provenance: Check not yet implemented for %q", url)
}
