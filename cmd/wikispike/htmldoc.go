package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// typeofFilePrefix and typeofMathExtension are the two Parsoid `typeof`
// values this pipeline matches on repeatedly — stripChrome's catch-all
// image/math removal (strip.go), renderImages' image-placeholder pass
// (images.go) and renderMath's math-placeholder pass (math.go) each used to
// spell these out as their own string literal; named once here instead
// (TRIAGE T6, M3.6).
const (
	typeofFilePrefix    = "mw:File"
	typeofMathExtension = "mw:Extension/math"
)

// selectorFilePrefix and selectorMathExtension are the goquery CSS selectors
// built from the constants above — "starts with" for the file typeof
// (Parsoid suffixes it, e.g. "mw:File/Thumb"), exact match for the math
// extension typeof (Parsoid never suffixes it).
var (
	selectorFilePrefix    = fmt.Sprintf(`[typeof^=%q]`, typeofFilePrefix)
	selectorMathExtension = fmt.Sprintf(`[typeof=%q]`, typeofMathExtension)
)

// transformHTML parses html into a goquery document, lets mutate rewrite it
// in place, and re-serializes the result — the parse/mutate/re-serialize
// shape stripChrome, stripLinks, renderMath and renderImages each repeated
// as their own boilerplate (TRIAGE T6, M3.6). stage names the calling
// pipeline step (e.g. "strip chrome") for error context only.
func transformHTML(html, stage string, mutate func(*goquery.Document)) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("parsing HTML to %s: %w", stage, err)
	}
	mutate(doc)
	out, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("re-serializing HTML after %s: %w", stage, err)
	}
	return out, nil
}
