package main

import (
	"errors"
	"testing"

	"github.com/ctkrug/provenance-check/internal/provenance"
)

func TestExitCodeAllClearIsZero(t *testing.T) {
	results := []provenance.BatchResult{
		{Result: provenance.Result{Verdict: provenance.VerdictClear}},
		{Result: provenance.Result{Verdict: provenance.VerdictCaution}},
	}
	if got := exitCode(results); got != 0 {
		t.Errorf("exitCode(clear+caution) = %d, want 0", got)
	}
}

func TestExitCodeRestrictedIsNonZero(t *testing.T) {
	results := []provenance.BatchResult{
		{Result: provenance.Result{Verdict: provenance.VerdictClear}},
		{Result: provenance.Result{Verdict: provenance.VerdictRestricted}},
	}
	if got := exitCode(results); got == 0 {
		t.Error("exitCode with a restricted result = 0, want non-zero")
	}
}

func TestExitCodeUnresolvedErrorIsNonZero(t *testing.T) {
	results := []provenance.BatchResult{
		{Result: provenance.Result{Verdict: provenance.VerdictClear}},
		{Err: errors.New("provenance: unsupported source")},
	}
	if got := exitCode(results); got == 0 {
		t.Error("exitCode with an unresolved URL = 0, want non-zero")
	}
}

func TestExitCodeEmptyResultsIsZero(t *testing.T) {
	if got := exitCode(nil); got != 0 {
		t.Errorf("exitCode(nil) = %d, want 0", got)
	}
}
