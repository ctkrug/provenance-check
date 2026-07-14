package provenance

import (
	"strings"
	"testing"
)

const mitText = `MIT License

Copyright (c) 2026 Example

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction...`

const apache2Text = `Apache License
Version 2.0, January 2004
http://www.apache.org/licenses/

TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION`

const bsd3Text = `Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice.
2. Redistributions in binary form must reproduce the above copyright notice.
3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived from
   this software without specific prior written permission.`

const bsd2Text = `Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice.
2. Redistributions in binary form must reproduce the above copyright notice.`

const iscText = `Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.`

const unlicenseText = `This is free and unencumbered software released into the public domain.`

const ccByNc4Text = `Creative Commons Attribution-NonCommercial 4.0 International Public License`

func TestDetectSPDXStandardLicenses(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{"MIT", mitText, "MIT"},
		{"Apache-2.0", apache2Text, "Apache-2.0"},
		{"BSD-3-Clause", bsd3Text, "BSD-3-Clause"},
		{"BSD-2-Clause", bsd2Text, "BSD-2-Clause"},
		{"ISC", iscText, "ISC"},
		{"Unlicense", unlicenseText, "Unlicense"},
		{"CC-BY-NC-4.0", ccByNc4Text, "CC-BY-NC-4.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := DetectSPDX(tc.text)
			if !ok {
				t.Fatalf("DetectSPDX(%s) ok = false, want true", tc.name)
			}
			if got != tc.want {
				t.Errorf("DetectSPDX(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestDetectSPDXUnknownTextReturnsExplicitUnknown(t *testing.T) {
	got, ok := DetectSPDX("All rights reserved. Do whatever you want, I guess.")
	if ok {
		t.Fatalf("DetectSPDX(unrecognized text) ok = true, want false (got %q)", got)
	}
	if got != "" {
		t.Errorf("DetectSPDX(unrecognized text) = %q, want empty string", got)
	}
}

// TestDetectSPDXExcludesBSD2FromNonAttributableText covers the BSD-2-Clause
// signature's Excludes clause on its own: text with both "redistributions"
// phrases but no "endorse or promote products" fails the BSD-3-Clause check,
// but must NOT then be misidentified as BSD-2-Clause just because it also
// contains "neither the name" (an incomplete/mangled BSD-3 variant, not a
// real BSD-2 grant — BSD-2 never has a "neither the name" clause at all).
func TestDetectSPDXExcludesBSD2FromNonAttributableText(t *testing.T) {
	text := `Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice.
2. Redistributions in binary form must reproduce the above copyright notice.
3. Neither the name of the copyright holder may be used to endorse this
   software without specific prior written permission.
`
	if id, ok := DetectSPDX(text); ok {
		t.Errorf("DetectSPDX(mangled BSD-3 fragment) = (%q, true), want ok=false", id)
	}
}

func TestDetectSPDXEmptyTextReturnsUnknown(t *testing.T) {
	if _, ok := DetectSPDX(""); ok {
		t.Fatal("DetectSPDX(\"\") ok = true, want false")
	}
	if _, ok := DetectSPDX("   \n\t  "); ok {
		t.Fatal("DetectSPDX(whitespace) ok = true, want false")
	}
}

func TestDetectSPDXIsCaseInsensitive(t *testing.T) {
	got, ok := DetectSPDX("permission is hereby granted, free of charge, to any person...")
	if !ok || got != "MIT" {
		t.Errorf("DetectSPDX(lowercase MIT) = (%q, %v), want (MIT, true)", got, ok)
	}
}

// FuzzDetectSPDX asserts the "explicit unknown, never a wrong guess"
// contract holds for arbitrary input, not just the hand-picked fixtures
// above: DetectSPDX must never panic, and whenever it reports ok=true the
// identifier must be one from the known signature list — never a value
// conjured from the input text itself.
func FuzzDetectSPDX(f *testing.F) {
	f.Add(mitText)
	f.Add(apache2Text)
	f.Add("")
	f.Add("   \n\t  ")
	f.Add(strings.Repeat("license ", 10000))
	f.Add("💩 not a license 中文 \x00\x01")

	known := map[string]bool{}
	for _, sig := range spdxSignatures {
		known[sig.ID] = true
	}

	f.Fuzz(func(t *testing.T, text string) {
		id, ok := DetectSPDX(text)
		if ok && !known[id] {
			t.Errorf("DetectSPDX(%q) returned unknown id %q", text, id)
		}
		if !ok && id != "" {
			t.Errorf("DetectSPDX(%q) = (%q, false), want empty id when ok=false", text, id)
		}
	})
}
