package provenance

import "strings"

// spdxSignature is one entry in the ordered detector list: if all Contains
// substrings are present in a license text (case-insensitive) and no
// exclusion substring is present, the text is identified as ID.
type spdxSignature struct {
	ID       string
	Contains []string
	Excludes []string
}

// spdxSignatures is checked in order, most-specific first, so texts that
// share boilerplate with a broader license (e.g. BSD-3-Clause containing
// everything BSD-2-Clause does, plus more) resolve to the specific one.
var spdxSignatures = []spdxSignature{
	{
		ID:       "BSD-3-Clause",
		Contains: []string{"redistributions in binary form", "neither the name", "endorse or promote products"},
	},
	{
		ID:       "BSD-2-Clause",
		Contains: []string{"redistributions in binary form", "redistributions of source code"},
		Excludes: []string{"neither the name"},
	},
	{
		ID:       "Apache-2.0",
		Contains: []string{"apache license", "version 2.0"},
	},
	{
		ID:       "MIT",
		Contains: []string{"permission is hereby granted, free of charge"},
	},
	{
		ID:       "ISC",
		Contains: []string{"permission to use, copy, modify, and/or distribute this software"},
	},
	{
		ID:       "MPL-2.0",
		Contains: []string{"mozilla public license"},
	},
	{
		ID:       "GPL-3.0",
		Contains: []string{"gnu general public license", "version 3"},
	},
	{
		ID:       "Unlicense",
		Contains: []string{"this is free and unencumbered software released into the public domain"},
	},
	{
		ID:       "CC-BY-NC-4.0",
		Contains: []string{"attribution-noncommercial 4.0 international"},
	},
	{
		ID:       "CC-BY-NC-SA-4.0",
		Contains: []string{"attribution-noncommercial-sharealike 4.0 international"},
	},
}

// DetectSPDX identifies the SPDX license identifier for a block of license
// text using a small library of distinguishing phrases from common license
// boilerplate. It returns ok=false rather than a guess when no signature
// matches, per the "explicit unknown, never a wrong guess" requirement.
func DetectSPDX(licenseText string) (id string, ok bool) {
	if strings.TrimSpace(licenseText) == "" {
		return "", false
	}
	lower := strings.ToLower(licenseText)
	for _, sig := range spdxSignatures {
		if matchesSignature(lower, sig) {
			return sig.ID, true
		}
	}
	return "", false
}

func matchesSignature(lower string, sig spdxSignature) bool {
	for _, excl := range sig.Excludes {
		if strings.Contains(lower, excl) {
			return false
		}
	}
	for _, want := range sig.Contains {
		if !strings.Contains(lower, want) {
			return false
		}
	}
	return true
}
