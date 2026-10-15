---
title: Code blocks probably keep the launch Theme's colours after ctrl+t
status: open
priority: 3
issue_type: bug
labels:
- needs-triage
created_at: 2026-10-15T20:00:56.907937724+00:00
updated_at: 2026-10-15T20:00:56.907937724+00:00
---

# Description

## Suspected bug (unverified)

Glamour v2.0.1 registers its code-block Chroma style once per process, under a fixed name, and never replaces it (`charm.land/glamour/v2@v2.0.1/ansi/codeblock.go:85-90`: "Don't register the style if it's already registered"). noggin passes a per-Theme `ansi.Chroma` (`chromaStyle(t)`, `cmd/noggin/theme_markdown.go`) and re-renders open viewers on a Theme switch (`viewer.restyle`), but has no workaround for the registry. So after `ctrl+t`, fenced code blocks most likely keep the colours of the Theme the process launched with, while the rest of the document switches.

A known relative: the muted (unfocused) render already keeps focused code colours for the same reason. That one was accepted when the Frontmatter Card fallback shipped.

## To confirm

Open a document with a fenced code block in a real terminal (or a `make check-pty` check that asserts code-token SGR bytes before and after `ctrl+t`), cycle the Theme, and compare code-token colours to the new Theme's `SyntaxTheme`. `TestThemeKeyRepaintsTheWholeFrame` doesn't catch it, because `go test` renders without colour.

## Fix direction

Give each Theme (and its muted variant) a distinct Chroma style name. `ansi.StyleConfig` doesn't expose the name, so this means bypassing or patching Glamour's code-block element. Alternatively, clear or overwrite the `styles.Registry` entry before re-rendering (process-global and racy across viewers, so check carefully).

See `docs/dev/glamour-gotchas.md`.
