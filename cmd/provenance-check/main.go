// Command provenance-check flags AI/ML training restrictions in dataset and
// repo licenses. See https://github.com/ctkrug/provenance-check for usage.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ctkrug/provenance-check/internal/provenance"
)

func main() {
	flag.Usage = printUsage
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		urls = readURLsFromStdin()
	}
	if len(urls) == 0 {
		printUsage()
		os.Exit(2)
	}

	exitCode := 0
	for _, r := range provenance.BatchCheck(urls) {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "provenance-check: %s: %v\n", r.Result.URL, r.Err)
			exitCode = 1
			continue
		}
		printResult(r.Result)
		if r.Result.Verdict == provenance.VerdictRestricted {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

// printResult renders one line per URL with its badge and SPDX license,
// followed by an indented line quoting the flagged clause and its source
// file for any non-clear verdict.
func printResult(r provenance.Result) {
	fmt.Printf("%-10s %-16s %s\n", strings.ToUpper(string(r.Verdict)), r.License, r.URL)
	if r.Clause != "" {
		fmt.Printf("           clause: %q (%s)\n", r.Clause, r.Source)
	}
}

func readURLsFromStdin() []string {
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return nil // no piped input
	}
	var urls []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			urls = append(urls, line)
		}
	}
	return urls
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: provenance-check <url> [<url> ...]")
	fmt.Fprintln(os.Stderr, "       cat urls.txt | provenance-check")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Flags a GitHub repo or Hugging Face dataset/model license for AI/ML")
	fmt.Fprintln(os.Stderr, "training restrictions. URLs are checked concurrently.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Exits non-zero if any URL is restricted or fails to resolve.")
}
