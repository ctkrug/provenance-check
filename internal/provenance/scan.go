package provenance

import "strings"

// clauseMatch is one clause-library hit against a piece of source text.
type clauseMatch struct {
	Verdict Verdict
	Text    string // the exact matched clause, verbatim from the source
	Source  string // which document it came from, e.g. "LICENSE" or "README.md"
}

// severity ranks verdicts so the strongest signal across every scanned
// document wins, restricted always beating caution.
func severity(v Verdict) int {
	switch v {
	case VerdictRestricted:
		return 2
	case VerdictCaution:
		return 1
	default:
		return 0
	}
}

// scanText runs the full clause library against a single document and
// returns its strongest match, if any.
func scanText(text, source string) (clauseMatch, bool) {
	var best clauseMatch
	found := false
	for _, c := range clauseLibrary {
		loc := c.re.FindStringIndex(text)
		if loc == nil {
			continue
		}
		match := clauseMatch{
			Verdict: c.Verdict,
			Text:    strings.TrimSpace(text[loc[0]:loc[1]]),
			Source:  source,
		}
		if !found || severity(match.Verdict) > severity(best.Verdict) {
			best = match
			found = true
		}
	}
	return best, found
}

// document is one named piece of source text to scan, e.g. a LICENSE file
// or a README.
type document struct {
	Text   string
	Source string
}

// scanDocuments runs the clause library over every named document and
// returns the single strongest match across all of them. Documents are
// scanned in the order given; ties in severity keep the earlier document's
// match, so callers should list the most authoritative source (e.g. LICENSE)
// first.
func scanDocuments(docs ...document) (clauseMatch, bool) {
	var best clauseMatch
	found := false
	for _, doc := range docs {
		if strings.TrimSpace(doc.Text) == "" {
			continue
		}
		m, ok := scanText(doc.Text, doc.Source)
		if !ok {
			continue
		}
		if !found || severity(m.Verdict) > severity(best.Verdict) {
			best = m
			found = true
		}
	}
	return best, found
}
