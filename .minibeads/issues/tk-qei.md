---
title: Classify Markdown, Text and binary files; refuse binary and oversize opens
status: open
priority: 2
issue_type: feature
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
created_at: 2026-10-14T18:19:25.169292474+00:00
updated_at: 2026-10-14T18:19:25.169292474+00:00
---

# Description

Replace the Readable policy (`internal/tree/policy.go` `IsReadable`) with Markdown/Text/binary classification (D140).

- `IsMarkdown(name)`: `.md`/`.markdown`/`.mdown`/`.mkd` or an extensionless `README`/`CHANGELOG`, case-insensitive. `.txt` and `LICENSE` are no longer Markdown.
- Binary sniff: a NUL byte in the first ~8KB, cached by (path, mtime, size).
- Refuse binary files and files over ~10MB with a Footer message (`cannot open <name>: binary file` / `… too large`) on every open route: Files Nav, Picker, `noggin <file>`, Recent/Starred, history, Refresh.
- Until the File Filter lands, the Tree and Pickers keep showing Markdown files only, so behaviour matches today's `Markdown` filter.

**Acceptance**
- Tests cover the classification, the sniff cache and each refused route.
- README "What gets shown" is updated.
