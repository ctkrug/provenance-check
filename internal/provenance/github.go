package provenance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// githubAPIBase and githubRawBase are overridden in tests to point at a
// local httptest server instead of the real GitHub hosts.
var (
	githubAPIBase = "https://api.github.com"
	githubRawBase = "https://raw.githubusercontent.com"
)

// httpClient is shared by every fetcher; a single client reuses connections
// across the concurrent fetches BatchCheck issues.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// licenseFileNames and readmeFileNames are tried in order; the first one
// that resolves with HTTP 200 wins. Order reflects real-world frequency.
var (
	licenseFileNames = []string{"LICENSE", "LICENSE.md", "LICENSE.txt"}
	readmeFileNames  = []string{"README.md", "README", "readme.md"}
)

type githubRepoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

// fetchGitHub resolves a repo's default branch, then fetches its LICENSE
// and README from raw.githubusercontent.com. A missing LICENSE or README is
// not an error (that's a legitimate "unknown" classification outcome) — only
// an unresolvable repo is.
func fetchGitHub(src parsedSource) (classifyInput, error) {
	branch, err := githubDefaultBranch(src.Owner, src.Repo)
	if err != nil {
		return classifyInput{}, err
	}

	licenseText, licenseName := fetchFirstMatch(githubRawBase, src.Owner, src.Repo, branch, licenseFileNames)
	readmeText, readmeName := fetchFirstMatch(githubRawBase, src.Owner, src.Repo, branch, readmeFileNames)

	return classifyInput{
		LicenseText:   licenseText,
		LicenseSource: firstNonEmpty(licenseName, "LICENSE"),
		ReadmeText:    readmeText,
		ReadmeSource:  firstNonEmpty(readmeName, "README.md"),
	}, nil
}

func githubDefaultBranch(owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", githubAPIBase, owner, repo)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("provenance: fetch github repo %s/%s: %w", owner, repo, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("provenance: github repo %s/%s not found", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("provenance: github repo %s/%s: unexpected status %d", owner, repo, resp.StatusCode)
	}

	var info githubRepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("provenance: github repo %s/%s: decode response: %w", owner, repo, err)
	}
	if info.DefaultBranch == "" {
		return "", fmt.Errorf("provenance: github repo %s/%s: no default branch reported", owner, repo)
	}
	return info.DefaultBranch, nil
}

// fetchFirstMatch tries each candidate file name in order and returns the
// body and name of the first that resolves with HTTP 200, or ("", "") if
// none do.
func fetchFirstMatch(rawBase, owner, repo, branch string, names []string) (text, matchedName string) {
	for _, name := range names {
		url := fmt.Sprintf("%s/%s/%s/%s/%s", rawBase, owner, repo, branch, name)
		if body, ok := httpGetOK(url); ok {
			return body, name
		}
	}
	return "", ""
}

func httpGetOK(url string) (string, bool) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}
	return string(body), true
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
