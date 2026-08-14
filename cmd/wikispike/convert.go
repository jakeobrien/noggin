package main

import (
	"fmt"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

// htmlToMarkdown converts Parsoid article HTML to markdown. Stripping of
// non-article chrome (navboxes, infoboxes, refs, images, links) is M2.2/M2.3;
// this is bare conversion only, table plugin included so wikitable content
// tables survive (M2.1's acceptance criterion).
func htmlToMarkdown(html string) (string, error) {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
	markdown, err := conv.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("converting to markdown: %w", err)
	}
	return markdown, nil
}
