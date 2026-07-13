package provenance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckUnsupportedURLReturnsErrorMentioningIt(t *testing.T) {
	url := "https://gitlab.com/example/dataset"
	_, err := Check(url)
	if err == nil || !strings.Contains(err.Error(), url) {
		t.Errorf("Check(%q) error = %v, want error mentioning the URL", url, err)
	}
}

func TestVerdictConstants(t *testing.T) {
	for _, v := range []Verdict{VerdictClear, VerdictCaution, VerdictRestricted} {
		if v == "" {
			t.Fatal("verdict constant must not be empty")
		}
	}
}

// withCombinedTestServer serves both GitHub and Hugging Face endpoints from
// one httptest server, letting a single test exercise Check end-to-end for
// either source without live network access.
func withCombinedTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	prevAPI, prevRaw, prevHF := githubAPIBase, githubRawBase, huggingFaceBase
	githubAPIBase, githubRawBase, huggingFaceBase = srv.URL, srv.URL, srv.URL
	t.Cleanup(func() {
		githubAPIBase, githubRawBase, huggingFaceBase = prevAPI, prevRaw, prevHF
	})
}

func TestCheckEndToEndGitHubClearVerdict(t *testing.T) {
	withCombinedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/example/permissive-repo":
			_, _ = fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/example/permissive-repo/main/LICENSE":
			_, _ = fmt.Fprint(w, mitText)
		case "/example/permissive-repo/main/README.md":
			_, _ = fmt.Fprint(w, "# Permissive Repo\nNothing unusual here.")
		default:
			http.NotFound(w, r)
		}
	})

	result, err := Check("https://github.com/example/permissive-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictClear {
		t.Errorf("verdict = %q, want clear", result.Verdict)
	}
	if result.License != "MIT" {
		t.Errorf("license = %q, want MIT", result.License)
	}
	if result.URL != "https://github.com/example/permissive-repo" {
		t.Errorf("URL = %q, want the original checked URL", result.URL)
	}
}

func TestCheckEndToEndGitHubRestrictedVerdict(t *testing.T) {
	withCombinedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/example/restricted-dataset":
			_, _ = fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/example/restricted-dataset/main/README.md":
			_, _ = fmt.Fprint(w, "You are not permitted to use this dataset for AI training purposes.")
		default:
			http.NotFound(w, r)
		}
	})

	result, err := Check("https://github.com/example/restricted-dataset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictRestricted {
		t.Errorf("verdict = %q, want restricted", result.Verdict)
	}
	if result.Clause == "" {
		t.Error("expected the matched clause to be quoted")
	}
}

func TestCheckEndToEndHuggingFaceDataset(t *testing.T) {
	withCombinedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/datasets/example/no-training-dataset/raw/main/README.md" {
			_, _ = fmt.Fprint(w, "---\nlicense: cc-by-nc-4.0\n---\n# Example Dataset")
			return
		}
		http.NotFound(w, r)
	})

	result, err := Check("https://huggingface.co/datasets/example/no-training-dataset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictCaution {
		t.Errorf("verdict = %q, want caution (CC-BY-NC ambiguity)", result.Verdict)
	}
	if result.License != "CC-BY-NC-4.0" {
		t.Errorf("license = %q, want CC-BY-NC-4.0", result.License)
	}
}

func TestCheckGitHubRepoNotFoundPropagatesError(t *testing.T) {
	withCombinedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, err := Check("https://github.com/ghost/repo")
	if err == nil {
		t.Fatal("expected an error for a repo that doesn't resolve")
	}
}
