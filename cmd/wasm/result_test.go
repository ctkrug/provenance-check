package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctkrug/provenance-check/internal/provenance"
)

func TestToCheckResultSuccess(t *testing.T) {
	result := provenance.Result{
		Verdict: provenance.VerdictRestricted,
		License: "unknown",
		Clause:  "not for use in training machine learning models",
		Source:  "LICENSE",
	}

	got := toCheckResult("https://github.com/o/r", result, nil)

	if got.Verdict != "restricted" {
		t.Errorf("Verdict = %q, want %q", got.Verdict, "restricted")
	}
	if got.Error != "" {
		t.Errorf("Error = %q, want empty on success", got.Error)
	}
	if got.Clause != result.Clause {
		t.Errorf("Clause = %q, want %q", got.Clause, result.Clause)
	}
}

func TestToCheckResultError(t *testing.T) {
	got := toCheckResult("not-a-url", provenance.Result{}, errors.New("provenance: unsupported source"))

	if got.Verdict != "error" {
		t.Errorf("Verdict = %q, want %q", got.Verdict, "error")
	}
	if got.Error == "" {
		t.Error("Error is empty, want the underlying error message")
	}
	if got.License != "" || got.Clause != "" {
		t.Errorf("expected zero-value License/Clause on error, got %+v", got)
	}
}

func TestCheckJSONUnsupportedSourceIsFastAndValid(t *testing.T) {
	data, err := checkJSON("not a url at all")
	if err != nil {
		t.Fatalf("checkJSON returned an error, want a JSON-encoded error result: %v", err)
	}
	if !strings.Contains(data, `"verdict":"error"`) {
		t.Errorf("checkJSON output = %s, want a verdict:error field", data)
	}
	if !strings.Contains(data, `"url":"not a url at all"`) {
		t.Errorf("checkJSON output = %s, want the original url echoed back", data)
	}
}

func TestCheckJSONEmptyURL(t *testing.T) {
	data, err := checkJSON("")
	if err != nil {
		t.Fatalf("checkJSON(\"\") returned an error: %v", err)
	}
	if !strings.Contains(data, `"verdict":"error"`) {
		t.Errorf("checkJSON(\"\") = %s, want verdict:error for an empty URL", data)
	}
}
