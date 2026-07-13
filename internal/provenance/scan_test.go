package provenance

import (
	"strings"
	"testing"
)

// Each fixture below is verbatim-ish real-world phrasing for one clause
// pattern in clauses.json (docs/BACKLOG.md requires at least 5 such
// fixtures with the matched sentence captured verbatim).
var clauseFixtures = []struct {
	clauseID string
	verdict  Verdict
	text     string
	wantSub  string // substring the matched clause must contain
}{
	{
		clauseID: "explicit-no-ml-training",
		verdict:  VerdictRestricted,
		text:     "This dataset may not be used to train machine learning models without written permission.",
		wantSub:  "may not be used to train",
	},
	{
		clauseID: "no-ai-training-addendum",
		verdict:  VerdictRestricted,
		text:     "You are not permitted to use this dataset for AI training purposes of any kind.",
		wantSub:  "not permitted to use this dataset for AI training",
	},
	{
		clauseID: "no-training-data-for-generative-models",
		verdict:  VerdictRestricted,
		text:     "This content may not be used as training data for any generative model without prior authorization.",
		wantSub:  "may not be used as training data",
	},
	{
		clauseID: "openrail-use-restrictions",
		verdict:  VerdictRestricted,
		text:     "By downloading this model you agree not to use the Model or Derivatives of the Model for any restricted purpose.",
		wantSub:  "you agree not to use the Model or Derivatives of the Model",
	},
	{
		clauseID: "responsible-ai-license-restricted-uses",
		verdict:  VerdictRestricted,
		text:     "This License contains restrictions on use as set out in the Section titled Restricted Uses of this Agreement.",
		wantSub:  "restrictions on use",
	},
	{
		clauseID: "no-commercial-ml-use",
		verdict:  VerdictCaution,
		text:     "This model is not intended for commercial use without a separate license agreement.",
		wantSub:  "not intended for commercial use",
	},
	{
		clauseID: "research-only-use",
		verdict:  VerdictCaution,
		text:     "This dataset is licensed for research purposes only and may not be redistributed.",
		wantSub:  "licensed for research purposes only",
	},
}

func TestScanTextMatchesKnownPhrasings(t *testing.T) {
	for _, f := range clauseFixtures {
		t.Run(f.clauseID, func(t *testing.T) {
			match, ok := scanText(f.text, "README.md")
			if !ok {
				t.Fatalf("scanText did not match fixture for clause %q", f.clauseID)
			}
			if match.Verdict != f.verdict {
				t.Errorf("verdict = %q, want %q", match.Verdict, f.verdict)
			}
			if match.Source != "README.md" {
				t.Errorf("source = %q, want README.md", match.Source)
			}
			if !strings.Contains(match.Text, f.wantSub) {
				t.Errorf("matched clause = %q, want substring %q", match.Text, f.wantSub)
			}
		})
	}
}

func TestScanTextNoMatchOnCleanText(t *testing.T) {
	_, ok := scanText("This project is released under the MIT License. Enjoy!", "LICENSE")
	if ok {
		t.Fatal("scanText matched a clean MIT license text; want no match")
	}
}

func TestScanTextEmptyInput(t *testing.T) {
	if _, ok := scanText("", "LICENSE"); ok {
		t.Fatal("scanText(\"\") matched; want no match")
	}
}

func TestScanDocumentsPicksStrongestAcrossDocuments(t *testing.T) {
	license := "This model is not intended for commercial use without a separate license."
	readme := "You are not permitted to use this dataset for AI training purposes."
	match, ok := scanDocuments(
		document{Text: license, Source: "LICENSE"},
		document{Text: readme, Source: "README.md"},
	)
	if !ok {
		t.Fatal("expected a match across documents")
	}
	if match.Verdict != VerdictRestricted {
		t.Errorf("verdict = %q, want restricted (readme's clause should win over license's caution)", match.Verdict)
	}
	if match.Source != "README.md" {
		t.Errorf("source = %q, want README.md", match.Source)
	}
}

func TestScanDocumentsSkipsBlankDocuments(t *testing.T) {
	match, ok := scanDocuments(
		document{Text: "", Source: "LICENSE"},
		document{Text: "   \n  ", Source: "README.md"},
	)
	if ok {
		t.Fatalf("expected no match against blank documents, got %+v", match)
	}
}

func TestScanDocumentsNoMatchesReturnsFalse(t *testing.T) {
	_, ok := scanDocuments(document{Text: "MIT License, no restrictions here.", Source: "LICENSE"})
	if ok {
		t.Fatal("expected no match on clean license text")
	}
}
