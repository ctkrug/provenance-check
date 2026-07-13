package provenance

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestBatchCheckPreservesOrderAndIsolatesErrors(t *testing.T) {
	withCombinedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/example/good-repo":
			fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/example/good-repo/main/LICENSE":
			fmt.Fprint(w, mitText)
		default:
			http.NotFound(w, r)
		}
	})

	urls := []string{
		"https://github.com/example/good-repo",
		"https://gitlab.com/unsupported/host",
		"https://github.com/example/good-repo",
	}
	results := BatchCheck(urls)

	if len(results) != len(urls) {
		t.Fatalf("got %d results, want %d", len(results), len(urls))
	}
	if results[0].Err != nil {
		t.Errorf("results[0].Err = %v, want nil", results[0].Err)
	}
	if results[0].Result.Verdict != VerdictClear {
		t.Errorf("results[0].Result.Verdict = %q, want clear", results[0].Result.Verdict)
	}
	if results[1].Err == nil {
		t.Error("results[1].Err = nil, want an unsupported-source error")
	}
	if results[2].Err != nil {
		t.Errorf("results[2].Err = %v, want nil (one failed URL must not affect its neighbors)", results[2].Err)
	}
}

func TestBatchCheckRunsConcurrently(t *testing.T) {
	const delay = 100 * time.Millisecond
	const n = 5

	withCombinedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		switch r.URL.Path {
		case "/repos/example/slow-repo":
			fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/example/slow-repo/main/LICENSE":
			fmt.Fprint(w, mitText)
		case "/example/slow-repo/main/README.md":
			fmt.Fprint(w, "# Slow Repo")
		default:
			http.NotFound(w, r)
		}
	})

	urls := make([]string, n)
	for i := range urls {
		urls[i] = "https://github.com/example/slow-repo"
	}

	start := time.Now()
	results := BatchCheck(urls)
	elapsed := time.Since(start)

	for i, r := range results {
		if r.Err != nil {
			t.Fatalf("results[%d].Err = %v, want nil", i, r.Err)
		}
	}
	// Each URL issues 3 sequential requests (branch lookup, LICENSE,
	// README), so a serial batch of n URLs would take roughly n*3*delay
	// (1.5s here). A concurrent batch should track just the slowest single
	// check's ~3*delay, not the sum across all n.
	budget := delay * 6
	if elapsed > budget {
		t.Errorf("BatchCheck took %v for %d URLs with a %v per-request delay; want under %v (looks serial, not concurrent)", elapsed, n, delay, budget)
	}
}

func TestBatchCheckEmptyInput(t *testing.T) {
	results := BatchCheck(nil)
	if len(results) != 0 {
		t.Errorf("BatchCheck(nil) = %d results, want 0", len(results))
	}
}
