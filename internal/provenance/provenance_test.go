package provenance

import "testing"

func TestCheckReturnsErrorForUnimplementedEngine(t *testing.T) {
	_, err := Check("https://github.com/example/dataset")
	if err == nil {
		t.Fatal("expected Check to return an error before the classification engine is built")
	}
}

func TestVerdictConstants(t *testing.T) {
	for _, v := range []Verdict{VerdictClear, VerdictCaution, VerdictRestricted} {
		if v == "" {
			t.Fatal("verdict constant must not be empty")
		}
	}
}
