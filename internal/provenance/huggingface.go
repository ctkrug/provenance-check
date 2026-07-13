package provenance

import (
	"fmt"
	"strings"
)

// huggingFaceBase is overridden in tests to point at a local httptest
// server instead of the real Hugging Face host.
var huggingFaceBase = "https://huggingface.co"

// hfBranches are tried in order when fetching raw file content; Hugging
// Face repos default to "main" but older ones may still use "master".
var hfBranches = []string{"main", "master"}

// hfLicenseAliases maps Hugging Face's lowercase license slugs (from a
// card's YAML front matter) to their canonical SPDX-ish identifier.
// Unrecognized slugs pass through verbatim rather than being dropped.
var hfLicenseAliases = map[string]string{
	"mit":                   "MIT",
	"apache-2.0":            "Apache-2.0",
	"bsd-3-clause":          "BSD-3-Clause",
	"bsd-2-clause":          "BSD-2-Clause",
	"cc-by-4.0":             "CC-BY-4.0",
	"cc-by-nc-4.0":          "CC-BY-NC-4.0",
	"cc-by-nc-sa-4.0":       "CC-BY-NC-SA-4.0",
	"openrail":              "OpenRAIL",
	"bigscience-openrail-m": "BigScience-OpenRAIL-M",
	"creativeml-openrail-m": "CreativeML-OpenRAIL-M",
	"gpl-3.0":               "GPL-3.0",
	"mpl-2.0":               "MPL-2.0",
	"unlicense":             "Unlicense",
	"isc":                   "ISC",
}

// fetchHuggingFace fetches a dataset or model card's README and, if present,
// a sibling LICENSE file. The card's YAML front-matter `license:` field, when
// present, is authoritative over text sniffing (see classify's SPDXOverride).
func fetchHuggingFace(src parsedSource) (classifyInput, error) {
	basePath := src.Repo
	if src.IsHFDataset {
		basePath = "datasets/" + src.Repo
	}

	var readmeText, branch string
	for _, candidate := range hfBranches {
		if body, ok := httpGetOK(hfRawURL(basePath, candidate, "README.md")); ok {
			readmeText = body
			branch = candidate
			break
		}
	}
	if readmeText == "" {
		return classifyInput{}, fmt.Errorf("provenance: huggingface resource %q not found", src.Repo)
	}

	licenseText, licenseName := "", ""
	for _, name := range licenseFileNames {
		if body, ok := httpGetOK(hfRawURL(basePath, branch, name)); ok {
			licenseText, licenseName = body, name
			break
		}
	}

	spdxOverride, _ := parseHFLicense(readmeText)

	return classifyInput{
		SPDXOverride:  spdxOverride,
		LicenseText:   licenseText,
		LicenseSource: firstNonEmpty(licenseName, "LICENSE"),
		ReadmeText:    readmeText,
		ReadmeSource:  "README.md",
	}, nil
}

func hfRawURL(basePath, branch, file string) string {
	return fmt.Sprintf("%s/%s/raw/%s/%s", huggingFaceBase, basePath, branch, file)
}

// parseHFLicense reads the `license:` field out of a card's leading YAML
// front matter (delimited by "---" lines) and normalizes it to a canonical
// identifier. It returns ok=false if there is no front matter or no license
// field, distinct from an empty value.
func parseHFLicense(readme string) (id string, ok bool) {
	lines := strings.Split(readme, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", false
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			return "", false
		}
		val, found := strings.CutPrefix(trimmed, "license:")
		if !found {
			continue
		}
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if val == "" {
			return "", false
		}
		return normalizeHFLicenseSlug(val), true
	}
	return "", false
}

func normalizeHFLicenseSlug(slug string) string {
	if id, ok := hfLicenseAliases[strings.ToLower(strings.TrimSpace(slug))]; ok {
		return id
	}
	return slug
}
