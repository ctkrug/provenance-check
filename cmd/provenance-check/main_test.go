package main

import (
	"os/exec"
	"strings"
	"testing"
)

// buildBinary compiles the CLI once per test run and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := t.TempDir() + "/provenance-check"
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

// Unsupported-host URLs fail during parsing, before any network call, so
// these tests stay deterministic without depending on live GitHub/Hugging
// Face access. End-to-end fetch/classify behavior is covered against a
// local test server in the internal/provenance package.
func TestStdinURLsAreCheckedAndUnsupportedOnesReportErrors(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader("https://gitlab.com/example/one\n\nhttps://bitbucket.org/example/two\n")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit when every URL is unsupported")
	}
	text := string(out)
	if !strings.Contains(text, "gitlab.com/example/one") {
		t.Errorf("expected output to mention the first URL, got: %s", text)
	}
	if !strings.Contains(text, "bitbucket.org/example/two") {
		t.Errorf("expected output to mention the second URL, got: %s", text)
	}
}

func TestArgvURLsOverrideStdin(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "https://gitlab.com/example/argv-url")
	cmd.Stdin = strings.NewReader("https://gitlab.com/example/stdin-url\n")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for an unsupported argv URL")
	}
	text := string(out)
	if !strings.Contains(text, "argv-url") {
		t.Errorf("expected output to mention the argv URL, got: %s", text)
	}
	if strings.Contains(text, "stdin-url") {
		t.Errorf("stdin should be ignored when argv URLs are given, got: %s", text)
	}
}

func TestNoArgsAndNoStdinPrintsUsage(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	cmd.Stdin = nil
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit code 2 when no URLs are provided")
	}
	if !strings.Contains(string(out), "usage: provenance-check") {
		t.Errorf("expected usage text, got: %s", out)
	}
}
