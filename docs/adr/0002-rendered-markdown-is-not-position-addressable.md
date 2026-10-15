# Rendered markdown is not mapped back to source positions

Glamour reflows, wraps and drops syntax, and highlight offsets drift once content carries ANSI styling. So a position in the rendered output can't be reliably tied to a position in the source. We chose not to build a source↔rendered position map. That has several consequences. The Document Viewer has no cursor. Select/copy, wiki-links and URL following are deferred as one group, so that a cursor, if it is ever bought, pays for all of them at once. Find in Markdown Render Mode matches the rendered text and marks the whole line. A Grep Result lands in Markdown mode by occurrence ordinal: the Nth match in the file is taken to be the Nth match in the render, which degrades to the first match. Exact positions come from Text Render Mode, which shows raw source and needs no map.

## Considered Options

- **Build the position map**: the only route to exact find and a cursor over formatted text. Rejected so far because it fights the libraries rather than using them.
- **Make the viewer raw-only**: rejected because finding while seeing the formatting still matters.

_Formerly D4, D34, D62, D140 (find/landing parts)._
