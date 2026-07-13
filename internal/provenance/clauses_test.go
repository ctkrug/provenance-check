package provenance

import "testing"

func TestClauseLibraryLoadsFromEmbeddedData(t *testing.T) {
	if len(clauseLibrary) == 0 {
		t.Fatal("clauseLibrary is empty; expected entries loaded from clauses.json")
	}
	for _, c := range clauseLibrary {
		if c.ID == "" {
			t.Error("clause entry has empty ID")
		}
		if c.re == nil {
			t.Errorf("clause %q: pattern not compiled", c.ID)
		}
		if c.Verdict != VerdictCaution && c.Verdict != VerdictRestricted {
			t.Errorf("clause %q: verdict = %q, want caution or restricted", c.ID, c.Verdict)
		}
	}
}

func TestLoadClausesRejectsInvalidVerdict(t *testing.T) {
	bad := `[{"id":"x","verdict":"clear","pattern":"x","description":"d"}]`
	if _, err := loadClauses([]byte(bad)); err == nil {
		t.Fatal("loadClauses with verdict \"clear\" should error; clear is not a clause outcome")
	}
}

func TestLoadClausesRejectsInvalidRegexp(t *testing.T) {
	bad := `[{"id":"x","verdict":"caution","pattern":"(unterminated","description":"d"}]`
	if _, err := loadClauses([]byte(bad)); err == nil {
		t.Fatal("loadClauses with an invalid regexp should error")
	}
}

func TestLoadClausesRejectsMalformedJSON(t *testing.T) {
	if _, err := loadClauses([]byte("not json")); err == nil {
		t.Fatal("loadClauses with malformed JSON should error")
	}
}

func TestLoadClausesEmptyArrayIsValid(t *testing.T) {
	clauses, err := loadClauses([]byte("[]"))
	if err != nil {
		t.Fatalf("loadClauses([]) unexpected error: %v", err)
	}
	if len(clauses) != 0 {
		t.Errorf("loadClauses([]) = %d entries, want 0", len(clauses))
	}
}
