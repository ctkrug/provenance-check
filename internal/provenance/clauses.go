package provenance

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
)

//go:embed clauses.json
var clausesData []byte

// clauseDef is the on-disk shape of one clause library entry: a reviewable
// data record, not a hand-written regexp call, so a newly discovered
// phrasing is a data change (see docs/VISION.md).
type clauseDef struct {
	ID          string `json:"id"`
	Verdict     string `json:"verdict"`
	Pattern     string `json:"pattern"`
	Description string `json:"description"`
}

// compiledClause is a clauseDef with its pattern compiled and its verdict
// validated once at load time.
type compiledClause struct {
	ID          string
	Verdict     Verdict
	Description string
	re          *regexp.Regexp
}

var clauseLibrary = mustLoadClauses(clausesData)

// mustLoadClauses parses and compiles the embedded clause library. It panics
// on malformed data because that indicates a broken build, not bad input —
// the same contract as regexp.MustCompile.
func mustLoadClauses(data []byte) []compiledClause {
	clauses, err := loadClauses(data)
	if err != nil {
		panic(fmt.Sprintf("provenance: clause library failed to load: %v", err))
	}
	return clauses
}

func loadClauses(data []byte) ([]compiledClause, error) {
	var defs []clauseDef
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("parse clause library: %w", err)
	}
	compiled := make([]compiledClause, 0, len(defs))
	for _, def := range defs {
		verdict := Verdict(def.Verdict)
		if verdict != VerdictCaution && verdict != VerdictRestricted {
			return nil, fmt.Errorf("clause %q: invalid verdict %q", def.ID, def.Verdict)
		}
		re, err := regexp.Compile(def.Pattern)
		if err != nil {
			return nil, fmt.Errorf("clause %q: invalid pattern: %w", def.ID, err)
		}
		compiled = append(compiled, compiledClause{
			ID:          def.ID,
			Verdict:     verdict,
			Description: def.Description,
			re:          re,
		})
	}
	return compiled, nil
}
