package main

import (
	"strings"
	"testing"
)

// figureHTML builds a minimal Parsoid-shaped figure, matching the real
// shape seen from a live fetch (an <a><img/></a> wrapped in <figure>, an
// optional <figcaption>, and the img's resource/alt attributes).
func figureHTML(resource, alt, caption string) string {
	var figcaption string
	if caption != "" {
		figcaption = "<figcaption>" + caption + "</figcaption>"
	}
	altAttr := ""
	if alt != "" {
		altAttr = ` alt="` + alt + `"`
	}
	return `<figure typeof="mw:File/Thumb"><a href="./` + resource + `"><img resource="./` + resource + `"` + altAttr + `/></a>` + figcaption + `</figure>`
}

// TestRenderImagesPrefersCaption covers SPEC S3a's top priority: a figure
// with a caption uses the caption text, even when alt-text is also present.
func TestRenderImagesPrefersCaption(t *testing.T) {
	input := `<p>` + figureHTML("File:Example.png", "an alt description", "The caption text") + `</p>`

	got, err := renderImages(input)
	if err != nil {
		t.Fatalf("renderImages: %v", err)
	}
	if !strings.Contains(got, "[Image: The caption text]") {
		t.Errorf("expected the caption-derived placeholder, got:\n%s", got)
	}
	if strings.Contains(got, "an alt description") {
		t.Errorf("did not expect alt-text to surface when a caption exists:\n%s", got)
	}
}

// TestRenderImagesFallsBackToAltThenFilename covers SPEC S3a's fallback
// chain: no caption falls back to alt-text; no caption and no alt-text
// falls back to a filename derived from the file reference.
func TestRenderImagesFallsBackToAltThenFilename(t *testing.T) {
	t.Run("alt", func(t *testing.T) {
		input := figureHTML("File:Example.png", "an alt description", "")
		got, err := renderImages(input)
		if err != nil {
			t.Fatalf("renderImages: %v", err)
		}
		if !strings.Contains(got, "[Image: an alt description]") {
			t.Errorf("expected the alt-derived placeholder, got:\n%s", got)
		}
	})

	t.Run("filename", func(t *testing.T) {
		input := figureHTML("File:Sanzio_01_Parmenides.jpg", "", "")
		got, err := renderImages(input)
		if err != nil {
			t.Fatalf("renderImages: %v", err)
		}
		if !strings.Contains(got, "[Image: Sanzio 01 Parmenides]") {
			t.Errorf("expected the filename-derived placeholder (underscores as spaces, extension stripped), got:\n%s", got)
		}
	})
}

// TestRenderImagesFlattensCaptionMarkup covers a caption carrying nested
// markup (a wikilink, as real captions do, e.g. "Rider\u2013Waite Tarot"): the
// placeholder keeps only the link's text, not markup or a URL, matching
// SPEC S3a's "no clickable link" requirement.
func TestRenderImagesFlattensCaptionMarkup(t *testing.T) {
	input := `<figure typeof="mw:File/Thumb"><a href="./File:Example.png"><img resource="./File:Example.png"/></a><figcaption>Deck from the <a href="./Rider-Waite_Tarot">Rider-Waite Tarot</a></figcaption></figure>`

	got, err := renderImages(input)
	if err != nil {
		t.Fatalf("renderImages: %v", err)
	}
	if !strings.Contains(got, "[Image: Deck from the Rider-Waite Tarot]") {
		t.Errorf("expected a flattened, linkless placeholder, got:\n%s", got)
	}
	if strings.Contains(got, "href") || strings.Contains(got, "<a") {
		t.Errorf("did not expect any link markup to survive, got:\n%s", got)
	}
}

// TestRenderImagesLeavesUnrecoverableNodeForStripChrome covers the edge
// case where an image node has no caption, no alt-text and no recoverable
// filename at all: renderImages makes no substitution, so it is left for
// stripChrome's own image selector to remove, matching pre-M3.5 behaviour.
func TestRenderImagesLeavesUnrecoverableNodeForStripChrome(t *testing.T) {
	input := `<span typeof="mw:File"><img/></span>`

	got, err := renderImages(input)
	if err != nil {
		t.Fatalf("renderImages: %v", err)
	}
	if !strings.Contains(got, `typeof="mw:File"`) {
		t.Errorf("expected the untouched node to survive renderImages for stripChrome to remove, got:\n%s", got)
	}
}

// TestArticleMarkdownFromHTMLKeepsImagePlaceholder is the full-pipeline
// check: an image that used to vanish without a trace (pre-D104) now
// survives as a backtick-wrapped, bracketed placeholder next to its
// original position, with no dangling brackets or link markup, through
// fetch's real successor stages (render maths, render images, strip,
// delink, convert, clean up).
func TestArticleMarkdownFromHTMLKeepsImagePlaceholder(t *testing.T) {
	input := `<html><body><section>
<p>Intro prose.</p>
` + figureHTML("File:Example.png", "", "A caption about the figure") + `
<p>More prose follows.</p>
</section></body></html>`

	got, err := articleMarkdownFromHTML(input)
	if err != nil {
		t.Fatalf("articleMarkdownFromHTML: %v", err)
	}
	if !strings.Contains(got, "`[Image: A caption about the figure]`") {
		t.Errorf("expected the caption placeholder wrapped in backticks in the final markdown, got:\n%s", got)
	}
	if !strings.Contains(got, "Intro prose.") || !strings.Contains(got, "More prose follows.") {
		t.Errorf("expected the surrounding prose to survive intact, got:\n%s", got)
	}
	if strings.Contains(got, "![") {
		t.Errorf("did not expect markdown image syntax to survive, got:\n%s", got)
	}
	if strings.Contains(got, "[]") {
		t.Errorf("did not expect a dangling empty-bracket gap, got:\n%s", got)
	}
}

// TestArticleMarkdownLiveSamplesShowImagePlaceholders is M3.5's acceptance
// check against the four S5 sample titles (shared live-fetch cache, M2.5):
// every sample that carries at least one genuine content image (a
// <figure typeof="mw:File/Thumb">, distinct from a chrome icon such as an
// ambox's <span typeof="mw:File"> maintenance-box image, which stripChrome
// removes wholesale regardless of any placeholder) now shows a
// backtick-wrapped, bracketed placeholder in the final markdown, not
// markdown image syntax, and no dangling empty-bracket gap is left behind.
// Opt-in like the other live tests so `make check` stays hermetic:
// WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestArticleMarkdownLiveSamplesShowImagePlaceholders(t *testing.T) {
	skipUnlessLive(t)
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			raw := fetchLiveSample(t, title)
			if !strings.Contains(raw, "<figure") {
				t.Skipf("%q: raw fetch carries no content <figure> image, nothing to place a placeholder for", title)
			}
			markdown := liveMarkdown(t, title)
			if !strings.Contains(markdown, "`[Image: ") {
				t.Errorf("%q: expected at least one backtick-wrapped [Image: ...] placeholder, got none", title)
			}
			if strings.Contains(markdown, "![") {
				t.Errorf("%q: markdown image syntax survived instead of a text placeholder", title)
			}
			if strings.Contains(markdown, "[]") {
				t.Errorf("%q: a dangling empty-bracket gap survived", title)
			}
		})
	}
}
