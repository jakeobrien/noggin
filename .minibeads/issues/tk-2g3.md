---
title: Exact find and Grep landing in Text Render Mode
status: open
priority: 2
issue_type: feature
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
  tk-eru: blocks
created_at: 2026-10-14T18:19:25.176169932+00:00
updated_at: 2026-10-14T18:19:25.199025640+00:00
---

# Description

In Text mode (D140):

- Find matches against the source lines, smart-case as today, so a match can span soft-wrapped rows. Only the matched span is highlighted, not the whole line. With no ANSI in Text mode, this avoids the `SetHighlights` drift (`.charm/knowledge/viewport-search-ansi-offsets.md`). `n`/`N` and wrap-around behave as today.
- Activating a Grep Result into a Text-mode viewer lands on ripgrep's exact line and column, with that match as the current match of a Committed Find.
- Markdown mode find and D34's landing stay unchanged.

**Acceptance**
- Tests cover a match spanning a soft wrap, span highlighting, a match inside a link URL or `**bold**` (found in Text), and exact Grep landing.
