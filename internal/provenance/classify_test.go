package provenance

import "testing"

func TestClassifyClearOnPlainMITLicense(t *testing.T) {
	result := classify(classifyInput{
		LicenseText:   mitText,
		LicenseSource: "LICENSE",
		ReadmeText:    "A perfectly ordinary open-source project.",
		ReadmeSource:  "README.md",
	})
	if result.Verdict != VerdictClear {
		t.Errorf("verdict = %q, want clear (no false positive on a plain MIT license)", result.Verdict)
	}
	if result.License != "MIT" {
		t.Errorf("license = %q, want MIT", result.License)
	}
	if result.Clause != "" {
		t.Errorf("clause = %q, want empty for a clear verdict", result.Clause)
	}
}

func TestClassifyUnknownLicenseWithNoClauseIsClear(t *testing.T) {
	result := classify(classifyInput{
		LicenseText:   "All rights reserved.",
		LicenseSource: "LICENSE",
	})
	if result.License != "unknown" {
		t.Errorf("license = %q, want unknown", result.License)
	}
	if result.Verdict != VerdictClear {
		t.Errorf("verdict = %q, want clear (unrecognized license text alone is not a restriction)", result.Verdict)
	}
}

func TestClassifyRestrictedFromReadmeClause(t *testing.T) {
	result := classify(classifyInput{
		LicenseText:   mitText,
		LicenseSource: "LICENSE",
		ReadmeText:    "You are not permitted to use this dataset for AI training purposes.",
		ReadmeSource:  "README.md",
	})
	if result.Verdict != VerdictRestricted {
		t.Errorf("verdict = %q, want restricted", result.Verdict)
	}
	if result.Source != "README.md" {
		t.Errorf("source = %q, want README.md", result.Source)
	}
	if result.Clause == "" {
		t.Error("expected the matched clause to be quoted, got empty string")
	}
	if result.License != "MIT" {
		t.Errorf("license = %q, want MIT (the SPDX id is still reported even under a restricted verdict)", result.License)
	}
}

func TestClassifyCautionFromCCByNCLicenseAlone(t *testing.T) {
	result := classify(classifyInput{
		SPDXOverride:  "CC-BY-NC-4.0",
		LicenseSource: "LICENSE",
	})
	if result.Verdict != VerdictCaution {
		t.Errorf("verdict = %q, want caution (CC-BY-NC is ambiguous for commercial ML use)", result.Verdict)
	}
	if result.Clause == "" {
		t.Error("expected a synthesized clause explaining the NonCommercial ambiguity")
	}
}

func TestClassifyClauseMatchOutranksCCByNCAmbiguity(t *testing.T) {
	result := classify(classifyInput{
		SPDXOverride:  "CC-BY-NC-4.0",
		LicenseSource: "LICENSE",
		ReadmeText:    "You are not permitted to use this dataset for AI training purposes.",
		ReadmeSource:  "README.md",
	})
	if result.Verdict != VerdictRestricted {
		t.Errorf("verdict = %q, want restricted (an explicit clause is stronger than the CC-BY-NC heuristic)", result.Verdict)
	}
}

func TestClassifySPDXOverrideTakesPrecedenceOverTextSniffing(t *testing.T) {
	result := classify(classifyInput{
		SPDXOverride:  "Apache-2.0",
		LicenseText:   mitText, // deliberately contradicts the override
		LicenseSource: "LICENSE",
	})
	if result.License != "Apache-2.0" {
		t.Errorf("license = %q, want Apache-2.0 (source-supplied override wins over text sniffing)", result.License)
	}
}
