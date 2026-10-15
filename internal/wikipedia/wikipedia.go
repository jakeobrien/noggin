// Package wikipedia converts Wikimedia's Parsoid article HTML into markdown
// at the quality the C4a spike (cmd/wikispike) certified, and — from M2.2/M2.3
// onward — provides the title-search and canonical-fetch capability the
// Wikipedia Source (C6) is built on. cmd/noggin imports this package;
// cmd/wikispike stays as a thin CLI over it for eyeballing conversion quality
// across many articles.
package wikipedia

// ArticleMarkdownFromHTML runs the conversion pipeline (render maths, render
// images, strip chrome, delink, convert, clean up) on already-fetched Parsoid
// article HTML and returns the resulting markdown. Callers are responsible
// for fetching the HTML (cmd/wikispike's exact-title fetch today; M2.3's
// canonical fetch for cmd/noggin).
func ArticleMarkdownFromHTML(html string) (string, error) {
	mathed, err := renderMath(html)
	if err != nil {
		return "", err
	}
	imaged, err := renderImages(mathed)
	if err != nil {
		return "", err
	}
	stripped, err := stripChrome(imaged)
	if err != nil {
		return "", err
	}
	delinked, err := stripLinks(stripped)
	if err != nil {
		return "", err
	}
	markdown, err := htmlToMarkdown(delinked)
	if err != nil {
		return "", err
	}
	return cleanupMarkdown(markdown), nil
}
