---
title: 'Spec: Markdown/Text Render Mode and File Filter'
status: open
priority: 2
issue_type: epic
labels:
- spec
- needs-triage
created_at: 2026-10-14T18:19:25.161522807+00:00
updated_at: 2026-10-14T18:19:25.161522807+00:00
---

# Description

Widen noggin from markdown-only to any text file, and add a raw Text view that gives exact Find/Grep highlighting, while keeping rendered-markdown Find available.

Full decision and rationale: `.charm/DECISIONS.md` **D140**. Vocabulary: `.charm/GLOSSARY.md` (File Filter, Render Mode, Markdown file, Text file; Readable file is retired).

## Summary

- **Markdown file**: `.md`/`.markdown`/`.mdown`/`.mkd`, or an extensionless `README`/`CHANGELOG` (case-insensitive). **Text file**: any other non-binary file. **Binary**: a NUL byte in the first ~8KB.
- **File Filter** (`Markdown` | `All`):
  - Per Root, session-only, default `Markdown`.
  - Toggled by `a` (Files Nav) and `alt+a` (Picker).
  - Bounds the Tree and the File/Grep Active Root and All Roots scopes. All Roots uses the Active Root's filter.
  - Recent and Starred ignore it.
- **Render Mode** (`Markdown` | `Text`):
  - Per viewer, sticky, session-only, default `Markdown`. `m` toggles it.
  - A Text file always displays as text, and `m` on it is refused.
  - The Location line ends `· Markdown` / `· Text`.
  - The Picker has no mode control.
- **Text rendering**: raw source, soft wrap, dim line-number gutter, no colour. The Picker Preview uses it too.
- **Find in Text**: exact span highlighting against source lines. Grep lands on ripgrep's exact line and column. Find in Markdown is unchanged.
- **Toggling mid-document**: the find query and match ordinal are kept, falling back to ripgrep's exact location. Without a find, scroll is kept proportionally.
- **Safety**: binaries and files over ~10MB are refused on every open route.
