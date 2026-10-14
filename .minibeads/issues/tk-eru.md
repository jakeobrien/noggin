---
title: Text Render Mode in the Document Viewer (m toggle)
status: closed
priority: 2
issue_type: feature
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
  tk-qei: blocks
created_at: 2026-10-14T18:19:25.173146599+00:00
updated_at: 2026-10-14T18:40:39.991597165+00:00
closed_at: 2026-10-14T18:40:39.991597082+00:00
---

# Description

Add a per-viewer Render Mode (D140).

- Add a field on `viewer`. It is sticky across open, Peek and history `[`/`]`, session-only, and defaults to `Markdown`. A new DV2 starts on `Markdown`.
- `m` toggles the mode from any non-modal pane. It targets DV1, or DV2 when DV2 has focus (the same rule as `/` and `e`).
- Text rendering:
  - raw source, frontmatter raw;
  - tabs expanded;
  - soft-wrapped to the pane width;
  - a dim line-number gutter, blank on continuation rows;
  - no colour;
  - a muted variant for unfocused panes (D136).
- A Text file always renders as Text without changing the viewer's mode. `m` on a Text file is refused with a Footer message.
- A Wikipedia article in Text shows its converted markdown.
- The Location line ends `· Markdown` / `· Text`.
- A toggle resets the find state for now. Preserving it is a separate ticket.

**Acceptance**
- Tests cover the mode toggle, stickiness across documents and history, Text files, articles and the Location line.
- README and the Help modal are updated.
