package provenance

import (
	"strings"
	"testing"
)

func TestCheckReturnsErrorForUnimplementedEngine(t *testing.T) {
	_, err := Check("https://github.com/example/dataset")
	if err == nil {
		t.Fatal("expected Check to return an error before the classification engine is built")
	}
}

func TestCheckErrorMentionsTheURL(t *testing.T) {
	urls := []string{
		"https://github.com/example/dataset",
		"https://huggingface.co/datasets/example",
	}
	for _, url := range urls {
		_, err := Check(url)
		if err == nil || !strings.Contains(err.Error(), url) {
			t.Errorf("Check(%q) error = %v, want error mentioning the URL", url, err)
		}
	}
}

func TestVerdictConstants(t *testing.T) {
	for _, v := range []Verdict{VerdictClear, VerdictCaution, VerdictRestricted} {
		if v == "" {
			t.Fatal("verdict constant must not be empty")
		}
	}
}
