package provenance

import "strings"

// ccByNCClause is the synthesized clause text shown when a CC-BY-NC family
// license is detected but no explicit AI-training clause was found. The
// license itself is the signal: NonCommercial terms are ambiguous for
// commercial ML training use even without a bespoke addendum.
const ccByNCClause = "License is a Creative Commons NonCommercial variant; " +
	"commercial ML training use is ambiguous under NonCommercial terms."

// classifyInput is everything classify needs to reach a verdict for one
// checked URL, gathered by a source-specific fetcher (github.go,
// huggingface.go).
type classifyInput struct {
	// SPDXOverride, when non-empty, is an authoritative SPDX-ish identifier
	// supplied directly by the source (e.g. a Hugging Face card's `license:`
	// front-matter field) rather than sniffed from license text.
	SPDXOverride string
	// LicenseText is the raw LICENSE file content, if any was found.
	LicenseText string
	// LicenseSource names the document LicenseText came from, e.g. "LICENSE".
	LicenseSource string
	// ReadmeText is the raw README / model-card content, if any was found.
	ReadmeText string
	// ReadmeSource names the document ReadmeText came from, e.g. "README.md".
	ReadmeSource string
}

// classify is the single place SPDX detection and clause scanning combine
// into a Result. Both the GitHub and Hugging Face code paths funnel through
// here so the two source types never risk divergent matching logic.
func classify(in classifyInput) Result {
	spdxID := in.SPDXOverride
	known := spdxID != ""
	if !known {
		if id, ok := DetectSPDX(in.LicenseText); ok {
			spdxID = id
			known = true
		}
	}
	license := spdxID
	if !known {
		license = "unknown"
	}

	if match, ok := scanDocuments(
		document{Text: in.LicenseText, Source: in.LicenseSource},
		document{Text: in.ReadmeText, Source: in.ReadmeSource},
	); ok {
		return Result{
			Verdict: match.Verdict,
			License: license,
			Clause:  match.Text,
			Source:  match.Source,
		}
	}

	if strings.HasPrefix(spdxID, "CC-BY-NC") {
		return Result{
			Verdict: VerdictCaution,
			License: license,
			Clause:  ccByNCClause,
			Source:  in.LicenseSource,
		}
	}

	return Result{Verdict: VerdictClear, License: license}
}
