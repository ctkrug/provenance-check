package provenance

import (
	"fmt"
	"net/url"
	"strings"
)

// sourceKind identifies which fetcher a parsed URL belongs to.
type sourceKind int

const (
	sourceGitHub sourceKind = iota
	sourceHuggingFace
)

// parsedSource is a URL resolved to a concrete, fetchable identity. It
// carries no network state — parsing is pure and independently testable
// from fetching.
type parsedSource struct {
	Kind        sourceKind
	Owner       string // GitHub owner; empty for Hugging Face
	Repo        string // GitHub repo name, or Hugging Face namespace/name path
	IsHFDataset bool   // true for huggingface.co/datasets/..., false for a model repo
}

// hfPathKeywords are Hugging Face URL segments that mark the end of a
// repo's identity path (e.g. huggingface.co/org/model/tree/main), so extra
// path segments beyond the identity don't get folded into the repo name.
var hfPathKeywords = map[string]bool{
	"tree": true, "blob": true, "resolve": true, "raw": true,
	"commit": true, "discussions": true,
}

// ParseSource classifies a pasted URL as a GitHub repo, a Hugging Face
// dataset/model, or unsupported. Unsupported input (a non-GitHub/HF host, a
// malformed URL, or a URL with no repo path) always returns a descriptive
// error rather than panicking or silently skipping the URL.
func parseSource(rawURL string) (parsedSource, error) {
	trimmed := strings.TrimSpace(rawURL)
	u, err := url.Parse(trimmed)
	if err != nil {
		return parsedSource{}, fmt.Errorf("provenance: unsupported source: %q", rawURL)
	}
	host := strings.ToLower(u.Host)
	segments := pathSegments(u.Path)

	switch host {
	case "github.com", "www.github.com":
		if len(segments) < 2 || segments[0] == "" || segments[1] == "" {
			return parsedSource{}, fmt.Errorf("provenance: unsupported source: %q", rawURL)
		}
		owner := segments[0]
		repo := strings.TrimSuffix(segments[1], ".git")
		return parsedSource{Kind: sourceGitHub, Owner: owner, Repo: repo}, nil

	case "huggingface.co", "www.huggingface.co":
		if len(segments) >= 1 && segments[0] == "datasets" {
			name := hfIdentityPath(segments[1:])
			if name == "" {
				return parsedSource{}, fmt.Errorf("provenance: unsupported source: %q", rawURL)
			}
			return parsedSource{Kind: sourceHuggingFace, Repo: name, IsHFDataset: true}, nil
		}
		name := hfIdentityPath(segments)
		if name == "" {
			return parsedSource{}, fmt.Errorf("provenance: unsupported source: %q", rawURL)
		}
		return parsedSource{Kind: sourceHuggingFace, Repo: name, IsHFDataset: false}, nil

	default:
		return parsedSource{}, fmt.Errorf("provenance: unsupported source: %q", rawURL)
	}
}

func pathSegments(p string) []string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

// hfIdentityPath takes at most the first two non-keyword path segments
// (namespace/name, or just name for org-less models like "gpt2") as a
// repo's identity, stopping at the first segment that marks a Hugging Face
// UI route (tree, blob, resolve, ...).
func hfIdentityPath(segments []string) string {
	var identity []string
	for _, s := range segments {
		if s == "" || hfPathKeywords[s] {
			break
		}
		identity = append(identity, s)
		if len(identity) == 2 {
			break
		}
	}
	return strings.Join(identity, "/")
}
