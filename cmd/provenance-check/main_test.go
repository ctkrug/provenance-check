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

func TestStdinURLsAreChecked(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader("https://github.com/example/dataset\n\nhttps://huggingface.co/datasets/example\n")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit while the classification engine is unimplemented")
	}
	text := string(out)
	if !strings.Contains(text, "github.com/example/dataset") {
		t.Errorf("expected output to mention the first URL, got: %s", text)
	}
	if !strings.Contains(text, "huggingface.co/datasets/example") {
		t.Errorf("expected output to mention the second URL, got: %s", text)
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
