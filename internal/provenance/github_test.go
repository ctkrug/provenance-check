package provenance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// withGitHubTestServer points githubAPIBase and githubRawBase at a local
// httptest server for the duration of the test, then restores them.
func withGitHubTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	prevAPI, prevRaw := githubAPIBase, githubRawBase
	githubAPIBase = srv.URL
	githubRawBase = srv.URL
	t.Cleanup(func() {
		githubAPIBase = prevAPI
		githubRawBase = prevRaw
	})
}

func TestFetchGitHubResolvesDefaultBranchAndFiles(t *testing.T) {
	withGitHubTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/example/repo":
			_, _ = fmt.Fprint(w, `{"default_branch":"trunk"}`)
		case r.URL.Path == "/example/repo/trunk/LICENSE":
			_, _ = fmt.Fprint(w, mitText)
		case r.URL.Path == "/example/repo/trunk/README.md":
			_, _ = fmt.Fprint(w, "# Example\nAn ordinary project.")
		default:
			http.NotFound(w, r)
		}
	})

	in, err := fetchGitHub(parsedSource{Kind: sourceGitHub, Owner: "example", Repo: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(in.LicenseText, "MIT License") {
		t.Errorf("LicenseText = %q, want MIT boilerplate", in.LicenseText)
	}
	if in.LicenseSource != "LICENSE" {
		t.Errorf("LicenseSource = %q, want LICENSE", in.LicenseSource)
	}
	if !strings.Contains(in.ReadmeText, "ordinary project") {
		t.Errorf("ReadmeText = %q, want the fetched README body", in.ReadmeText)
	}
}

func TestFetchGitHubTriesAlternateLicenseFileNames(t *testing.T) {
	withGitHubTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/example/repo":
			_, _ = fmt.Fprint(w, `{"default_branch":"main"}`)
		case r.URL.Path == "/example/repo/main/LICENSE.md":
			_, _ = fmt.Fprint(w, apache2Text)
		default:
			http.NotFound(w, r)
		}
	})

	in, err := fetchGitHub(parsedSource{Kind: sourceGitHub, Owner: "example", Repo: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.LicenseSource != "LICENSE.md" {
		t.Errorf("LicenseSource = %q, want LICENSE.md (LICENSE and LICENSE.txt are 404)", in.LicenseSource)
	}
}

func TestFetchGitHubMissingLicenseAndReadmeIsNotAnError(t *testing.T) {
	withGitHubTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/example/repo" {
			_, _ = fmt.Fprint(w, `{"default_branch":"main"}`)
			return
		}
		http.NotFound(w, r)
	})

	in, err := fetchGitHub(parsedSource{Kind: sourceGitHub, Owner: "example", Repo: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.LicenseText != "" || in.ReadmeText != "" {
		t.Errorf("expected empty LicenseText/ReadmeText when neither file exists, got %+v", in)
	}
}

func TestFetchGitHubRepoNotFoundIsAnError(t *testing.T) {
	withGitHubTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, err := fetchGitHub(parsedSource{Kind: sourceGitHub, Owner: "ghost", Repo: "repo"})
	if err == nil {
		t.Fatal("expected an error for a repo the API reports 404 on")
	}
}
