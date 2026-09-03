package main

import (
	stdhtml "html"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// renderImages replaces each Wikipedia image (any element whose typeof
// starts with "mw:File" — figures, thumbnails, inline file references) with
// a plain-text placeholder, so the image survives the pipeline as a short
// textual hint instead of vanishing without a trace when stripChrome would
// otherwise remove it — the S6a image exception (D104, M3.5).
//
// SPEC S3a's priority order, applied per image: the figure's caption,
// falling back to the img's alt-text, falling back to a filename derived
// from the file reference. All three are already present on Parsoid's HTML
// for free — no OCR, no extra fetch — so this is the same near-zero-cost
// shape M3.4's renderMath established: compute a plain-text replacement per
// matched node and leave every other pipeline stage untouched. No image is
// fetched, decoded or rendered; the placeholder is wrapped in square
// brackets so it reads as a placeholder rather than body prose, and carries
// no link (stripLinks, M2.3, would have stripped any link anyway — this
// just never gives it one). It is also wrapped in a <code> element so it
// renders as inline code (backtick-delimited) rather than plain prose —
// visually distinct from surrounding text, same reasoning as the square
// brackets, one layer further. Using an actual <code> element rather than
// literal backtick characters lets the html-to-markdown converter handle
// escaping if a caption/alt/filename itself contains a backtick, instead of
// this code having to reimplement that.
func renderImages(html string) (string, error) {
	return transformHTML(html, "render images", func(doc *goquery.Document) {
		doc.Find(selectorFilePrefix).Each(func(_ int, s *goquery.Selection) {
			// A citation superscript (e.g. the "[1][2]" after a caption's
			// closing parenthesis) can sit inside a figcaption; stripChrome's
			// own removal of these (strip.go) runs after renderImages, so
			// without this the ref-marker digits leak straight into the
			// placeholder text below (M3.5a, TRIAGE T11). Removed from the
			// live node, not a clone: this image node is either about to be
			// replaced wholesale with the placeholder, or left for
			// stripChrome to remove outright — either way the mutation is
			// harmless beyond this pass.
			s.Find(selectorCitationSup).Remove()
			desc := imagePlaceholderText(s)
			if desc == "" {
				// No usable description at all (vanishingly rare — even a
				// bare filename should be recoverable): leave the node in
				// place for stripChrome's own image selector to remove,
				// same as before this experiment existed.
				return
			}
			s.ReplaceWithHtml("<code>" + stdhtml.EscapeString("[Image: "+desc+"]") + "</code>")
		})
	})
}

// imagePlaceholderText picks an image's best available description, in
// SPEC S3a's priority order: caption, then alt-text, then filename. Returns
// "" only if all three are unavailable.
func imagePlaceholderText(s *goquery.Selection) string {
	// elementText (strip.go, M3.3b) excludes inlined <style>/<script>
	// descendants before reading text — the same templatestyles-CSS-leak
	// failure class M3.3b fixed for sisterProjectInlineLine can just as
	// easily land inside a figcaption (TRIAGE T6), so a raw .Text() call
	// here would risk the same CSS-prepended-to-wording corruption.
	if caption := strings.TrimSpace(elementText(s.Find("figcaption").First())); caption != "" {
		return caption
	}

	img := s
	if !s.Is("img") {
		img = s.Find("img").First()
	}

	if alt, ok := img.Attr("alt"); ok {
		if alt = strings.TrimSpace(alt); alt != "" {
			return alt
		}
	}

	return imageFilename(s, img)
}

// imageFilename recovers a human-readable filename from whichever attribute
// carries it: the img's resource attribute first (present on every Parsoid
// file image), falling back to the wrapping link's href. Both are shaped
// "./File:Some_name.ext" or "File:Some_name.ext"; this strips the "File:"
// prefix, the extension, and turns underscores back into spaces (the
// reverse of Wikipedia's title-to-URL convention, cmd/wikispike/fetch.go's
// articleURL).
func imageFilename(s *goquery.Selection, img *goquery.Selection) string {
	raw, ok := img.Attr("resource")
	if !ok || raw == "" {
		raw, ok = s.Find("a[href]").First().Attr("href")
		if !ok || raw == "" {
			return ""
		}
	}
	name := strings.TrimPrefix(raw, "./")
	name = strings.TrimPrefix(name, "File:")
	name = strings.TrimSuffix(name, path.Ext(name))
	name = strings.ReplaceAll(name, "_", " ")
	return strings.TrimSpace(name)
}
