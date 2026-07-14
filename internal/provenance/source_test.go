package provenance

import (
	"strings"
	"testing"
)

func TestParseSourceGitHubRepo(t *testing.T) {
	src, err := parseSource("https://github.com/example/dataset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Kind != sourceGitHub || src.Owner != "example" || src.Repo != "dataset" {
		t.Errorf("got %+v, want github example/dataset", src)
	}
}

func TestParseSourceGitHubRepoWithTrailingPathAndDotGit(t *testing.T) {
	src, err := parseSource("https://github.com/example/dataset.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Repo != "dataset" {
		t.Errorf("Repo = %q, want dataset (without .git suffix)", src.Repo)
	}

	src2, err := parseSource("https://github.com/example/dataset/tree/main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src2.Owner != "example" || src2.Repo != "dataset" {
		t.Errorf("got %+v, want example/dataset ignoring trailing /tree/main", src2)
	}
}

func TestParseSourceHuggingFaceDataset(t *testing.T) {
	src, err := parseSource("https://huggingface.co/datasets/example/no-training-dataset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Kind != sourceHuggingFace || !src.IsHFDataset || src.Repo != "example/no-training-dataset" {
		t.Errorf("got %+v, want HF dataset example/no-training-dataset", src)
	}
}

func TestParseSourceHuggingFaceModelWithOrg(t *testing.T) {
	src, err := parseSource("https://huggingface.co/meta-llama/Llama-3-8B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Kind != sourceHuggingFace || src.IsHFDataset || src.Repo != "meta-llama/Llama-3-8B" {
		t.Errorf("got %+v, want HF model meta-llama/Llama-3-8B", src)
	}
}

func TestParseSourceHuggingFaceModelWithoutOrg(t *testing.T) {
	src, err := parseSource("https://huggingface.co/gpt2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Repo != "gpt2" {
		t.Errorf("Repo = %q, want gpt2", src.Repo)
	}
}

func TestParseSourceHuggingFaceStopsAtUIRouteKeyword(t *testing.T) {
	src, err := parseSource("https://huggingface.co/datasets/example/card/blob/main/README.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Repo != "example/card" {
		t.Errorf("Repo = %q, want example/card (stopping before /blob/...)", src.Repo)
	}
}

func TestParseSourceUnsupportedHost(t *testing.T) {
	_, err := parseSource("https://gitlab.com/example/dataset")
	if err == nil {
		t.Fatal("expected an error for a non-GitHub/HF host")
	}
}

func TestParseSourceMalformedURL(t *testing.T) {
	cases := []string{
		"not a url at all",
		"",
		"github.com/missing-scheme/repo",
		"https://github.com/",
		"https://github.com/owner-only",
	}
	for _, raw := range cases {
		if _, err := parseSource(raw); err == nil {
			t.Errorf("parseSource(%q): expected an error, got none", raw)
		}
	}
}

func TestParseSourceUnsupportedErrorMentionsURL(t *testing.T) {
	raw := "https://gitlab.com/example/dataset"
	_, err := parseSource(raw)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !strings.Contains(got, raw) {
		t.Errorf("error %q does not mention the offending URL %q", got, raw)
	}
}

// FuzzParseSource asserts parseSource's core promise for arbitrary pasted
// text, not just the hand-picked fixtures above: it must never panic, and
// it must never return ok with an empty Owner/Repo identity (a "successful"
// parse with nothing to fetch would just surface as a confusing empty-body
// result three layers downstream instead of a clear "unsupported source").
func FuzzParseSource(f *testing.F) {
	seeds := []string{
		"https://github.com/example/dataset",
		"https://huggingface.co/datasets/example/name",
		"https://huggingface.co/gpt2",
		"not a url at all",
		"",
		"github.com/missing-scheme/repo",
		"https://github.com/../../etc/passwd",
		"https://github.com/%2e%2e/%2e%2e",
		"ftp://github.com/owner/repo",
		"https://github.com:99999/owner/repo",
		"https://xn--n3h.github.com/owner/repo",
		"https://github.com/💩/repo",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		src, err := parseSource(raw)
		if err != nil {
			return
		}
		if src.Kind == sourceGitHub && (src.Owner == "" || src.Repo == "") {
			t.Errorf("parseSource(%q) returned ok github source with empty identity: %+v", raw, src)
		}
		if src.Kind == sourceHuggingFace && src.Repo == "" {
			t.Errorf("parseSource(%q) returned ok huggingface source with empty identity: %+v", raw, src)
		}
	})
}
