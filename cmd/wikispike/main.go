// Command wikispike is an EXPERIMENTAL, THROWAWAY spike (cycle C4a). It is not
// part of the noggin TUI, is not reachable from it, and must never be imported
// by cmd/noggin or internal/. It exists only to answer one question: can a
// Wikipedia article be turned into markdown worth reading inside noggin?
//
// Usage:
//
//	wikispike <article title> <output path>
//
// It takes an exact article title (no search, no completion), fetches the
// article from Wikimedia's public REST API, strips everything that is not the
// article, converts the remainder to markdown, and writes it to the output
// path. Expect it to be deleted or gutted when the cycle closes.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usage = `wikispike — EXPERIMENTAL Wikipedia-to-markdown spike (not part of noggin)

usage: wikispike <article title> <output path>

  <article title>  exact Wikipedia article title, e.g. "Major Arcana"
  <output path>    file to write the converted markdown to
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "wikispike:", err)
		os.Exit(1)
	}
}

// run is the whole command, kept out of main so it can be driven by a test.
// Anything printed for the operator goes to out; failures come back as errors.
func run(args []string, out io.Writer) error {
	if len(args) != 2 {
		fmt.Fprint(out, usage)
		return errors.New("expected exactly two arguments: <article title> <output path>")
	}
	title, outPath := args[0], args[1]
	if title == "" {
		return errors.New("article title is empty")
	}
	if outPath == "" {
		return errors.New("output path is empty")
	}

	markdown, err := pipeline(title)
	if err != nil {
		return err
	}

	// Convenience for writing samples into a scratch root that may not exist
	// yet (M3.1); a spike should not make the operator mkdir by hand.
	if dir := filepath.Dir(outPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(outPath, []byte(markdown), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Fprintf(out, "wrote %s (%d bytes) for %q\n", outPath, len(markdown), title)
	return nil
}

// pipeline is the conversion step run() invokes, indirected through a variable
// so run()'s own test does not need the network.
var pipeline = articleMarkdown

// articleMarkdown is the spike's pipeline: fetch, strip, convert. M1.2 has the
// fetch and M2.1 the conversion; stripping non-article chrome before
// conversion is M2.2/M2.3.
func articleMarkdown(title string) (string, error) {
	html, err := fetchArticleHTML(restBase, title)
	if err != nil {
		return "", err
	}
	return htmlToMarkdown(html)
}
