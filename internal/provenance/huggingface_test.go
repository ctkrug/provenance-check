package provenance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withHuggingFaceTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	prev := huggingFaceBase
	huggingFaceBase = srv.URL
	t.Cleanup(func() { huggingFaceBase = prev })
}

func TestFetchHuggingFaceDatasetParsesLicenseFrontMatter(t *testing.T) {
	card := "---\nlicense: apache-2.0\ntags:\n  - text\n---\n# Example Dataset\nAn ordinary dataset."
	withHuggingFaceTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/datasets/example/no-training-dataset/raw/main/README.md" {
			_, _ = fmt.Fprint(w, card)
			return
		}
		http.NotFound(w, r)
	})

	in, err := fetchHuggingFace(parsedSource{Kind: sourceHuggingFace, Repo: "example/no-training-dataset", IsHFDataset: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.SPDXOverride != "Apache-2.0" {
		t.Errorf("SPDXOverride = %q, want Apache-2.0", in.SPDXOverride)
	}
	if !strings.Contains(in.ReadmeText, "Example Dataset") {
		t.Errorf("ReadmeText missing card body: %q", in.ReadmeText)
	}
}

func TestFetchHuggingFaceFallsBackToMasterBranch(t *testing.T) {
	withHuggingFaceTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gpt2/raw/main/README.md":
			http.NotFound(w, r)
		case "/gpt2/raw/master/README.md":
			_, _ = fmt.Fprint(w, "---\nlicense: mit\n---\n# GPT-2")
		default:
			http.NotFound(w, r)
		}
	})

	in, err := fetchHuggingFace(parsedSource{Kind: sourceHuggingFace, Repo: "gpt2", IsHFDataset: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.SPDXOverride != "MIT" {
		t.Errorf("SPDXOverride = %q, want MIT", in.SPDXOverride)
	}
}

func TestFetchHuggingFaceMissingReadmeIsAnError(t *testing.T) {
	withHuggingFaceTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, err := fetchHuggingFace(parsedSource{Kind: sourceHuggingFace, Repo: "ghost/model", IsHFDataset: false})
	if err == nil {
		t.Fatal("expected an error when the README can't be found on any branch")
	}
}

func TestFetchHuggingFaceNoFrontMatterLeavesOverrideEmpty(t *testing.T) {
	withHuggingFaceTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/example/plain-model/raw/main/README.md" {
			_, _ = fmt.Fprint(w, "# Plain Model\nNo front matter here.")
			return
		}
		http.NotFound(w, r)
	})

	in, err := fetchHuggingFace(parsedSource{Kind: sourceHuggingFace, Repo: "example/plain-model", IsHFDataset: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.SPDXOverride != "" {
		t.Errorf("SPDXOverride = %q, want empty (no front matter present)", in.SPDXOverride)
	}
}

// TestFetchHuggingFaceFetchesSiblingLicenseFile covers the case a card ships
// an explicit LICENSE file alongside its README — GitHub sources hit this
// path routinely, but no existing Hugging Face fixture served a LICENSE, so
// the fetch loop's success branch was untested.
func TestFetchHuggingFaceFetchesSiblingLicenseFile(t *testing.T) {
	withHuggingFaceTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/example/with-license/raw/main/README.md":
			_, _ = fmt.Fprint(w, "# With License\nSee LICENSE.")
		case "/example/with-license/raw/main/LICENSE":
			_, _ = fmt.Fprint(w, mitText)
		default:
			http.NotFound(w, r)
		}
	})

	in, err := fetchHuggingFace(parsedSource{Kind: sourceHuggingFace, Repo: "example/with-license", IsHFDataset: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.LicenseSource != "LICENSE" {
		t.Errorf("LicenseSource = %q, want LICENSE", in.LicenseSource)
	}
	if !strings.Contains(in.LicenseText, "MIT License") {
		t.Errorf("LicenseText = %q, want MIT boilerplate", in.LicenseText)
	}
}

func TestParseHFLicenseVariants(t *testing.T) {
	cases := []struct {
		name   string
		readme string
		wantID string
		wantOK bool
	}{
		{"quoted value", "---\nlicense: \"cc-by-nc-4.0\"\n---\nbody", "CC-BY-NC-4.0", true},
		{"unrecognized slug passes through", "---\nlicense: some-custom-license\n---\nbody", "some-custom-license", true},
		{"no front matter", "# Just a heading\nbody", "", false},
		{"front matter without license field", "---\ntags:\n  - x\n---\nbody", "", false},
		{"empty license value", "---\nlicense:\n---\nbody", "", false},
		{"front matter never closes and has no license field", "---\ntags:\n  - x\n  - y", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, ok := parseHFLicense(tc.readme)
			if ok != tc.wantOK || id != tc.wantID {
				t.Errorf("parseHFLicense(%q) = (%q, %v), want (%q, %v)", tc.readme, id, ok, tc.wantID, tc.wantOK)
			}
		})
	}
}
