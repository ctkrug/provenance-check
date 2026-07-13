// Command provenance-check flags AI/ML training restrictions in dataset and
// repo licenses. See https://github.com/ctkrug/provenance-check for usage.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

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
	for _, url := range urls {
		result, err := provenance.Check(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "provenance-check: %s: %v\n", url, err)
			exitCode = 1
			continue
		}
		fmt.Printf("%-8s %s\n", result.Verdict, result.URL)
	}
	os.Exit(exitCode)
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
	fmt.Fprintln(os.Stderr, "Flags a dataset or repo license for AI/ML training restrictions.")
}
